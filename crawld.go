// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	_ "github.com/lib/pq"

	"github.com/DevMine/crawld/config"
	"github.com/DevMine/crawld/crawlers"
	"github.com/DevMine/crawld/errbag"
	"github.com/DevMine/crawld/repo"
	"github.com/DevMine/crawld/tar"
)

func crawlingWorker(cs []crawlers.Crawler, crawlingInterval time.Duration) {
	for {
		var wg sync.WaitGroup

		wg.Add(len(cs))
		for _, c := range cs {
			glog.Infof("starting a goroutine for the %v crawler\n", reflect.TypeOf(c))
			go func(c crawlers.Crawler) {
				defer wg.Done()
				c.Crawl()
			}(c)
		}

		wg.Wait()

		glog.Infof("waiting for %v before re-starting the crawlers.\n", crawlingInterval)
		<-time.After(crawlingInterval)
	}
}

func repoWorker(db *sql.DB, langs []string, basePath string, fetchInterval time.Duration, useTar bool, maxWorkers uint, errBag *errbag.ErrBag) {
	clone := func(r repo.Repo) {
		glog.Infof("cloning %s into %s\n", r.URL(), r.AbsPath())
		if err := r.Clone(); err != nil {
			glog.Errorf("impossible to clone %s in %s ("+err.Error()+") skipping", r.URL(), r.AbsPath())
			errBag.Record(err)
		}
	}

	update := func(r repo.Repo) {
		glog.Infof("updating %s\n", r.AbsPath())
		if err := r.Update(); err != nil {
			glog.Warningf("impossible to update %s ("+err.Error()+")", r.AbsPath())
			errBag.Record(err)

			// we just want to skip on a network error
			if err == repo.ErrNetwork {
				return
			}

			// delete and reclone then
			glog.Infof("attempting to re-clone %s", r.AbsPath())
			if err2 := os.RemoveAll(r.AbsPath()); err2 != nil {
				glog.Errorf("cannot remove %s("+err2.Error()+")", r.AbsPath())
				errBag.Record(err)
			} else {
				clone(r)
			}
		}
	}

	createArchive := func(path string) {
		if err := tar.CreateInPlace(path); err != nil {
			glog.Error("impossible to create tar archive (" + path + ".tar ): " +
				err.Error())
			errBag.Record(err)
		}
	}

	for {
		glog.Info("starting the repositories fetcher")
		repos, err := getAllRepos(db, langs, basePath)
		if err != nil {
			fatal(err)
		}

		tasks := make(chan repo.Repo, len(repos))
		var wg sync.WaitGroup

		for _, r := range repos {
			tasks <- r
		}

		for w := uint(0); w < maxWorkers; w++ {
			wg.Add(1)
			go func() {
				for r := range tasks {
					// check if we have a tar archive of the repository in which case
					// we only need to update but extract apriori and recreate the tar
					// archive afterwards
					archive := r.AbsPath() + ".tar"
					if _, err = os.Stat(archive); err == nil {
						if err = tar.ExtractInPlace(archive); err != nil {
							glog.Warning("impossible to extract the tar archive (" + archive + ")" +
								", cannot update the repository: " + err.Error())
							// attempt to remove the eventual mess
							_ = os.Remove(archive)
							_ = os.RemoveAll(r.AbsPath())
							clone(r)
						} else {
							update(r)
						}
					} else {
						if _, err := os.Stat(r.AbsPath()); os.IsNotExist(err) || isDirEmpty(r.AbsPath()) {
							clone(r)
						} else {
							update(r)
						}
					}

					if useTar {
						createArchive(r.AbsPath())
					}

					if err = r.Cleanup(); err != nil {
						glog.Warning(err)
					}
				}
				wg.Done()
			}()
		}

		close(tasks)
		wg.Wait()

		glog.Infof("waiting for %v before re-starting the fetcher.\n", fetchInterval)
		<-time.After(fetchInterval)
	}
}

func isDirEmpty(path string) bool {
	fis, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	}

	return len(fis) == 0
}

func getAllRepos(db *sql.DB, langs []string, basePath string) ([]repo.Repo, error) {
	var inClause string
	if langs != nil && len(langs) > 0 {
		// Quote languages.
		for idx, val := range langs {
			langs[idx] = "'" + val + "'"
		}
		inClause = " WHERE LOWER(primary_language) IN (" + strings.Join(langs, ",") + ")"
	}

	rows, err := db.Query("SELECT vcs, clone_path, clone_url FROM repositories" + inClause)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	defer rows.Close()

	var repos []repo.Repo

	for rows.Next() {
		var vcs, clonePath, cloneURL string
		if err := rows.Scan(&vcs, &clonePath, &cloneURL); err != nil {
			glog.Error(err)
			continue
		}

		var newRepo repo.Repo
		var err error

		newRepo, err = repo.New(vcs, filepath.Join(basePath, clonePath), cloneURL)
		if err != nil {
			glog.Error(err)
			continue
		}

		repos = append(repos, newRepo)
	}

	return repos, nil
}

func checkCloneDir(cloneDir string) error {
	// check if clone path exists
	if fi, err := os.Stat(cloneDir); err == nil {
		glog.Error(err)
		return err
	} else if !fi.IsDir() {
		err = errors.New("clone path must be a directory")
		glog.Error(err)
		return err
	}

	// check if clone path is writable
	// note: since the directory already exists, then the file perm param is
	// useless
	file, err := os.OpenFile(cloneDir, os.O_RDWR, 0770)
	if err != nil {
		err = errors.New("clone path must be writable")
		glog.Error(err)
		return err
	}
	file.Close()

	return nil
}

func openDBSession(cfg config.DatabaseConfig) (*sql.DB, error) {
	dbURL := fmt.Sprintf(
		"user='%s' password='%s' host='%s' port=%d dbname='%s' sslmode='%s'",
		cfg.UserName, cfg.Password, cfg.HostName, cfg.Port, cfg.DBName, cfg.SSLMode)

	return sql.Open("postgres", dbURL)
}

func fatal(a ...interface{}) {
	glog.Error(a)
	os.Exit(1)
}

func main() {
	configPath := flag.String("c", "", "configuration file")
	disableCrawlers := flag.Bool("disable-crawlers", false, "disable the data crawlers")
	disableFetcher := flag.Bool("disable-fetcher", false, "disable the repositories fetcher")
	flag.Parse()

	// Make sure we finish writing logs before exiting.
	defer glog.Flush()

	if len(*configPath) == 0 {
		fatal("no configuration specified")
	}

	cfg, err := config.ReadConfig(*configPath)
	if err != nil {
		fatal(err)
	}

	db, err := openDBSession(cfg.Database)
	if err != nil {
		fatal(err)
	}
	defer db.Close()

	var cs []crawlers.Crawler

	for _, crawlerConfig := range cfg.Crawlers {
		c, err := crawlers.New(crawlerConfig, db)
		if err != nil {
			fatal(err)
		}

		cs = append(cs, c)
	}

	crawlingInterval, err := time.ParseDuration(cfg.CrawlingTimeInterval)
	if err != nil {
		fatal(err)
	}

	var wg sync.WaitGroup

	// start the crawling worker
	if !*disableCrawlers {
		wg.Add(1)
		go crawlingWorker(cs, crawlingInterval)
	}

	fetchInterval, err := time.ParseDuration(cfg.FetchTimeInterval)
	if err != nil {
		fatal(err)
	}

	// start the repo puller worker
	if !*disableFetcher {
		errBag, err := errbag.New(cfg.ThrottlerWaitTime, cfg.SlidingWindowSize, cfg.LeakInterval)
		if err != nil {
			glog.Error("impossible to start the repositories fetcher")
			return
		}
		errBag.Inflate()
		defer errBag.Deflate()

		wg.Add(1)
		go repoWorker(db, cfg.FetchLanguages, cfg.CloneDir, fetchInterval, cfg.TarRepos, cfg.MaxFetcherWorkers, errBag)
	}

	// wait until the cows come home saint
	wg.Wait()
}

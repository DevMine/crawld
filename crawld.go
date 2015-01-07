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
	"sync"
	"time"

	"github.com/golang/glog"
	_ "github.com/lib/pq"

	"github.com/DevMine/crawld/config"
	"github.com/DevMine/crawld/crawlers"
	"github.com/DevMine/crawld/repo"
	"github.com/DevMine/crawld/repo/git"
)

func crawlingWorker(cs []crawlers.Crawler, crawlingInterval time.Duration) {
	for {
		var wg sync.WaitGroup

		wg.Add(len(cs))
		for _, c := range cs {
			glog.Infof("crawlingWorker: starting a goroutine for the %v crawler\n", reflect.TypeOf(c))
			go func(c crawlers.Crawler) {
				defer wg.Done()
				c.Crawl()
			}(c)
		}

		wg.Wait()

		glog.Infof("Crawling worker: waiting for %v before re-starting.\n", crawlingInterval)
		<-time.After(crawlingInterval)
	}
}

func repoWorker(db *sql.DB, basePath string, fetchInterval time.Duration) {
	for {
		glog.Info("repoWorker: starting the repositories fetcher")

		repos, err := getAllRepos(db, basePath)
		if err != nil {
			fatal(err)
		}

		for _, r := range repos {
			if _, err := os.Stat(r.AbsPath()); os.IsNotExist(err) || isDirEmpty(r.AbsPath()) {
				glog.Infof("repoWorker: cloning %s into %s\n", r.URL(), r.AbsPath())
				_ = r.Clone()
				continue
			}
			glog.Infof("repoWorker: updating %s\n", r.AbsPath())
			_ = r.Update()
		}

		glog.Infof("Fetching worker: waiting for %v before re-starting.\n", fetchInterval)
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

func getAllRepos(db *sql.DB, basePath string) ([]repo.Repo, error) {
	rows, err := db.Query("SELECT vcs, clone_path, clone_url FROM repositories")
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

		switch vcs {
		case "git":
			newRepo, err = git.New(filepath.Join(basePath, clonePath), cloneURL)
			if err != nil {
				glog.Error(err)
				continue
			}
		default:
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
	disableCrawlers := flag.Bool("disable-crawlers", false, "disable the crawlers")
	disableFetchers := flag.Bool("disable-fetchers", false, "disable the fetchers")
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
		switch crawlerConfig.Type {
		case "github":
			glog.Info("main: github crawler selected")
			gh, err := crawlers.NewGitHubCrawler(crawlerConfig, cfg.CloneDir, db)
			if err != nil {
				fatal(err)
			}

			cs = append(cs, gh)
		}
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
	if !*disableFetchers {
		wg.Add(1)
		go repoWorker(db, cfg.CloneDir, fetchInterval)
	}

	// wait until the cows come home saint
	wg.Wait()
}

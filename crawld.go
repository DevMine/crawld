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
	"os/signal"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
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

// extend this structure later if required but for now the repository id sufficient
type dbRepo struct {
	repo.Repo
	id uint64
}

// channel used to communicate repositories IDs
var idChan chan uint64

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

func repoWorker(db *sql.DB, cfg *config.Config, startID uint64, errBag *errbag.ErrBag) {
	fetchInterval, err := time.ParseDuration(cfg.FetchTimeInterval)
	if err != nil {
		fatal(err)
	}

	callback := func(status errbag.Status) {
		if status.State == errbag.StatusThrottling {
			glog.Info("too many errors received; waiting for ", status.WaitTime, " seconds before resuming")
		}
	}

	clone := func(r repo.Repo) error {
		glog.Infof("cloning %s into %s\n", r.URL(), r.AbsPath())
		if err := r.Clone(); err != nil {
			glog.Errorf("impossible to clone %s in %s ("+err.Error()+") skipping", r.URL(), r.AbsPath())
			errBag.Record(err, callback)
			return err
		}
		return nil
	}

	update := func(r repo.Repo) error {
		glog.Infof("updating %s\n", r.AbsPath())
		if err := r.Update(); err != nil {
			glog.Warningf("impossible to update %s ("+err.Error()+")", r.AbsPath())
			errBag.Record(err, callback)

			// we just want to skip on a network error
			if err == repo.ErrNetwork {
				return err
			}

			// delete and reclone then
			glog.Infof("attempting to re-clone %s", r.AbsPath())
			if err2 := os.RemoveAll(r.AbsPath()); err2 != nil {
				glog.Errorf("cannot remove %s("+err2.Error()+")", r.AbsPath())
				errBag.Record(err, callback)
				return err
			}
			return clone(r)
		}
		return nil
	}

	for {
		glog.Info("starting the repositories fetcher")
		repos, err := getAllRepos(db, startID, cfg.FetchLanguages, cfg.CloneDir)
		if err != nil {
			fatal(err)
		}
		// next time, we want to get all repos from the first one
		startID = 0

		tasks := make(chan dbRepo, len(repos))
		var wg sync.WaitGroup

		for _, r := range repos {
			tasks <- r
		}
		// we don't want any routine to add new tasks in the queue now
		// if we don't close the channel now, the goroutines processing the
		// tasks will wait forever for new tasks and never return
		close(tasks)

		for w := uint(0); w < cfg.MaxFetcherWorkers; w++ {
			wg.Add(1)
			go func() {
				for r := range tasks {
					err := func() error {
						defer func() {
							if err = r.Cleanup(); err != nil {
								glog.Warning(err)
								errBag.Record(err, callback)
							}
						}()

						var tmpPath, tmpDest string
						var useTmpDir bool
						archive := r.AbsPath() + ".tar"

						if cfg.TarRepos {
							// we need to define the temp working directory then
							tmpPath, err = ioutil.TempDir(cfg.TmpDir, "repo-")
							if err != nil {
								glog.Error("cannot create temporary directory in " + cfg.TmpDir)
								errBag.Record(err, callback)
								return err
							}
							tmpDest = filepath.Join(tmpPath, filepath.Base(r.AbsPath()))
							useTmpDir = true

							defer func() {
								if err = os.RemoveAll(tmpPath); err != nil {
									glog.Warning("impossible to remove temporary directory: " + tmpPath)
									errBag.Record(err, callback)
								}
							}()
						}

						// if we have a tar archive, we need to extract it
						if fi, err := os.Stat(archive); err == nil {
							if useTmpDir && (bytesToGigaBytes(fi.Size()) < cfg.TmpDirFileSizeLimit) {
								if err = tar.Extract(filepath.Dir(tmpDest), archive); err != nil {
									glog.Warning("impossible to extract tar archive (" + archive + ")" +
										", cannot update repository: " + err.Error())
									// attempt to remove the eventual mess
									_ = os.Remove(archive)
									_ = os.RemoveAll(tmpDest)
								}
							} else {
								if err = tar.ExtractInPlace(archive); err != nil {
									glog.Warning("impossible to extract tar archive (" + archive + ")" +
										", cannot update repository: " + err.Error())
									// attempt to remove the eventual mess
									_ = os.Remove(archive)
									_ = os.RemoveAll(r.AbsPath())
								}
							}
						}

						if useTmpDir {
							path := r.AbsPath()
							r.SetAbsPath(tmpDest)
							if _, err := os.Stat(tmpDest); os.IsNotExist(err) || isDirEmpty(tmpDest) {
								useTmpDir = false
								// maybe we have it on main storage, not as a tar archive
								if _, err := os.Stat(path); os.IsNotExist(err) || isDirEmpty(path) {
									if err = clone(r); err != nil {
										return err
									}
								} else {
									r.SetAbsPath(path)
									if err = update(r); err != nil {
										return err
									}
								}
							} else {
								if err = update(r); err != nil {
									return err
								}
							}
							r.SetAbsPath(path)
						} else {
							if _, err := os.Stat(r.AbsPath()); os.IsNotExist(err) || isDirEmpty(r.AbsPath()) {
								if err = clone(r); err != nil {
									return err
								}
							} else {
								if err = update(r); err != nil {
									return err
								}
							}
						}

						if cfg.TarRepos {
							if useTmpDir {
								os.MkdirAll(filepath.Dir(r.AbsPath()), 0755)
								err = tar.Create(archive, tmpDest)
								// no need to remove tmpDest here since tmpPath is removed after processing
							} else {
								err = tar.CreateInPlace(r.AbsPath())
							}
							if err != nil {
								glog.Error("impossible to create tar archive ("+archive+"): ", err)
								errBag.Record(err, callback)
							}
						}
						return nil
					}()

					if err == nil {
						// notify we're done with this repository
						idChan <- r.id
					}
				}
				wg.Done()
			}()
		}

		wg.Wait()

		glog.Infof("waiting for %v before re-starting the fetcher.\n", fetchInterval)
		<-time.After(fetchInterval)
	}
}

func bytesToGigaBytes(bytes int64) float64 {
	return float64(bytes) / 1000000000.0
}

func isDirEmpty(path string) bool {
	fis, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	}

	return len(fis) == 0
}

func getAllRepos(db *sql.DB, startID uint64, langs []string, basePath string) ([]dbRepo, error) {
	inClause := fmt.Sprintf("WHERE id >= %d", startID)
	if langs != nil && len(langs) > 0 {
		// Quote languages.
		for idx, val := range langs {
			langs[idx] = "'" + val + "'"
		}
		inClause += " AND LOWER(primary_language) IN (" + strings.Join(langs, ",") + ")"
	}

	rows, err := db.Query("SELECT id, vcs, clone_path, clone_url FROM repositories " + inClause + " ORDER BY id")
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	defer rows.Close()

	var repos []dbRepo

	for rows.Next() {
		var vcs, clonePath, cloneURL string
		var id uint64
		if err := rows.Scan(&id, &vcs, &clonePath, &cloneURL); err != nil {
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

		repos = append(repos, dbRepo{Repo: newRepo, id: id})
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

	// start the repo puller worker
	if !*disableFetcher {
		errBag, err := errbag.New(cfg.ThrottlerWaitTime, cfg.SlidingWindowSize, cfg.LeakInterval)
		if err != nil {
			glog.Error("impossible to start the repositories fetcher")
			return
		}
		errBag.Inflate()

		var startID uint64
		lastFetchedIDFile := path.Join(cfg.CloneDir, "last_fetched_id")
		if bs, err := ioutil.ReadFile(lastFetchedIDFile); len(bs) != 0 && err == nil {
			if startID, err = strconv.ParseUint(string(bs), 10, 64); err != nil {
				glog.Warning("cannot convert (" + string(bs) + ") to a repository id, starting from 0...")
				startID = 0
			}
		} else {
			glog.Warning("cannot get last fetched repository id, starting from 0...")
			startID = 0
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		idChan = make(chan uint64)

		// this routines writes the last processed repository id in a file, getting it from idChan
		go func() {
			f, err := os.OpenFile(lastFetchedIDFile, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				glog.Fatal("cannot open file for writing (" + lastFetchedIDFile + "): " + err.Error())
			}

			// we want to make sure we close the file and do some housekeeping on interruption
			go func() {
				<-c
				fmt.Fprintln(os.Stderr, "caught signal, exiting now...")
				f.Sync()
				f.Close()
				errBag.Deflate()
				os.Exit(0)
			}()

			for id, ok := <-idChan; ok; id, ok = <-idChan {
				if _, err := f.Seek(0, 0); err != nil {
					glog.Warning("could not write ID to file:", id)
				} else {
					// pad with 0 up to 20 because the largest unsigned integer
					// of 64 bit fits in 20 digits in decimal format
					fmt.Fprintf(f, "%020d", id)
				}
			}
		}()

		wg.Add(1)
		go repoWorker(db, cfg, startID, errBag)
	}

	// wait until the cows come home saint
	wg.Wait()
}

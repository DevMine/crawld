// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crawlers provides crawlers for gathering source code repository
// information.
package crawlers

import (
	"database/sql"
	"errors"

	"github.com/DevMine/crawld/config"
)

// Crawler defines methods a crawler must implement.
type Crawler interface {
	// Crawl methods crawls data and put it into the database.
	Crawl()
}

// New creates a new crawler. cfg corresponds to the crawler configuration,
// db is an opened session to the database.
func New(cfg config.CrawlerConfig, db *sql.DB) (Crawler, error) {
	var newCrawler Crawler
	var err error

	switch cfg.Type {
	case "github":
		newCrawler, err = newGitHubCrawler(cfg, db)
	default:
		return nil, errors.New("unsupported crawler type: " + cfg.Type)
	}
	if err != nil {
		return nil, err
	}

	return newCrawler, nil
}

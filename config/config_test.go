// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config_test

import (
	"strings"
	"testing"

	"github.com/DevMine/crawld/config"
)

const (
	configPath = "../testdata/crawld.conf"

	expectedCloneDir             = "/var/crawld"
	expectedCrawlingTimeInterval = "12h"

	expectedCrawlersLen             = 1
	expectedCrawlerType             = "github"
	expectedCrawlerLanguages        = "go,ruby"
	expectedCrawlerLimit            = 0
	expectedCrawlerFork             = false
	expectedCrawlerOAuthAccessToken = "token here"
	expectedCrawlerUseSearchAPI     = false

	expectedDatabaseHostName = "localhost"
	expectedDatabasePort     = 5432
	expectedDatabaseUserName = "devmine"
	expectedDatabasePassword = "devmine"
	expectedDatabaseDBName   = "devmine"
)

func TestReadConfig(t *testing.T) {
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.CloneDir != expectedCloneDir {
		t.Errorf("clone_dir: expected '%s', found '%s'\n",
			expectedCloneDir, cfg.CloneDir)
	}

	if cfg.CrawlingTimeInterval != expectedCrawlingTimeInterval {
		t.Errorf("crawling_time_interval: expected '%s', found '%s'\n",
			expectedCrawlingTimeInterval, cfg.CrawlingTimeInterval)
	}

	if len(cfg.Crawlers) != expectedCrawlersLen {
		t.Errorf("len(crawlers): expected %d, found %d\n",
			expectedCrawlersLen, len(cfg.Crawlers))
	}

	if cfg.Crawlers[0].Type != expectedCrawlerType {
		t.Errorf("crawlers[0].type: expected '%s', found '%s'\n",
			expectedCrawlerType, cfg.Crawlers[0].Type)
	}

	if strings.Join(cfg.Crawlers[0].Languages, ",") != expectedCrawlerLanguages {
		t.Errorf("crawlers[0].languages: expected '%s', found '%s'\n",
			expectedCrawlerLanguages, strings.Join(cfg.Crawlers[0].Languages, ","))
	}

	if cfg.Crawlers[0].Limit != expectedCrawlerLimit {
		t.Errorf("crawlers[0].limit: expected %d, found %d\n",
			expectedCrawlerLimit, cfg.Crawlers[0].Limit)
	}

	if cfg.Crawlers[0].Fork != expectedCrawlerFork {
		t.Errorf("crawlers[0].fork: expected %t, found %t\n",
			expectedCrawlerFork, cfg.Crawlers[0].Fork)
	}

	if cfg.Crawlers[0].OAuthAccessToken != expectedCrawlerOAuthAccessToken {
		t.Errorf("crawlers[0].oAuth_access_token: expected '%s', found '%s'\n",
			expectedCrawlerOAuthAccessToken, cfg.Crawlers[0].OAuthAccessToken)
	}

	if cfg.Crawlers[0].UseSearchAPI != expectedCrawlerUseSearchAPI {
		t.Errorf("crawlers[0].use_search_api: expected %t, found %t\n",
			expectedCrawlerUseSearchAPI, cfg.Crawlers[0].UseSearchAPI)
	}

	if cfg.Database.HostName != expectedDatabaseHostName {
		t.Errorf("database.hostname: expected '%s', found '%s'\n",
			expectedDatabaseHostName, cfg.Database.HostName)
	}

	if cfg.Database.Port != expectedDatabasePort {
		t.Errorf("database.hostname: expected %d, found %d\n",
			expectedDatabasePort, cfg.Database.Port)
	}

	if cfg.Database.UserName != expectedDatabaseUserName {
		t.Errorf("database.username: expected '%s', found '%s'\n",
			expectedDatabaseUserName, cfg.Database.UserName)
	}

	if cfg.Database.Password != expectedDatabasePassword {
		t.Errorf("database.password: expected '%s', found '%s'\n",
			expectedDatabasePassword, cfg.Database.Password)
	}

	if cfg.Database.DBName != expectedDatabaseDBName {
		t.Errorf("database.dbname: expected '%s', found '%s'\n",
			expectedDatabaseDBName, cfg.Database.DBName)
	}
}

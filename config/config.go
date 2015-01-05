// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package config takes care of the configuration file parsing.
package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
	"time"
)

// sslModes corresponds to the SSL modes available for the connection to the
// PostgreSQL database.
// See http://www.postgresql.org/docs/9.4/static/libpq-ssl.html for details.
var sslModes = map[string]bool{
	"disable":     true,
	"require":     true,
	"verify-ca":   true,
	"verify-full": true,
}

// Config is the main configuration structure.
type Config struct {
	// CloneDir is the path to the folder where all repositories are cloned.
	CloneDir string `json:"clone_dir"`

	// Crawlers is a group of crawlers configuration.
	Crawlers []CrawlerConfig `json:"crawlers"`

	// CrawlingTimeInterval corresponds to the time to wait between 2 full
	// crawling periods.
	CrawlingTimeInterval string `json:"crawling_time_interval"`

	// FetchTimeInterval corresponds to the time to wait betweeb 2 full
	// repositories fetching periods.
	FetchTimeInterval string `json:"fetch_time_interval"`

	// Database is the database configuration.
	Database DatabaseConfig `json:"database"`
}

// CrawlerConfig is a configuration for a crawler.
type CrawlerConfig struct {
	// Type defines the crawler type (eg: "github").
	Type string `json:"type"`

	// Languages is the list of programming languages of interest.
	Languages []string `json:"languages"`

	// Limit limits the number of repositories to crawl. Set this value to 0 to
	// not use a limit. Otherwise, crawling will stop when "limit" repositories
	// have been fetched.
	// Note that the behavior is slightly different whether UseSearchAPI is set
	// to true or not. When using the search API, this limit correspond to the
	// number of repositories to crawl per language listed in "languages".
	// Otherwhise, this is a global limit, regardless of the language.
	Limit int64 `json:"limit"`

	// Fork indicate whether "fork" repositories need to be crawled or not.
	Fork bool `json:"fork"`

	// OAuthAccessToken is the API token. If not provided, crawld will work but
	// the number of API call is usually limited to a low number.
	// For instance, in the case of the GitHub crawler, unauthenticated
	// requests are limited to 60 per hour where authenticated requests goes up
	// to 5000 per hour.
	OAuthAccessToken string `json:"oauth_access_token"`
	UseSearchAPI     bool   `json:"use_search_api"`
}

// DatabaseConfig is a configuration for PostgreSQL database connection
// information
type DatabaseConfig struct {
	// HostName is the hostname, or IP address, of the database server.
	HostName string `json:"hostname"`

	// Port is the PostgreSQL port.
	Port uint `json:"port"`

	// UserName is the PostgreSQL user that has access to the database.
	UserName string `json:"username"`

	// Password is the password of the database user.
	Password string `json:"password"`

	// DBName is the database name.
	DBName string `json:"dbname"`

	// SSLMode defines the SSL mode for the connection to the database.
	// Refer to sslModes for the possible values and their meaning.
	SSLMode string `json:"ssl_mode"`
}

// ReadConfig reads a JSON formatted configuration file, verifies the values
// of the configuration parameters and fills the Config structure.
func ReadConfig(path string) (*Config, error) {
	// TODO maybe use a safer function like io.Copy
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err := json.Unmarshal(bs, cfg); err != nil {
		return nil, err
	}

	if err := cfg.verify(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c Config) verify() error {
	if len(strings.Trim(c.CloneDir, " ")) == 0 {
		return errors.New("clone_dir cannot be empty")
	}

	if _, err := time.ParseDuration(c.CrawlingTimeInterval); err != nil {
		return errors.New("config: invalid crawling time interval format")
	}

	if _, err := time.ParseDuration(c.FetchTimeInterval); err != nil {
		return errors.New("config: invalid fetch time interval format")
	}

	for _, cs := range c.Crawlers {
		if err := cs.verify(); err != nil {
			return err
		}
	}

	if err := c.Database.verify(); err != nil {
		return err
	}

	return nil
}

func (cc CrawlerConfig) verify() error {
	if len(strings.Trim(cc.Type, " ")) == 0 {
		return errors.New("config: crawler type cannot be empty")
	}

	if len(cc.Languages) == 0 {
		return errors.New("config: crawler must have at least one language")
	}

	return nil
}

func (dc DatabaseConfig) verify() error {
	if len(strings.Trim(dc.HostName, " ")) == 0 {
		return errors.New("config: database hostname cannot be empty")
	}

	if dc.Port <= 0 {
		return errors.New("config: database port must be greater than 0")
	}

	if len(strings.Trim(dc.UserName, " ")) == 0 {
		return errors.New("config: database username cannot be empty")
	}

	if len(strings.Trim(dc.DBName, " ")) == 0 {
		return errors.New("config: database name cannot be empty")
	}

	if _, ok := sslModes[dc.SSLMode]; !ok {
		return errors.New("config: database can only be disable, require, verify-ca or verify-full")
	}

	return nil
}

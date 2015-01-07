// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package git defines a Git repository type that implements the repo.Repo
// interface.
package git

import (
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
)

// GitRepo implements the Repo interface.
type GitRepo struct {
	absPath string
	gitBin  string
	url     string
}

// New creates a new GitRepo.
func New(absPath string, url string) (*GitRepo, error) {
	path, err := exec.LookPath("git")
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	return &GitRepo{absPath: absPath, gitBin: path, url: url}, nil
}

// AbsPath implements the AbsPath() method of the Repo interface.
func (gr GitRepo) AbsPath() string {
	return gr.absPath
}

// URL implements the URL() method of the Repo interface.
func (gr GitRepo) URL() string {
	return gr.url
}

// Clone implements the Clone() method of the Repo interface.
func (gr GitRepo) Clone() error {
	if ok, err := gr.isAvailable(); err != nil {
		return err
	} else if !ok {
		glog.Info(gr.url, " not available")
		return nil
	}

	out, err := exec.Command(gr.gitBin, "clone", "--quiet", gr.url, gr.absPath).CombinedOutput()
	glog.Info(string(out))
	if err != nil {
		glog.Error(err)
		return err
	}

	return nil
}

// Update implements the Update() method of the Repo interface.
func (gr GitRepo) Update() error {
	if ok, err := gr.isAvailable(); err != nil {
		return err
	} else if !ok {
		glog.Info(gr.url, " not available")
		return nil
	}

	err := os.Chdir(gr.absPath)
	if err != nil {
		glog.Error(err)
		return err
	}

	out, err := exec.Command(gr.gitBin, "pull", "--quiet").CombinedOutput()
	glog.Info(string(out))
	if err != nil {
		glog.Error(err)
		return err
	}

	return nil
}

// isAvailable queries the git repository in order to determine whether a the
// repository exists and is publicly available.
func (gr GitRepo) isAvailable() (bool, error) {
	u, err := url.Parse(gr.url)
	if err != nil {
		return false, err
	}

	switch u.Scheme {
	case "http", "https":
		queryURL := cleanURL(gr.url)
		if !hasGitExt(u.Path) {
			queryURL += ".git"
		}

		resp, err := http.Get(gr.url + "/git-upload-pack")
		if err != nil {
			glog.Warning(err)
			return false, nil
		}

		if resp.StatusCode != http.StatusOK {
			glog.Warningf("invalid HTTP status: expected %d, received %d",
				http.StatusOK, resp.StatusCode)
			return false, nil
		}
	default:
		glog.Warningf("protocol %s not supported", u.Scheme)
		return false, nil
	}

	return true, nil
}

// hasGitExt returns true if the path ends with a ".git" extension,
// false otherwise.
func hasGitExt(path string) bool {
	return filepath.Ext(path) == ".git"
}

// cleanURL removes the trailing slashes of an URL, if any.
func cleanURL(url string) string {
	return strings.TrimSuffix(url, "/")
}

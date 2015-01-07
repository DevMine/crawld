// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package git defines a Git repository type that implements the repo.Repo
// interface.
package git

import (
	"os"
	"os/exec"

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

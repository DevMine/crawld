// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package git

import (
	"os"
	"os/exec"

	"github.com/golang/glog"
)

type GitRepo struct {
	absPath string
	gitBin  string
	url     string
}

func New(absPath string, url string) (*GitRepo, error) {

	path, err := exec.LookPath("git")
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	return &GitRepo{absPath: absPath, gitBin: path, url: url}, nil
}

func (gr GitRepo) AbsPath() string {
	return gr.absPath
}

func (gr GitRepo) URL() string {
	return gr.url
}

func (gr GitRepo) Clone() error {

	out, err := exec.Command(gr.gitBin, "clone", gr.url, gr.absPath).CombinedOutput()
	glog.Info(string(out))
	if err != nil {
		glog.Error(err)
		return err
	}

	return nil
}

func (gr GitRepo) Update() error {

	err := os.Chdir(gr.absPath)
	if err != nil {
		glog.Error(err)
		return err
	}

	out, err := exec.Command(gr.gitBin, "pull").CombinedOutput()
	glog.Info(string(out))
	if err != nil {
		glog.Error(err)
		return err
	}

	return nil
}

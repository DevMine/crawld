// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repo

import (
	"errors"

	g2g "github.com/libgit2/git2go"
)

// gitRepo implements the Repo interface.
type gitRepo struct {
	absPath string
	r       *g2g.Repository
	url     string
}

// newGitRepo creates a new GitRepo. GitRepo implements the Repo interface
// for a git repository.
func newGitRepo(absPath string, url string) (*gitRepo, error) {
	// attempt opening the repository as it may already exist
	// ignore if it fails since it will be created at first call to Clone()
	r, _ := g2g.OpenRepository(absPath)

	return &gitRepo{absPath: absPath, url: url, r: r}, nil
}

// AbsPath implements the AbsPath() method of the Repo interface.
func (gr gitRepo) AbsPath() string {
	return gr.absPath
}

// URL implements the URL() method of the Repo interface.
func (gr gitRepo) URL() string {
	return gr.url
}

// Clone implements the Clone() method of the Repo interface.
func (gr gitRepo) Clone() error {
	var err error

	gr.r, err = g2g.Clone(gr.url, gr.absPath, &g2g.CloneOptions{})
	if err != nil {
		return g2gErrorToRepoError(err)
	}

	return nil
}

// Update implements the Update() method of the Repo interface.
// It fetches changes from remote and performs a fast-forward on the local
// branch so as to match the remote branch.
func (gr gitRepo) Update() error {
	var err error

	if gr.r == nil {
		gr.r, err = g2g.OpenRepository(gr.absPath)
		if err != nil {
			return g2gErrorToRepoError(err)
		}
	}

	origin, err := gr.r.LookupRemote("origin")
	if err != nil {
		return g2gErrorToRepoError(err)
	}

	if err = origin.Fetch([]string{}, nil, ""); err != nil {
		return g2gErrorToRepoError(err)
	}

	ref, err := gr.r.Head()
	if err != nil {
		return g2gErrorToRepoError(err)
	}

	if !ref.IsBranch() {
		return errors.New("repository reference is not a branch (likely in a detached HEAD state)")
	}

	remoteRef, err := ref.Branch().Upstream()
	if err != nil {
		return g2gErrorToRepoError(err)
	}
	if _, err = ref.SetTarget(remoteRef.Target(), nil, "pull: Fast-forward"); err != nil {
		return g2gErrorToRepoError(err)
	}

	var checkoutOpts g2g.CheckoutOpts
	checkoutOpts.Strategy = g2g.CheckoutForce

	if err = gr.r.CheckoutHead(&checkoutOpts); err != nil {
		return g2gErrorToRepoError(err)
	}

	return nil
}

// Cleanup implements the Cleanup() method of the Repo interface.
func (gr gitRepo) Cleanup() error {
	if gr.r != nil {
		gr.r.Free()
	}
	return nil
}

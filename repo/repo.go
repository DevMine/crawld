// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package repo defines a generic interface for Version Control Systems (VCS).
package repo

import (
	"errors"
)

// Repo abstracts a version control system (VCS) such as git, mercurial or
// others..
type Repo interface {
	// Clone clones a repository into a new directory.
	// Clone must return ErrNetworkUnreachable in case of connectivity
	// problems and ErrNoSpace in case of storage space problems.
	Clone() error

	// Update fetches the latest changes from a repository, using the
	// default branch.
	// Update must return ErrNetworkUnreachable in case of connectivity
	// problems and ErrNoSpace in case of storage space problems.
	Update() error

	// AbsPath gives the absolute path to the repository on disk.
	AbsPath() string

	// SetAbsPath can be used to change AbsPath, if necessary.
	SetAbsPath(path string)

	// URL gives the clone URL of the repository.
	URL() string

	// Cleanup shall be called when done using the Repo. It will take
	// care of closing any open files and the usual housekeeping.
	Cleanup() error
}

// New creates a new repository. vcsType corresponds to the VCS type
// (currently, only 'git' is supported) whereas clonePath corresponds to the
// absolute path to/for the repository on disk and cloneURL is the URL used
// for cloning/updating the repository.
func New(vcsType, clonePath string, cloneURL string) (Repo, error) {
	var newRepo Repo
	var err error

	switch vcsType {
	case "git":
		newRepo, err = newGitRepo(clonePath, cloneURL)
	default:
		return nil, errors.New("unsupported vcs repository type: " + vcsType)
	}
	if err != nil {
		return nil, err
	}

	return newRepo, nil
}

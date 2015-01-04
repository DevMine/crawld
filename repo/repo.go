// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repo

// Repo abstracts a version control system (VCS) such as git, mercurial or
// others..
type Repo interface {
	// Clone clones a repository into a new directory.
	Clone() error

	// Update fetches the latest changes from a repository, using the
	// default branch.
	Update() error

	// AbsPath gives the absolute path to the repository on disk.
	AbsPath() string

	// URL gives the clone URL of the repository.
	URL() string
}

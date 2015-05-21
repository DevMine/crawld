// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repo

import (
	"errors"

	g2g "github.com/libgit2/git2go"
)

var (
	// ErrNetwork represents any type of network error.
	ErrNetwork = errors.New("network error")

	// ErrNoSpace represents a space storage error.
	ErrNoSpace = errors.New("no space left on device")
)

// g2gErrorToRepoError returns a repo error when given a git2go error if it
// it finds a corresponding match or simply the given error otherwise.
// TODO when git2go adds support for ENOSPC type of error, update this method
// accordingly to return ErrNoSpace.
func g2gErrorToRepoError(err error) error {
	if g2g.IsErrorClass(err, g2g.ErrClassNet) {
		return ErrNetwork
	}
	return err
}

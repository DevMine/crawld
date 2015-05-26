// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errbag

import (
	"errors"
	"testing"
	"time"
)

func TestErrBag(t *testing.T) {
	var waitTime, errBagSize, leakInterval uint

	// test some bad input
	waitTime, errBagSize, leakInterval = 5, 60, 99
	if _, err := New(waitTime, errBagSize, leakInterval); err == nil {
		t.Fatal("leak interval < 100 shall not be permitted")
	}
	leakInterval = 1000
	errBagSize = 0
	if _, err := New(waitTime, errBagSize, leakInterval); err == nil {
		t.Fatal("errBagSize of 0 shall not be permitted")
	}
	errBagSize = 60
	waitTime = 0
	if _, err := New(waitTime, errBagSize, leakInterval); err == nil {
		t.Fatal("waitTime of 0 shall not be permitted")
	}

	waitTime = 5
	errBag, err := New(waitTime, errBagSize, leakInterval)
	if err != nil {
		t.Fatal(err)
	}

	errBag.Inflate()

	err = errors.New("foo error")
	var i uint

	// test that it does not block on less than 1 error per second
	start := time.Now()
	for i = 0; i < 4; i++ {
		errBag.Record(err)
	}
	elapsed := time.Since(start)
	if elapsed > time.Second*time.Duration(waitTime) {
		t.Fatal("throttling when error rate is low")
	}

	// now test throttling
	start = time.Now()
	for i = 0; i < errBagSize+3; i++ {
		errBag.Record(err)
	}
	elapsed = time.Since(start)
	if elapsed < time.Second*time.Duration(waitTime) {
		t.Fatal("failed to throttle")
	}

	// errBag is full of errors, deflate shall empty it
	errBag.Deflate()

	// attempting to record errors now shall panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("call to Record() shall panic")
		}
	}()
	errBag.Record(err)
}

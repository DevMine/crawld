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
	var waitTime, slidingWindow uint
	var leakRate float64

	waitTime, slidingWindow, leakRate = 5, 60, -1.0

	if _, err := New(waitTime, slidingWindow, leakRate); err == nil {
		t.Fatal("negative leak rate shall not be permitted")
	}

	leakRate = 1.0
	errBag, err := New(waitTime, slidingWindow, leakRate)
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
	for i = 0; i < slidingWindow+3; i++ {
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

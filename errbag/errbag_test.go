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

	// attempt recording an error without specifying a callback function
	// (it is expected to be valid)
	err = errors.New("foo error")
	errBag.Record(err, nil)

	var i uint
	// test that it does not block on less than 1 error per second
	start := time.Now()
	for i = 0; i < 2; i++ {
		errBag.Record(err, func(status Status) {
			if status.State != StatusOK {
				t.Error(errors.New("expected StatusOK"))
			}
			if status.WaitTime != 0 {
				t.Error(errors.New("no wait time expected"))
			}
		})
	}
	elapsed := time.Since(start)
	if elapsed > time.Second*time.Duration(waitTime) {
		t.Fatal("throttling when error rate is low")
	}

	// make sure the error pipeline is empty before starting new test
	// (we recorded 3 errors until now)
	time.Sleep(time.Duration(leakInterval) * time.Millisecond * (2 + 1))
	// kill errLeak routine to prevent error leaking
	errBag.done <- struct{}{}
	// make sure it has had time to stop
	time.Sleep(time.Duration(500) * time.Millisecond)

	// now test throttling
	start = time.Now()
	for i = 0; i < errBagSize+1; i++ {
		if i == errBagSize {
			// now that the bag is full, it shall throttle if attempting to
			// record a new error
			errBag.Record(err, func(status Status) {
				if status.State != StatusThrottling {
					t.Error(errors.New("expected StatusThrottling"))
				}
				if status.WaitTime != waitTime {
					t.Error(errors.New("expected different WaitTime"))
				}
			})
		} else {
			errBag.Record(err, nil)
		}
	}

	elapsed = time.Since(start)
	if elapsed < time.Second*time.Duration(waitTime) {
		t.Fatal("failed to throttle")
	}

	// since we stopped errLeak earlier to prevent leaking, restart it here
	errBag.Inflate()

	// errBag is full of errors, deflate shall empty it
	errBag.Deflate()

	// make sure it has had time to deflate
	time.Sleep(time.Duration(leakInterval) * time.Millisecond)

	// attempting to record errors now shall panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("call to Record() shall panic")
		}
	}()
	errBag.Record(err, nil)
}

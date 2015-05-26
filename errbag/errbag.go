// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package errbag implements an error rate based throttler. It can be used to
// to limit function calls rate once a certain error rate threshold has been
// reached.
package errbag

import (
	"errors"
	"time"
)

// ErrBag is very effective at preventing an error rate to reach a
// certain threshold.
type ErrBag struct {
	waitTime uint
	leakRate float64
	errChan  chan error
	done     chan struct{}
}

// New creates a new ErrBag, for safety purpose. waitTime corresponds to the
// number of seconds to wait when the error rate threshold is reached.
// slidingWindow is, in seconds, the size of the sliding window to consider
// for throttling. leakRate corresponds to the rate at which errors are leaked
// from the errbag in terms of errors per second. At a rate of 1, it will take
// exactly slidingWindow seconds to empty the errbag if it is full, considering
// no other errors are recorded during that time.
func New(waitTime, slidingWindow uint, leakRate float64) (*ErrBag, error) {
	if leakRate <= 0 {
		return nil, errors.New("leakRate cannot be less than or equal to 0")
	}
	// channels are closed when Deflate() is invoked
	errChan := make(chan error, slidingWindow)
	done := make(chan struct{}, 1)
	return &ErrBag{waitTime: waitTime, leakRate: leakRate, errChan: errChan, done: done}, nil
}

// Inflate needs to be called once to prepare the ErrBag. Once the ErrBag
// is not needed anymore, a proper call to Deflate() shall be made.
func (eb ErrBag) Inflate() {
	go func() {
		go eb.errLeak()
		// wait for the exit signal
		<-eb.done
		// by returning, we also kill the child goroutine (errLeak())
		return
	}()
}

// Deflate needs to be called when the errbag is of no use anymore.
// Calling Record() with a deflated errbag will induce a panic.
func (eb ErrBag) Deflate() {
	eb.done <- struct{}{}
	close(eb.done)
	close(eb.errChan)
}

// Record records an error if its value is non nil. It shall be called
// by any function returning an error in order to properly rate limit the
// errors produced. RecordError will wait for waitTime minutes if the error
// rate is too high.
// Note that record will panic if called after Deflate() has been called.
func (eb ErrBag) Record(err error) {
	if err != nil {
		select {
		case eb.errChan <- err:
		default:
			time.Sleep(time.Second * time.Duration(eb.waitTime))
		}
	}
}

// errLeak leaks error from the errbag at leakRate until the error channel
// is closed.
func (eb ErrBag) errLeak() {
	for _, ok := <-eb.errChan; ok; _, ok = <-eb.errChan {
		time.Sleep(time.Second * time.Duration(eb.leakRate))
	}
}

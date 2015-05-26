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
	waitTime     uint
	leakInterval uint
	errChan      chan struct{}
	done         chan struct{}
}

// New creates a new ErrBag, for safety purpose. waitTime corresponds to the
// number of seconds to wait when the error rate threshold is reached.
// errBagSize is, in seconds, the size of the sliding window to consider
// for throttling. You can see it as the size of the errbag. The larger it is,
// the larger the window to consider for error rate is. Consider this value
// along with the leakInterval. leakInterval corresponds to the time to wait,
// in milliseconds, before an error is discarded from the errbag. It must be
// equal or greater than 100, otherwise throttling will be ineffective.
func New(waitTime, errBagSize, leakInterval uint) (*ErrBag, error) {
	if waitTime == 0 {
		return nil, errors.New("setting waitTime to 0 would prevent throttling")
	}
	if errBagSize == 0 {
		return nil, errors.New("setting errBagSize to 0 would prevent throttling")
	}
	if leakInterval < 100 {
		return nil, errors.New("leakInterval must be greater than 100")
	}

	// channels are closed when Deflate() is invoked
	errChan := make(chan struct{}, errBagSize)
	done := make(chan struct{}, 1)
	return &ErrBag{waitTime: waitTime, leakInterval: leakInterval, errChan: errChan, done: done}, nil
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
// errors produced. RecordError will wait for waitTime seconds if the error
// rate is too high.
// Note that record will panic if called after Deflate() has been called.
func (eb ErrBag) Record(err error) {
	if err != nil {
		select {
		case eb.errChan <- struct{}{}:
		default:
			time.Sleep(time.Second * time.Duration(eb.waitTime))
		}
	}
}

// errLeak leaks error from the errbag at leakInterval until the error channel
// is closed.
func (eb ErrBag) errLeak() {
	for _, ok := <-eb.errChan; ok; _, ok = <-eb.errChan {
		time.Sleep(time.Millisecond * time.Duration(eb.leakInterval))
	}
}

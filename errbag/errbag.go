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

// Status structure is used as argument to CallbackFunc. It indicates the
// the sate of the errbag after having recorded an error.
type Status struct {
	// State indicates whether throttling had to be activated after an error
	// has been recorded (StatusThrottling) or if it was simply registered and
	// all is well (StatusOK).
	State int

	// WaitTime indicates for how long the Record() method will wait before
	// being available to record new errors.
	WaitTime uint
}

// CallbackFunc is used as an argument to the Record() method.
type CallbackFunc func(status Status)

const (
	// StatusThrottling indicates the errbag is throttling.
	StatusThrottling = iota

	// StatusOK indicates that all is well.
	StatusOK
)

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
	ready := make(chan bool)
	go func() {
		ready <- true
		eb.errLeak()
	}()
	// wait for the routine to be running
	<-ready
	close(ready)
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
// callback purpose is for the caller to be informed about the errbag status
// after an error has been recorded in order to help take the appropriate
// actions. nil can be passed if the caller is not interested in the status.
// Note that record will panic if called after Deflate() has been called.
func (eb ErrBag) Record(err error, callback CallbackFunc) {
	if err != nil {
		select {
		case eb.errChan <- struct{}{}:
			if callback != nil {
				callback(Status{State: StatusOK})
			}
		default:
			if callback != nil {
				callback(Status{State: StatusThrottling, WaitTime: eb.waitTime})
			}
			time.Sleep(time.Second * time.Duration(eb.waitTime))
		}
	}
}

// errLeak leaks error from the errbag at leakInterval until the error channel
// is closed.
func (eb ErrBag) errLeak() {
	for {
		select {
		case <-eb.done:
			return
		case <-eb.errChan:
			time.Sleep(time.Millisecond * time.Duration(eb.leakInterval))
		}
	}
}

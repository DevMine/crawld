// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crawlers

import (
	"bytes"
	"errors"
	"strings"
)

var (
	errTooManyCall      = errors.New("API rate limit exceeded")
	errUnavailable      = errors.New("resource unavailable")
	errRuntime          = errors.New("runtime error")
	errInvalidArgs      = errors.New("invalid arguments")
	errNilArg           = errors.New("nil argument")
	errInvalidParamType = errors.New("invalid parameter type")
)

type invalidStructError struct {
	message string
	fields  []string
}

func newInvalidStructError(msg string) *invalidStructError {
	return &invalidStructError{message: msg, fields: []string{}}
}

func (e *invalidStructError) AddField(f string) *invalidStructError {
	e.fields = append(e.fields, f)
	return e
}

func (e invalidStructError) FieldsLen() int {
	return len(e.fields)
}

func (e invalidStructError) Error() string {
	buf := bytes.NewBufferString(e.message)
	buf.WriteString("{ ")
	buf.WriteString(strings.Join(e.fields, ", "))
	buf.WriteString(" }\n")

	return buf.String()
}

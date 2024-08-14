// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lazyregexp is a thin wrapper over regexp, allowing the use of global
// regexp variables without forcing them to be compiled at init.
package lazyregexp

import (
	"os"
	"strings"
	"sync"

	"github.com/grafana/regexp"
)

// Regexp is a wrapper around regexp.Regexp, where the underlying regexp will be
// compiled the first time it is needed.
type Regexp struct {
	str   string
	posix bool
	once  sync.Once
	rx    *regexp.Regexp
}

func (r *Regexp) Re() *regexp.Regexp {
	r.once.Do(r.build)
	return r.rx
}

func (r *Regexp) build() {
	if r.posix {
		r.rx = regexp.MustCompilePOSIX(r.str)
	} else {
		r.rx = regexp.MustCompile(r.str)
	}
	r.str = ""
}

func (r *Regexp) FindSubmatch(s []byte) [][]byte {
	return r.Re().FindSubmatch(s)
}

func (r *Regexp) FindStringSubmatch(s string) []string {
	return r.Re().FindStringSubmatch(s)
}

func (r *Regexp) FindStringSubmatchIndex(s string) []int {
	return r.Re().FindStringSubmatchIndex(s)
}

func (r *Regexp) ReplaceAllString(src, repl string) string {
	return r.Re().ReplaceAllString(src, repl)
}

func (r *Regexp) FindString(s string) string {
	return r.Re().FindString(s)
}

func (r *Regexp) FindAllString(s string, n int) []string {
	return r.Re().FindAllString(s, n)
}

func (r *Regexp) MatchString(s string) bool {
	return r.Re().MatchString(s)
}

func (r *Regexp) SubexpNames() []string {
	return r.Re().SubexpNames()
}

func (r *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
	return r.Re().FindAllStringSubmatch(s, n)
}

func (r *Regexp) Split(s string, n int) []string {
	return r.Re().Split(s, n)
}

func (r *Regexp) ReplaceAllLiteralString(src, repl string) string {
	return r.Re().ReplaceAllLiteralString(src, repl)
}

func (r *Regexp) FindAllIndex(b []byte, n int) [][]int {
	return r.Re().FindAllIndex(b, n)
}

func (r *Regexp) FindAll(b []byte, n int) [][]byte {
	return r.Re().FindAll(b, n)
}

func (r *Regexp) Match(b []byte) bool {
	return r.Re().Match(b)
}

func (r *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
	return r.Re().ReplaceAllStringFunc(src, repl)
}

func (r *Regexp) ReplaceAll(src, repl []byte) []byte {
	return r.Re().ReplaceAll(src, repl)
}

func (r *Regexp) SubexpIndex(s string) int {
	return r.Re().SubexpIndex(s)
}

var inTest = len(os.Args) > 0 && strings.HasSuffix(strings.TrimSuffix(os.Args[0], ".exe"), ".test")

// New creates a new lazy regexp, delaying the compiling work until it is first
// needed. If the code is being run as part of tests, the regexp compiling will
// happen immediately.
func New(str string) *Regexp {
	lr := &Regexp{str: str}
	if inTest {
		// In tests, always compile the regexps early.
		lr.Re()
	}
	return lr
}

// NewPOSIX creates a new lazy regexp, delaying the compiling work until it is
// first needed. If the code is being run as part of tests, the regexp
// compiling will happen immediately.
func NewPOSIX(str string) *Regexp {
	lr := &Regexp{str: str, posix: true}
	if inTest {
		// In tests, always compile the regexps early.
		lr.Re()
	}
	return lr
}

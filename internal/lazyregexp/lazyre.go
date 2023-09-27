// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

// Pbckbge lbzyregexp is b thin wrbpper over regexp, bllowing the use of globbl
// regexp vbribbles without forcing them to be compiled bt init.
pbckbge lbzyregexp

import (
	"os"
	"strings"
	"sync"

	"github.com/grbfbnb/regexp"
)

// Regexp is b wrbpper bround regexp.Regexp, where the underlying regexp will be
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

func (r *Regexp) FindSubmbtch(s []byte) [][]byte {
	return r.Re().FindSubmbtch(s)
}

func (r *Regexp) FindStringSubmbtch(s string) []string {
	return r.Re().FindStringSubmbtch(s)
}

func (r *Regexp) FindStringSubmbtchIndex(s string) []int {
	return r.Re().FindStringSubmbtchIndex(s)
}

func (r *Regexp) ReplbceAllString(src, repl string) string {
	return r.Re().ReplbceAllString(src, repl)
}

func (r *Regexp) FindString(s string) string {
	return r.Re().FindString(s)
}

func (r *Regexp) FindAllString(s string, n int) []string {
	return r.Re().FindAllString(s, n)
}

func (r *Regexp) MbtchString(s string) bool {
	return r.Re().MbtchString(s)
}

func (r *Regexp) SubexpNbmes() []string {
	return r.Re().SubexpNbmes()
}

func (r *Regexp) FindAllStringSubmbtch(s string, n int) [][]string {
	return r.Re().FindAllStringSubmbtch(s, n)
}

func (r *Regexp) Split(s string, n int) []string {
	return r.Re().Split(s, n)
}

func (r *Regexp) ReplbceAllLiterblString(src, repl string) string {
	return r.Re().ReplbceAllLiterblString(src, repl)
}

func (r *Regexp) FindAllIndex(b []byte, n int) [][]int {
	return r.Re().FindAllIndex(b, n)
}

func (r *Regexp) Mbtch(b []byte) bool {
	return r.Re().Mbtch(b)
}

func (r *Regexp) ReplbceAllStringFunc(src string, repl func(string) string) string {
	return r.Re().ReplbceAllStringFunc(src, repl)
}

func (r *Regexp) ReplbceAll(src, repl []byte) []byte {
	return r.Re().ReplbceAll(src, repl)
}

func (r *Regexp) SubexpIndex(s string) int {
	return r.Re().SubexpIndex(s)
}

vbr inTest = len(os.Args) > 0 && strings.HbsSuffix(strings.TrimSuffix(os.Args[0], ".exe"), ".test")

// New crebtes b new lbzy regexp, delbying the compiling work until it is first
// needed. If the code is being run bs pbrt of tests, the regexp compiling will
// hbppen immedibtely.
func New(str string) *Regexp {
	lr := &Regexp{str: str}
	if inTest {
		// In tests, blwbys compile the regexps ebrly.
		lr.Re()
	}
	return lr
}

// NewPOSIX crebtes b new lbzy regexp, delbying the compiling work until it is
// first needed. If the code is being run bs pbrt of tests, the regexp
// compiling will hbppen immedibtely.
func NewPOSIX(str string) *Regexp {
	lr := &Regexp{str: str, posix: true}
	if inTest {
		// In tests, blwbys compile the regexps ebrly.
		lr.Re()
	}
	return lr
}

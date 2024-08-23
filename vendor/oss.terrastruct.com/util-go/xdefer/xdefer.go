// Package xdefer implements an extremely useful function, Errorf, to annotate all errors returned from a function transparently.
package xdefer

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

type deferError struct {
	s     string
	err   error
	frame xerrors.Frame
}

var _ interface {
	xerrors.Wrapper
	xerrors.Formatter
	Is(error) bool
} = deferError{}

func (e deferError) Unwrap() error {
	return e.err
}

func (e deferError) Format(f fmt.State, c rune) {
	xerrors.FormatError(e, f, c)
}

// Used to detect if there is a duplicate frame as a result
// of using xdefer and if so to ignore it.
type fakeXerrorsPrinter struct {
	s []string
}

func (fp *fakeXerrorsPrinter) Print(v ...interface{}) {
	fp.s = append(fp.s, fmt.Sprint(v...))
}

func (fp *fakeXerrorsPrinter) Printf(f string, v ...interface{}) {
	fp.s = append(fp.s, fmt.Sprintf(f, v...))
}

func (fp *fakeXerrorsPrinter) Detail() bool {
	return true
}

func (e deferError) shouldPrintFrame(p xerrors.Printer) bool {
	fm, ok := e.err.(xerrors.Formatter)
	if !ok {
		return true
	}

	fp := &fakeXerrorsPrinter{}
	e.frame.Format(fp)
	fp2 := &fakeXerrorsPrinter{}
	_ = fm.FormatError(fp2)
	if len(fp.s) >= 2 && len(fp2.s) >= 3 {
		if fp.s[1] == fp2.s[2] {
			// We don't need to print our frame into the real
			// xerrors printer as the next error will have it.
			return false
		}
	}
	return true
}

func (e deferError) FormatError(p xerrors.Printer) error {
	if e.s == "" {
		if e.shouldPrintFrame(p) {
			e.frame.Format(p)
		}
		return e.err
	}

	p.Print(e.s)
	if p.Detail() && e.shouldPrintFrame(p) {
		e.frame.Format(p)
	}
	return e.err
}

func (e deferError) Is(err error) bool {
	return xerrors.Is(e.err, err)
}

func (e deferError) Error() string {
	if e.s == "" {
		fp := &fakeXerrorsPrinter{}
		e.frame.Format(fp)
		if len(fp.s) < 1 {
			return e.err.Error()
		}
		return fmt.Sprintf("%v: %v", strings.TrimSpace(fp.s[0]), e.err)
	}
	return fmt.Sprintf("%v: %v", e.s, e.err)
}

// Errorf makes it easy to defer annotate an error for all return paths in a function.
// See the tests for how it's used.
//
// Pass s == "" to only annotate the location of the return.
func Errorf(err *error, s string, v ...interface{}) {
	if *err != nil {
		*err = deferError{
			s:     fmt.Sprintf(s, v...),
			err:   *err,
			frame: xerrors.Caller(1),
		}
	}
}

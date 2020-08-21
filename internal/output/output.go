// Package output provides types related to formatted terminal output.
package output

import (
	"fmt"
	"io"
	"sync"

	"github.com/mattn/go-runewidth"
)

// Writer defines a common set of methods that can be used to output status
// information.
//
// Note that the *f methods can accept Style instances in their arguments with
// the %s format specifier: if given, the detected colour support will be
// respected when outputting.
type Writer interface {
	// These methods only write the given message if verbose mode is enabled.
	Verbose(s string)
	Verbosef(format string, args ...interface{})
	VerboseLine(line FancyLine)

	// These methods write their messages unconditionally.
	Write(s string)
	Writef(format string, args ...interface{})
	WriteLine(line FancyLine)
}

// Output encapsulates a standard set of functionality for commands that need
// to output human-readable data.
//
// Output is not appropriate for machine-readable data, such as JSON.
type Output struct {
	w    io.Writer
	caps capabilities
	opts OutputOpts

	// Unsurprisingly, it would be bad if multiple goroutines wrote at the same
	// time, so we have a basic mutex to guard against that.
	lock sync.Mutex
}

type OutputOpts struct {
	// ForceColor ignores all terminal detection and enabled coloured output.
	ForceColor bool
	Verbose    bool
}

// newOutputPlatformQuirks provides a way for conditionally compiled code to
// hook into NewOutput to perform any required setup.
var newOutputPlatformQuirks func(o *Output) error

func NewOutput(w io.Writer, opts OutputOpts) *Output {
	caps := detectCapabilities()
	if opts.ForceColor {
		caps.Color = true
	}

	o := &Output{caps: caps, opts: opts, w: w}
	if newOutputPlatformQuirks != nil {
		if err := newOutputPlatformQuirks(o); err != nil {
			o.Verbosef("Error handling platform quirks: %v", err)
		}
	}

	return o
}

func (o *Output) Verbose(s string) {
	if o.opts.Verbose {
		o.Write(s)
	}
}

func (o *Output) Verbosef(format string, args ...interface{}) {
	if o.opts.Verbose {
		o.Writef(format, args...)
	}
}

func (o *Output) VerboseLine(line FancyLine) {
	if o.opts.Verbose {
		o.WriteLine(line)
	}
}

func (o *Output) Write(s string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	fmt.Fprintln(o.w, s)
}

func (o *Output) Writef(format string, args ...interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()
	fmt.Fprintf(o.w, format, o.caps.formatArgs(args)...)
	fmt.Fprint(o.w, "\n")
}

func (o *Output) WriteLine(line FancyLine) {
	o.lock.Lock()
	defer o.lock.Unlock()
	line.write(o.w, o.caps)
}

// Block starts a new block context. This should not be invoked if there is an
// active Pending or Progress context.
func (o *Output) Block(summary FancyLine) *Block {
	o.WriteLine(summary)
	return newBlock(runewidth.StringWidth(summary.emoji)+1, o)
}

// Pending sets up a new pending context. This should not be invoked if there
// is an active Block or Progress context. The emoji in the message will be
// ignored, as Pending will render its own spinner.
//
// A Pending instance must be disposed of via the Complete or Destroy methods.
func (o *Output) Pending(message FancyLine) Pending {
	return newPending(message, o)
}

// Progress sets up a new progress bar context. This should not be invoked if
// there is an active Block or Pending context.
//
// A Progress instance must be disposed of via the Complete or Destroy methods.
func (o *Output) Progress(bars []ProgressBar, opts *ProgressOpts) Progress {
	return newProgress(bars, o, opts)
}

// The utility functions below do not make checks for whether the terminal is a
// TTY, and should only be invoked from behind appropriate guards.

func (o *Output) clearCurrentLine() {
	fmt.Fprint(o.w, "\033[2K")
}

func (o *Output) moveDown(lines int) {
	fmt.Fprintf(o.w, "\033[%dB", lines)

	// Move the cursor to the leftmost column.
	fmt.Fprintf(o.w, "\033[%dD", o.caps.Width+1)
}

func (o *Output) moveUp(lines int) {
	fmt.Fprintf(o.w, "\033[%dA", lines)

	// Move the cursor to the leftmost column.
	fmt.Fprintf(o.w, "\033[%dD", o.caps.Width+1)
}

// writeStyle is a helper to write a style while respecting the terminal
// capabilities.
func (o *Output) writeStyle(style Style) {
	fmt.Fprintf(o.w, "%s", o.caps.formatArgs([]interface{}{style})...)
}

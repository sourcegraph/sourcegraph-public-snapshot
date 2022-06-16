// Package output provides types related to formatted terminal output.
package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/glamour"
	glamouransi "github.com/charmbracelet/glamour/ansi"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"

	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	Verbosef(format string, args ...any)
	VerboseLine(line FancyLine)

	// These methods write their messages unconditionally.
	Write(s string)
	Writef(format string, args ...any)
	WriteLine(line FancyLine)
}

type Context interface {
	Writer

	Close()
}

// Output encapsulates a standard set of functionality for commands that need
// to output human-readable data.
//
// Output is not appropriate for machine-readable data, such as JSON.
type Output struct {
	w       io.Writer
	caps    capabilities
	verbose bool

	// Unsurprisingly, it would be bad if multiple goroutines wrote at the same
	// time, so we have a basic mutex to guard against that.
	lock sync.Mutex
}

var _ sync.Locker = &Output{}

type OutputOpts struct {
	// ForceColor ignores all terminal detection and enabled coloured output.
	ForceColor bool
	// ForceTTY ignores all terminal detection and enables TTY output.
	ForceTTY bool

	// ForceHeight ignores all terminal detection and sets the height to this value.
	ForceHeight int
	// ForceWidth ignores all terminal detection and sets the width to this value.
	ForceWidth int

	// ForceDarkBackground ignores all terminal detection and sets whether the terminal
	// background is dark to this value.
	ForceDarkBackground bool

	Verbose bool
}

// newOutputPlatformQuirks provides a way for conditionally compiled code to
// hook into NewOutput to perform any required setup.
var newOutputPlatformQuirks func(o *Output) error

// newCapabilityWatcher returns a channel that receives a message when
// capabilities are updated. By default, no watching functionality is
// available.
var newCapabilityWatcher = func(opts OutputOpts) chan capabilities { return nil }

func NewOutput(w io.Writer, opts OutputOpts) *Output {
	caps, detectionErr := detectCapabilities(opts)

	o := &Output{caps: caps, verbose: opts.Verbose, w: w}
	if newOutputPlatformQuirks != nil {
		if err := newOutputPlatformQuirks(o); err != nil {
			o.Verbosef("Error handling platform quirks: %v", err)
		}
	}

	// If we got an error earlier, now is where we'll report it to the user.
	if detectionErr != nil {
		block := o.Block(Linef(EmojiWarning, StyleWarning, "An error was returned when detecting the terminal size and capabilities:"))
		block.Write("")
		block.Write(detectionErr.Error())
		block.Write("")
		block.Write("Execution will continue, but please report this, along with your operating")
		block.Write("system, terminal, and any other details, to:")
		block.Write("  https://github.com/sourcegraph/sourcegraph/issues/new")
		block.Close()
	}

	// Set up a watcher so we can adjust the size of the output if the terminal
	// is resized.
	if c := newCapabilityWatcher(opts); c != nil {
		go func() {
			for caps := range c {
				o.caps = caps
			}
		}()
	}

	return o
}

func (o *Output) Lock() {
	o.lock.Lock()

	// Hide the cursor while we update: this reduces the jitteriness of the
	// whole thing, and some terminals are smart enough to make the update we're
	// about to render atomic if the cursor is hidden for a short length of
	// time.
	o.w.Write([]byte("\033[?25l"))
}

func (o *Output) SetVerbose() {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.verbose = true
}

func (o *Output) UnsetVerbose() {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.verbose = false
}

func (o *Output) Unlock() {
	// Show the cursor once more.
	o.w.Write([]byte("\033[?25h"))

	o.lock.Unlock()
}

func (o *Output) Verbose(s string) {
	if o.verbose {
		o.Write(s)
	}
}

func (o *Output) Verbosef(format string, args ...any) {
	if o.verbose {
		o.Writef(format, args...)
	}
}

func (o *Output) VerboseLine(line FancyLine) {
	if o.verbose {
		o.WriteLine(line)
	}
}

func (o *Output) Write(s string) {
	o.Lock()
	defer o.Unlock()
	fmt.Fprintln(o.w, s)
}

func (o *Output) Writef(format string, args ...any) {
	o.Lock()
	defer o.Unlock()
	fmt.Fprintf(o.w, format, o.caps.formatArgs(args)...)
	fmt.Fprint(o.w, "\n")
}

func (o *Output) WriteLine(line FancyLine) {
	o.Lock()
	defer o.Unlock()
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

// ProgressWithStatusBars sets up a new progress bar context with StatusBar
// contexts. This should not be invoked if there is an active Block or Pending
// context.
//
// A Progress instance must be disposed of via the Complete or Destroy methods.
func (o *Output) ProgressWithStatusBars(bars []ProgressBar, statusBars []*StatusBar, opts *ProgressOpts) ProgressWithStatusBars {
	return newProgressWithStatusBars(bars, statusBars, o, opts)
}

type readWriter struct {
	io.Reader
	io.Writer
}

// PromptPassword tries to securely prompt a user for sensitive input.
func (o *Output) PromptPassword(input io.Reader, prompt FancyLine) (string, error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	// Render the prompt
	prompt.Prompt = true
	var promptText bytes.Buffer
	prompt.write(&promptText, o.caps)

	// If input is a file and terminal, read from it directly
	if f, ok := input.(*os.File); ok {
		fd := int(f.Fd())
		if term.IsTerminal(fd) {
			_, _ = o.w.Write(promptText.Bytes())
			val, err := term.ReadPassword(fd)
			_, _ = o.w.Write([]byte("\n")) // once we've read an input
			return string(val), err
		}
	}

	// Otherwise, create a terminal
	t := term.NewTerminal(&readWriter{Reader: input, Writer: o.w}, "")
	_ = t.SetSize(o.caps.Width, o.caps.Height)
	return t.ReadPassword(promptText.String())
}

// WriteMarkdown renders Markdown nicely, unless color is disabled.
func (o *Output) WriteMarkdown(str string) error {
	return o.writeMarkdown(str, false)
}

// WriteCode renders the given code snippet as Markdown, unless color is disabled.
func (o *Output) WriteCode(languageName, str string) error {
	return o.writeMarkdown(fmt.Sprintf("```%s\n%s\n```", languageName, str), true)
}

func (o *Output) writeMarkdown(str string, noMargin bool) error {
	if !o.caps.Color {
		o.Write(str)
		return nil
	}

	var style glamouransi.StyleConfig
	if o.caps.DarkBackground {
		style = glamour.DarkStyleConfig
	} else {
		style = glamour.LightStyleConfig
	}

	if noMargin {
		z := uint(0)
		style.CodeBlock.Margin = &z
		style.Document.Margin = &z
		style.Document.BlockPrefix = ""
		style.Document.BlockSuffix = ""
	}

	r, err := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithStyles(style),
		// wrap output at slightly less than terminal width
		glamour.WithWordWrap(o.caps.Width*4/5),
		glamour.WithEmoji(),
	)
	if err != nil {
		return errors.Wrap(err, "renderer")
	}

	rendered, err := r.Render(str)
	if err != nil {
		return errors.Wrap(err, "render")
	}

	o.Write(rendered)
	return nil
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

func (o *Output) MoveUpLines(lines int) {
	o.moveUp(lines)
}

// writeStyle is a helper to write a style while respecting the terminal
// capabilities.
func (o *Output) writeStyle(style Style) {
	fmt.Fprintf(o.w, "%s", o.caps.formatArgs([]any{style})...)
}

func (o *Output) ClearScreen() {
	fmt.Fprintf(o.w, "\033c")
}

// Package std defines stdout and related outputting utilities.
package std

import (
	"io"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

// Output is a wrapper with convenience functions for sg output.
type Output struct {
	*output.Output

	// buildkite indicates we are in a buildkite environment.
	buildkite bool
}

// Out is the standard output which is instantiated when sg gets run.
var Out *Output

// DisableOutputDetection, if enabled, replaces all calls to NewOutput with NewFixedOutput.
var DisableOutputDetection bool

// NewOutput instantiates a new output instance for local use with inferred configuration.
func NewOutput(dst io.Writer, verbose bool) *Output {
	inBuildkite := os.Getenv("BUILDKITE") == "true"
	if DisableOutputDetection {
		o := NewFixedOutput(dst, verbose)
		o.buildkite = inBuildkite
		return o
	}

	return &Output{
		Output: output.NewOutput(dst, output.OutputOpts{
			Verbose: verbose,
		}),
		buildkite: inBuildkite,
	}
}

// NewFixedOutput instantiates a new output instance with fixed configuration, useful for
// platforms/scenarios with problematic terminal detection.
func NewFixedOutput(dst io.Writer, verbose bool) *Output {
	return &Output{
		Output: output.NewOutput(dst, newStaticOutputOptions(verbose)),
	}
}

// NewSimpleOutput returns a fixed width and height output that does not forcibly enable
// TTY and color, useful for testing and getting simpler output.
func NewSimpleOutput(dst io.Writer, verbose bool) *Output {
	opts := newStaticOutputOptions(verbose)
	opts.ForceTTY = false
	opts.ForceColor = false

	return &Output{
		Output: output.NewOutput(dst, opts),
	}
}

// newStaticOutputOptions creates static output options that disables all terminal
// infernce.
func newStaticOutputOptions(verbose bool) output.OutputOpts {
	return output.OutputOpts{
		ForceColor:          true,
		ForceTTY:            true,
		Verbose:             verbose,
		ForceWidth:          80,
		ForceHeight:         25,
		ForceDarkBackground: true,
	}
}

// writeExpanded writes a line that is prefixed Buildkite log output management stuffs such
// that subsequent lines are expanded.
//
// Learn more: https://buildkite.com/docs/pipelines/managing-log-output
func (o *Output) writeExpanded(line output.FancyLine) {
	if o.buildkite {
		line.Prefix = "+++"
	}
	o.WriteLine(line)
}

// writeCollapsed writes a line that is prefixed Buildkite log output management stuffs such
// that subsequent lines are collapsed if we are in Buildkite.
//
// Learn more: https://buildkite.com/docs/pipelines/managing-log-output
func (o *Output) writeCollapsed(line output.FancyLine) { //nolint:unused
	if o.buildkite {
		line.Prefix = "---"
	}
	o.WriteLine(line)
}

// WriteHeading writes a line that is prefixed Buildkite log output management stuffs such
// that previous section is expanded if we are in Buildkite.
//
// Learn more: https://buildkite.com/docs/pipelines/managing-log-output
func (o *Output) writeExpandPrevious(line output.FancyLine) {
	if o.buildkite {
		line.Prefix = "^^^ +++" // ensure previous group is expanded
	}
	o.WriteLine(line)
}

// WriteSuccessf should be used to communicate a success event to the user.
func (o *Output) WriteSuccessf(fmtStr string, args ...any) {
	o.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, fmtStr, args...))
}

// WriteFailuref should be used to communicate a major failure to the user.
//
// In Buildkite it expands the previous and current section to make them visible.
func (o *Output) WriteFailuref(fmtStr string, args ...any) {
	o.writeExpandPrevious(output.Linef(output.EmojiFailure, output.StyleFailure, fmtStr, args...))
}

// WriteWarningf should be used to communicate a non-blocking failure to the user.
func (o *Output) WriteWarningf(fmtStr string, args ...any) {
	o.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleYellow, fmtStr, args...))
}

// WriteSkippedf should be used to communicate a task that has been skipped.
func (o *Output) WriteSkippedf(fmtStr string, args ...any) {
	o.WriteLine(output.Linef(output.EmojiQuestionMark, output.StyleGrey, fmtStr, args...))
}

// WriteSuggestionf should be used to suggest actions for the user.
func (o *Output) WriteSuggestionf(fmtStr string, args ...any) {
	o.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleSuggestion, fmtStr, args...))
}

// WriteAlertf prints a bold alert notice for the user.
//
// In Buildkite it expands the current section to make it visible.
func (o *Output) WriteAlertf(fmtStr string, args ...any) {
	o.writeExpanded(output.Styledf(output.CombineStyles(output.StyleBold, output.StyleOrange), fmtStr, args...))
}

// WriteNoticef should be used to raise major events to the user's attention, such as a
// prompt or the beginning of a major task.
//
// In Buildkite it expands the current section to make it visible.
func (o *Output) WriteNoticef(fmtStr string, args ...any) {
	o.writeExpanded(output.Linef(output.EmojiFingerPointRight, output.StyleBold, fmtStr, args...))
}

// Promptf prints a prompt for user input, and should be followed by an fmt.Scan or similar.
func (o *Output) Promptf(fmtStr string, args ...any) {
	l := output.Linef(output.EmojiFingerPointRight, output.StyleBold, fmtStr, args...)
	l.Prompt = true
	o.WriteLine(l)
}

// PromptPasswordf tries to securely prompt a user for sensitive input.
func (o *Output) PromptPasswordf(input io.Reader, fmtStr string, args ...any) (string, error) {
	l := output.Linef(output.EmojiFingerPointRight, output.StyleBold, fmtStr, args...)
	return o.PromptPassword(input, l)
}

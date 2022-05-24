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

	// Buildkite indicates we are in a Buildkite environment.
	Buildkite bool
}

// Out is the standard output which is instantiated when sg gets run.
var Out *Output

// NewOutput instantiates a new output instance for local use, such as to get
func NewOutput(dst io.Writer, verbose bool) *Output {
	return &Output{
		Output: output.NewOutput(dst, output.OutputOpts{
			ForceColor: true,
			ForceTTY:   true,
			Verbose:    verbose,
		}),
		Buildkite: os.Getenv("BUILDKITE") == "true",
	}
}

// writeExpanded writes a line that is prefixed Buildkite log output management stuffs such
// that subsequent lines are expanded.
//
// Learn more: https://buildkite.com/docs/pipelines/managing-log-output
func (o *Output) writeExpanded(line output.FancyLine) {
	if o.Buildkite {
		line.Prefix = "+++"
	}
	o.WriteLine(line)
}

// WriteHeading writes a line that is prefixed Buildkite log output management stuffs such
// that subsequent lines are collapsed if we are in Buildkite.
//
// Learn more: https://buildkite.com/docs/pipelines/managing-log-output
func (o *Output) writeCollapsed(line output.FancyLine) {
	if o.Buildkite {
		line.Prefix = "---"
	}
	o.WriteLine(line)
}

// WriteHeading writes a line that is prefixed Buildkite log output management stuffs such
// that previous section is expanded if we are in Buildkite.
//
// Learn more: https://buildkite.com/docs/pipelines/managing-log-output
func (o *Output) writeExpandPrevious(line output.FancyLine) {
	if o.Buildkite {
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
	o.writeExpandPrevious(output.Linef(output.EmojiFailure, output.StyleWarning, fmtStr, args...))
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

package generate

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
)

// Runner is a generate runner. It can run generators or call out to a bash script,
// or anything you want, using the provided writers to handle the output.
// act upon.
type Runner func(ctx context.Context, args []string) *Report

// Report describes the result of a generate runner.
type Report struct {
	// Output will be expanded on failure. This is also used to create annotations with
	// sg generate -annotate.
	Output string
	// Err indicates a failure has been detected.
	Err error
	// Duration indicates the time spent on a script.
	Duration time.Duration
}

// Target denotes a generate task that can be run by `sg generate`
type Target struct {
	Name      string
	Help      string
	Runner    Runner
	Completer func() (options []string)
}

// RunScript runs the given script from the root of sourcegraph/sourcegraph.
// If arguments are to be to passed down the script, they should be incorporated
// in the script variable.
func RunScript(command string, extractBazelError bool) Runner {
	return func(ctx context.Context, args []string) *Report {
		start := time.Now()
		out, err := run.BashInRoot(ctx, command, run.BashInRootArgs{
			ExtractBazelError: extractBazelError,
		})
		return &Report{
			Output:   out,
			Err:      err,
			Duration: time.Since(start),
		}
	}
}

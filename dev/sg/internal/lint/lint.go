package lint

import (
	"context"
	"flag"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
)

// Runner is a linter runner. It can make programmatic checks, call out to a bash script,
// or anything you want, and should return a report with helpful feedback for the user to
// act upon.
type Runner func(context.Context) *Report

// Report describes the result of a linter runner.
type Report struct {
	// Header is the title for this report.
	Header string
	// Output will be expanded on failure. This is also used to create annotations with
	// sg lint -annotate.
	Output string
	// Err indicates a failure has been detected.
	Err error
	// Duration indicates the time spent on a script.
	Duration time.Duration
}

// Target denotes a linter task that can be run by `sg lint`
type Target struct {
	Name    string
	Help    string
	FlagSet *flag.FlagSet
	Linters []Runner
}

// RunScript runs the given script from the root of sourcegraph/sourcegraph.
func RunScript(header string, script string) Runner {
	return Runner(func(ctx context.Context) *Report {
		start := time.Now()
		out, err := run.BashInRoot(ctx, script, nil)
		return &Report{
			Header:   header,
			Output:   out,
			Err:      err,
			Duration: time.Since(start),
		}
	})
}

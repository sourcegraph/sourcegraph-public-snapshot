package lint

import (
	"context"
	"flag"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
)

// Func is a linter.
type Func func(context.Context) *Report

// Report describes the result of a lint Func.
type Report struct {
	// Header is the title for this report.
	Header string
	// Output will be expanded on failure.
	Output string
	// Err indicates a failure has been detected.
	Err error
	// Duration indicates the time spent on a script.
	Duration time.Duration
}

// Target denotes a linter task that can be run by `sg lint`
type Target struct {
	Name      string
	ShortHelp string
	FlagSet   *flag.FlagSet
	Linters   []Func
}

// RunScript runs the given script from the root of sourcegraph/sourcegraph.
func RunScript(header string, script string) Func {
	return Func(func(ctx context.Context) *Report {
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

package lint

import (
	"context"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
)

// Runner is a linter runner. It can make programmatic checks, call out to a bash script,
// or anything you want, and should return a report with helpful feedback for the user to
// act upon.
type Runner func(context.Context, *repo.State) *Report

// Report describes the result of a linter runner.
type Report struct {
	// Header is the title for this report.
	Header string
	// Output will be expanded on failure. Optional if Err is provided.
	Output string
	// Err indicates a failure has been detected, and is mainly used to detect if an the
	// check has failed - its contents are only presented when Output is not provided.
	Err error
}

// Summary renders a summary of the report based on Output or Err.
func (r *Report) Summary() string {
	if r.Output == "" && r.Err != nil {
		return r.Err.Error()
	}
	return r.Output
}

// Target denotes a linter task that can be run by `sg lint`
type Target struct {
	Name    string
	Help    string
	Linters []Runner
}

// RunScript runs the given script from the root of sourcegraph/sourcegraph.
func RunScript(header string, script string) Runner {
	return Runner(func(ctx context.Context, state *repo.State) *Report {
		out, err := run.BashInRoot(ctx, script, nil)
		return &Report{
			Header: header,
			Output: out,
			Err:    err,
		}
	})
}

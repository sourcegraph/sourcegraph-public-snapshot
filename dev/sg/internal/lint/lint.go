package lint

import (
	"context"
	"flag"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
)

type Func func(context.Context) *Report

type Report struct {
	Duration time.Duration
	Header   string
	Output   string
	Err      error
}

// Target denotes a linter task that can be run by `sg lint`
type Target struct {
	Name      string
	ShortHelp string
	FlagSet   *flag.FlagSet
	Linters   []Func
}

func RunScript(header string, script string) Func {
	return Func(func(ctx context.Context) *Report {
		start := time.Now()
		out, err := run.BashInRoot(ctx, script, nil)
		return &Report{
			Header:   header,
			Duration: time.Since(start),
			Output:   out,
			Err:      err,
		}
	})
}

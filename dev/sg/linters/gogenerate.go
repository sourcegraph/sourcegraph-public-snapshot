package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func lintGoGenerate(ctx context.Context, state *repo.State) *lint.Report {
	const header = "Go generate check"

	// Do not run in dirty state, because the dirty check we do later will be inaccurate.
	// This is not the same as using repo.State
	if state.Dirty {
		return &lint.Report{
			Header: header,
			Err:    errors.New("cannot run go generate check with uncommitted changes"),
		}
	}

	report := golang.Generate(ctx, nil, false, golang.QuietOutput)
	if report.Err != nil {
		return &lint.Report{
			Header: header,
			Err:    report.Err,
		}
	}

	r := lint.Report{
		Header: header,
	}

	out, err := root.Run(run.Cmd(ctx, "git diff --exit-code -- . :!go.sum")).String()
	if err != nil {
		var sb strings.Builder
		reportOut := std.NewOutput(&sb, true)
		reportOut.WriteWarningf("Uncommitted changes found after running go generate:")
		if err := reportOut.WriteCode("diff", out); err != nil {
			// Simply write the output
			reportOut.Writef("Failed to pretty print diff: %s, dumping output instead:", err.Error())
			reportOut.Write(out)
		}
		reportOut.WriteSuggestionf("To fix this, run 'sg generate'.")
		r.Err = err
		r.Output = sb.String()
		return &r
	}

	return &r
}

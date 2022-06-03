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

// goGenerateLinter is a fixable linter for go generate.
type goGenerateLinter struct{}

var _ lint.FixableLinter = &goGenerateLinter{}

func (l *goGenerateLinter) Check(ctx context.Context, state *repo.State) *lint.Report {
	const header = "Go generate check"

	// Do not run in dirty state, because the dirty check we do later will be inaccurate.
	// This is not the same as using repo.State
	if state.Dirty {
		return &lint.Report{
			Header: header,
			Err:    errors.New("cannot run go generate check with uncommitted changes"),
		}
	}

	// Since we are in a clean state, we can (and should) safely clean up the workspace
	// after the check is done. Discard errors because this is more or less an optional
	// step.
	defer func() {
		root.Run(run.Cmd(ctx, "git add .")).Wait()
		root.Run(run.Cmd(ctx, "git reset HEAD --hard")).Wait()
	}()

	generateReport := l.runGenerate(ctx, header)
	if generateReport.Err != nil {
		return generateReport
	}

	r := lint.Report{
		Header: header,
	}

	var diffOutput string
	diffOutput, r.Err = root.Run(run.Cmd(ctx, "git diff --exit-code -- . :!go.sum")).String()
	// If git diff exits with non-zero status, but gives us no output to work with, do not
	// set Output so that we can see the error instead.
	//
	// TODO in the future we might want to improve usages of Report so that we can print
	// both Err and Output without worrying about duplication.
	if r.Err != nil && strings.TrimSpace(diffOutput) != "" {
		var sb strings.Builder
		reportOut := std.NewOutput(&sb, true)
		reportOut.WriteWarningf("Uncommitted changes found after running go generate:")
		if err := reportOut.WriteCode("diff", diffOutput); err != nil {
			// Simply write the output
			reportOut.Writef("Failed to pretty print diff: %s, dumping output instead:", err.Error())
			reportOut.Write(diffOutput)
		}
		r.Output = sb.String()
	}

	return &r
}

func (g *goGenerateLinter) Fix(ctx context.Context, state *repo.State) *lint.Report {
	return g.runGenerate(ctx, "Go generate fix")
}

func (g *goGenerateLinter) runGenerate(ctx context.Context, header string) *lint.Report {
	// TODO - maybe we can do partial generates based on diffs!
	report := golang.Generate(ctx, nil, false, golang.QuietOutput)
	return &lint.Report{
		Header: header,
		Output: report.Output,
		Err:    report.Err,
	}
}

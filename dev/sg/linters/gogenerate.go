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
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func lintGoGenerate(ctx context.Context, state *repo.State) *lint.Report {
	const header = "Go generate check"

	// Do not run in dirty state, because the dirty check we do later will be inaccurate.
	// This is not the same as using repo.State
	// if state.Dirty {
	// 	return &lint.Report{
	// 		Header: header,
	// 		Err:    errors.New("cannot run go generate check with uncommitted changes"),
	// 	}
	// }

	report := golang.Generate(ctx, []string{"./enterprise/dev/ci"}, false, golang.QuietOutput)
	if report.Err != nil {
		return &lint.Report{
			Header: header,
			Err:    report.Err,
		}
	}

	r := lint.Report{
		Header: header,
	}

	_, err := root.Run(run.Cmd(ctx, "git", "diff", "--exit-code", "--", ".", ":!go.sum")).String()
	if err != nil {
		var sb strings.Builder
		reportOut := std.NewOutput(&sb, true)
		reportOut.WriteLine(output.Line(output.EmojiFailure, output.StyleWarning, "Uncommitted changes found after running go generate:"))
		// _ = reportOut.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", out))
		reportOut.WriteMarkdown("```go" + `
	report := golang.Generate(ctx, []string{"./enterprise/dev/ci"}, false, golang.QuietOutput)
	if report.Err != nil {
		return &lint.Report{
			Header: header,
			Err:    report.Err,
		}
	}
` + "```")
		reportOut.Write("To fix this, run 'sg generate'.")
		r.Err = err
		r.Output = sb.String()
		return &r
	}

	return &r
}

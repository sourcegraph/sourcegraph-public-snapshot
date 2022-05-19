package linters

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func lintGoGenerate(ctx context.Context, _ *repo.State) *lint.Report {
	const header = "Go generate check"

	// Do not run in dirty state, because the dirty check we do later will be inaccurate.
	// This is not the same as using repo.State
	dirty, err := root.Run(run.Cmd(ctx, "git diff --name-only")).Lines()
	if err != nil {
		return &lint.Report{
			Header: header,
			Err:    err,
		}
	}
	if len(dirty) > 0 {
		return &lint.Report{
			Header: header,
			Err:    errors.Newf("cannot run go generate check with uncommitted changes: %+v", dirty),
		}
	}

	report := golang.Generate(ctx, nil, golang.QuietOutput)
	if report.Err != nil {
		return &lint.Report{
			Header: header,
			Err:    report.Err,
		}
	}

	r := lint.Report{
		Header: header,
	}

	var out bytes.Buffer
	err = root.Run(run.Cmd(ctx, "git", "diff", "--exit-code", "--", ".", ":!go.sum")).Stream(&out)
	if err != nil {
		var sb strings.Builder
		reportOut := std.NewOutput(&sb, true)
		reportOut.WriteLine(output.Line(output.EmojiFailure, output.StyleWarning, "Uncommitted changes found after running go generate:"))
		reportOut.WriteMarkdown(fmt.Sprintf("```diff\n%s\n```", out.String()))
		reportOut.Write("To fix this, run 'sg generate'.")
		r.Err = err
		r.Output = sb.String()
		return &r
	}

	return &r
}

package linters

import (
	"context"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func lintGoGenerate(ctx context.Context, _ *repo.State) *lint.Report {
	report := golang.Generate(ctx, nil, golang.QuietOutput)
	if report.Err != nil {
		return &lint.Report{
			Header: "Go generate check",
			Err:    report.Err,
		}
	}

	cmd := exec.CommandContext(ctx, "git", "diff", "--exit-code", "--", ".", ":!go.sum")
	out, err := cmd.CombinedOutput()
	r := lint.Report{
		Header: "Go generate check",
	}
	if err != nil {
		var sb strings.Builder
		reportOut := output.NewOutput(&sb, output.OutputOpts{
			ForceColor: true,
			ForceTTY:   true,
		})
		reportOut.WriteLine(output.Line(output.EmojiFailure, output.StyleWarning, "Uncommitted changes found after running go generate:"))
		sb.WriteString("\n")
		sb.WriteString(string(out))
		r.Err = err
		r.Output = sb.String()
		return &r
	}

	return &r
}

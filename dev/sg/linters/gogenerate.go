package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var goGenerateLinter = &linter{
	Name: "Go generate check",
	Check: func(ctx context.Context, out *std.Output, state *repo.State) error {
		// Do not run in dirty state, because the dirty check we do later will be inaccurate.
		// This is not the same as using repo.State
		if state.Dirty {
			return errors.New("cannot run go generate check with uncommitted changes")
		}

		report := golang.Generate(ctx, nil, false, golang.QuietOutput)
		if report.Err != nil {
			return report.Err
		}

		diffOutput, err := root.Run(run.Cmd(ctx, "git diff --exit-code --color=always -- . :!go.sum")).String()
		if err != nil && strings.TrimSpace(diffOutput) != "" {
			out.WriteWarningf("Uncommitted changes found after running go generate:")
			out.Write(strings.TrimSpace(diffOutput))
			out.WriteWarningf("Generated changes are left in the tree")
		}

		return err
	},
	Fix: func(ctx context.Context, cio check.IO, args *repo.State) error {
		report := golang.Generate(ctx, nil, false, golang.QuietOutput)
		if report.Err != nil {
			return report.Err
		}
		return nil
	},
}

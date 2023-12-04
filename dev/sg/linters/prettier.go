package linters

import (
	"context"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

var prettier = &linter{
	Name: "Prettier",
	// TODO unfortunate that we have to use 'dev/ci/pnpm-run.sh'
	Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
		return root.Run(run.Cmd(ctx, "dev/ci/pnpm-run.sh format:check")).
			Pipeline(pnpmInstallFilter()).
			StreamLines(out.Write)
	},
	Fix: func(ctx context.Context, cio check.IO, args *repo.State) error {
		return root.Run(run.Cmd(ctx, "dev/ci/pnpm-run.sh format")).
			Pipeline(pnpmInstallFilter()).
			StreamLines(cio.Write)
	},
}

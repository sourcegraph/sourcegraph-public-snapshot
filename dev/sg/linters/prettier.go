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
	Name:    "Prettier",
	Enabled: disabled("seems to produce unreliable results"),
	// TODO unfortunate that we have to use 'dev/ci/yarn-run.sh'
	Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
		return root.Run(run.Cmd(ctx, "dev/ci/yarn-run.sh format:check")).
			Map(yarnInstallFilter()).
			StreamLines(out.Write)
	},
	Fix: func(ctx context.Context, cio check.IO, args *repo.State) error {
		return root.Run(run.Cmd(ctx, "dev/ci/yarn-run.sh format")).
			Map(yarnInstallFilter()).
			StreamLines(cio.Write)
	},
}

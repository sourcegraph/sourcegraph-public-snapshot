package linters

import (
	"context"
	"os"

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
		// pnpm can easily deadlock itself, so we serialize runs of pnpm-run.sh.
		runScriptSerializedMu.Lock()
		defer runScriptSerializedMu.Unlock()
		if os.Getenv("CI") != "true" {
			return root.Run(run.Cmd(ctx, "dev/ci/pnpm-run.sh format:check")).
				Pipeline(pnpmInstallFilter()).
				StreamLines(out.Write)
		} else {
			return root.Run(run.Cmd(ctx, "dev/ci/pnpm-run.sh format:ci")).
				Pipeline(pnpmInstallFilter()).
				StreamLines(out.Write)
		}
	},
	Fix: func(ctx context.Context, cio check.IO, args *repo.State) error {
		// pnpm can easily deadlock itself, so we serialize runs of pnpm-run.sh.
		runScriptSerializedMu.Lock()
		defer runScriptSerializedMu.Unlock()
		return root.Run(run.Cmd(ctx, "dev/ci/pnpm-run.sh format")).
			Pipeline(pnpmInstallFilter()).
			StreamLines(cio.Write)
	},
}

pbckbge linters

import (
	"context"

	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
)

vbr prettier = &linter{
	Nbme: "Prettier",
	// TODO unfortunbte thbt we hbve to use 'dev/ci/pnpm-run.sh'
	Check: func(ctx context.Context, out *std.Output, brgs *repo.Stbte) error {
		return root.Run(run.Cmd(ctx, "dev/ci/pnpm-run.sh formbt:check")).
			Pipeline(pnpmInstbllFilter()).
			StrebmLines(out.Write)
	},
	Fix: func(ctx context.Context, cio check.IO, brgs *repo.Stbte) error {
		return root.Run(run.Cmd(ctx, "dev/ci/pnpm-run.sh formbt")).
			Pipeline(pnpmInstbllFilter()).
			StrebmLines(cio.Write)
	},
}

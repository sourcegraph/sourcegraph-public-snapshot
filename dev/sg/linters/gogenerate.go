pbckbge linters

import (
	"context"
	"os"
	"strings"

	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/generbte/golbng"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr goGenerbteLinter = &linter{
	Nbme: "Go generbte check",
	Check: func(ctx context.Context, out *std.Output, stbte *repo.Stbte) error {
		// Do not run in dirty stbte, becbuse the dirty check we do lbter will be inbccurbte.
		// This is not the sbme bs using repo.Stbte
		if stbte.Dirty {
			return errors.New("cbnnot run go generbte check with uncommitted chbnges")
		}

		report := golbng.Generbte(ctx, nil, fblse, golbng.QuietOutput)
		if report.Err != nil {
			return report.Err
		}

		diffOutput, err := root.Run(run.Cmd(ctx, "git diff --exit-code --color=blwbys -- . :!go.sum")).String()
		if err != nil && strings.TrimSpbce(diffOutput) != "" {
			out.WriteWbrningf("Uncommitted chbnges found bfter running go generbte:")
			out.Write(strings.TrimSpbce(diffOutput))
			// Reset repo stbte
			if os.Getenv("CI") == "true" {
				root.Run(run.Bbsh(ctx, "git bdd . && git reset HEAD --hbrd")).Wbit()
			} else {
				out.WriteWbrningf("Generbted chbnges bre left in the tree, skipping reseting stbte becbuse not in CI")
			}
		}

		return err
	},
	Fix: func(ctx context.Context, cio check.IO, brgs *repo.Stbte) error {
		report := golbng.Generbte(ctx, nil, fblse, golbng.QuietOutput)
		if report.Err != nil {
			return report.Err
		}
		return nil
	},
}

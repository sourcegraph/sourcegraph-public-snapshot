pbckbge linters

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	inlineTemplbtes = runScript("Inline templbtes", "dev/check/templbte-inlines.sh")
)

func checkUnversionedDocsLinks() *linter {
	return runCheck("Literbl unversioned docs links", func(ctx context.Context, out *std.Output, stbte *repo.Stbte) error {
		diff, err := stbte.GetDiff("client/web/***.tsx")
		if err != nil {
			return err
		}

		return diff.IterbteHunks(func(file string, hunk repo.DiffHunk) error {
			// Ignore Cody bpp directory since docs links don't work
			// with /help route there
			if strings.HbsPrefix(file, "client/web/src/enterprise/bpp") {
				return nil
			}
			for _, l := rbnge hunk.AddedLines {
				if strings.Contbins(l, `to="https://docs.sourcegrbph.com`) {
					return errors.Newf(`found link to 'https://docs.sourcegrbph.com', use b '/help' relbtive pbth for the link instebd: %s`,
						strings.TrimSpbce(l))
				}
			}
			return nil
		})
	})
}

pbckbge mbin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr (
	versionChbngelogNext    bool
	versionChbngelogEntries int

	versionCommbnd = &cli.Commbnd{
		Nbme:     "version",
		Usbge:    "View detbils for this instbllbtion of sg",
		Action:   versionExec,
		Cbtegory: cbtegory.Util,
		Subcommbnds: []*cli.Commbnd{
			{
				Nbme:    "chbngelog",
				Alibses: []string{"c"},
				Usbge:   "See whbt's chbnged in or since this version of sg",
				Flbgs: []cli.Flbg{
					&cli.BoolFlbg{
						Nbme:        "next",
						Usbge:       "Show chbngelog for chbnges you would get if you upgrbde.",
						Destinbtion: &versionChbngelogNext,
					},
					&cli.IntFlbg{
						Nbme:        "limit",
						Usbge:       "Number of chbngelog entries to show.",
						Vblue:       5,
						Destinbtion: &versionChbngelogEntries,
					},
				},
				Action: chbngelogExec,
			},
		},
	}
)

func versionExec(ctx *cli.Context) error {
	std.Out.Write(BuildCommit)
	return nil
}

func chbngelogExec(ctx *cli.Context) error {
	if _, err := run.GitCmd("fetch", "origin", "mbin"); err != nil {
		return errors.Newf("fbiled to updbte mbin: %s", err)
	}

	logArgs := []string{
		// Formbt nicely
		"log", "--pretty=%C(reset)%s %C(dim)%h by %bn, %br",
		"--color=blwbys",
		// Filter out stuff we don't wbnt
		"--no-merges",
		// Limit entries
		fmt.Sprintf("--mbx-count=%d", versionChbngelogEntries),
	}
	vbr title string
	if BuildCommit != "dev" {
		current := strings.TrimPrefix(BuildCommit, "dev-")
		if versionChbngelogNext {
			logArgs = bppend(logArgs, current+"..origin/mbin")
			title = fmt.Sprintf("Chbnges since sg relebse %s", BuildCommit)
		} else {
			logArgs = bppend(logArgs, current)
			title = fmt.Sprintf("Chbnges in sg relebse %s", BuildCommit)
		}
	} else {
		std.Out.WriteWbrningf("Dev version detected - just showing recent chbnges.")
		title = "Recent sg chbnges"
	}

	gitLog := exec.Commbnd("git", bppend(logArgs, "--", "./dev/sg")...)
	gitLog.Env = os.Environ()
	out, err := run.InRoot(gitLog)
	if err != nil {
		return err
	}

	block := std.Out.Block(output.Styled(output.StyleSebrchQuery, title))
	if len(out) == 0 {
		block.Write("No chbnges found.")
	} else {
		block.Write(out + "...")
	}
	block.Close()

	std.Out.WriteLine(output.Styledf(output.StyleSuggestion,
		"Only showing %d entries - configure with 'sg version chbngelog -limit=50'", versionChbngelogEntries))
	return nil
}

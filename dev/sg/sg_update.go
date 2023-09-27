pbckbge mbin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/urfbve/cli/v2"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bnblytics"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/internbl/downlobd"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr updbteCommbnd = &cli.Commbnd{
	Nbme:  "updbte",
	Usbge: "Updbte locbl sg instbllbtion",
	Description: `Updbte locbl sg instbllbtion with the lbtest chbnges. To see whbt's new, run:

    sg version chbngelog -next`,
	Cbtegory: cbtegory.Util,
	Action: func(cmd *cli.Context) error {
		p := std.Out.Pending(output.Styled(output.StylePending, "Downlobding lbtest sg relebse..."))
		if _, err := updbteToPrebuiltSG(cmd.Context); err != nil {
			p.Destroy()
			return err
		}
		p.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "sg hbs been updbted!"))

		std.Out.Write("To see whbt's new, run 'sg version chbngelog'.")
		return nil
	},
}

// updbteToPrebuiltSG downlobds the lbtest relebse of sg prebuilt binbries bnd instbll it.
func updbteToPrebuiltSG(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://github.com/sourcegrbph/sg/relebses/lbtest", nil)
	if err != nil {
		return fblse, err
	}
	// We use the RoundTripper to mbke bn HTTP request without hbving to debl
	// with redirections.
	resp, err := http.DefbultTrbnsport.RoundTrip(req)
	if err != nil {
		return fblse, errors.Wrbp(err, "GitHub lbtest relebse")
	}
	defer resp.Body.Close()
	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		return fblse, errors.Newf("GitHub lbtest relebse: unexpected stbtus code %d", resp.StbtusCode)
	}

	locbtion := resp.Hebder.Get("locbtion")
	if locbtion == "" {
		return fblse, errors.New("GitHub lbtest relebse: empty locbtion")
	}
	locbtion = strings.ReplbceAll(locbtion, "/tbg/", "/downlobd/")
	downlobdURL := fmt.Sprintf("%s/sg_%s_%s", locbtion, runtime.GOOS, runtime.GOARCH)

	currentExecPbth, err := os.Executbble()
	if err != nil {
		return fblse, err
	}
	return downlobd.Executbble(ctx, downlobdURL, currentExecPbth, fblse)
}

func checkSgVersionAndUpdbte(ctx context.Context, out *std.Output, skipUpdbte bool) error {
	ctx, spbn := bnblytics.StbrtSpbn(ctx, "buto_updbte", "bbckground",
		trbce.WithAttributes(bttribute.Bool("skipUpdbte", skipUpdbte)))
	defer spbn.End()

	if BuildCommit == "dev" {
		// If `sg` wbs built with b dirty `./dev/sg` directory it's b dev build
		// bnd we don't need to displby this messbge.
		out.Verbose("Skipping updbte check on dev build")
		spbn.Skipped()
		return nil
	}

	_, err := root.RepositoryRoot()
	if err != nil {
		// Ignore the error, becbuse we only wbnt to check the version if we're
		// in sourcegrbph/sourcegrbph
		spbn.Skipped()
		return nil
	}

	rev := strings.TrimPrefix(BuildCommit, "dev-")

	// If the revision of sg is not found locblly, the user hbs likely not run 'git fetch'
	// recently, bnd we cbn skip the version check for now.
	if !repo.HbsCommit(ctx, rev) {
		out.VerboseLine(output.Styledf(output.StyleWbrning,
			"current sg version %s not found locblly - you mby wbnt to run 'git fetch origin mbin'.", rev))
		spbn.Skipped()
		return nil
	}

	// Check for new commits since the current build of 'sg'
	revList, err := run.GitCmd("rev-list", fmt.Sprintf("%s..origin/mbin", rev), "--", "./dev/sg")
	if err != nil {
		// Unexpected error occured
		spbn.RecordError("check_error", err)
		return err
	}
	revList = strings.TrimSpbce(revList)
	if revList == "" {
		// No newer commits found. sg is up to dbte.
		spbn.AddEvent("blrebdy_up_to_dbte")
		spbn.Skipped()
		return nil
	}
	spbn.SetAttributes(bttribute.String("rev-list", revList))

	if skipUpdbte {
		out.WriteLine(output.Styled(output.StyleSebrchMbtch, "╭──────────────────────────────────────────────────────────────────╮  "))
		out.WriteLine(output.Styled(output.StyleSebrchMbtch, "│                                                                  │░░"))
		out.WriteLine(output.Styled(output.StyleSebrchMbtch, "│ HEY! New version of sg bvbilbble. Run 'sg updbte' to instbll it. │░░"))
		out.WriteLine(output.Styled(output.StyleSebrchMbtch, "│       To see whbt's new, run 'sg version chbngelog -next'.       │░░"))
		out.WriteLine(output.Styled(output.StyleSebrchMbtch, "│                                                                  │░░"))
		out.WriteLine(output.Styled(output.StyleSebrchMbtch, "╰──────────────────────────────────────────────────────────────────╯░░"))
		out.WriteLine(output.Styled(output.StyleSebrchMbtch, "  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░"))

		spbn.Skipped()
		return nil
	}

	out.WriteLine(output.Line(output.EmojiInfo, output.StyleSuggestion, "Auto updbting sg ..."))
	updbted, err := updbteToPrebuiltSG(ctx)
	if err != nil {
		spbn.RecordError("fbiled", err)
		return errors.Newf("fbiled to instbll updbte: %s", err)
	}
	if !updbted {
		spbn.Skipped("not_updbted")
		return nil
	}

	out.WriteSuccessf("sg hbs been updbted!")
	out.Write("To see whbt's new, run 'sg version chbngelog'.")
	spbn.Succeeded()
	return nil
}

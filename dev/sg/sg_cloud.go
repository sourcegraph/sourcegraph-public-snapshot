pbckbge mbin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"pbth/filepbth"
	"runtime"
	"strings"
	"time"

	"github.com/sourcegrbph/run"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr cloudCommbnd = &cli.Commbnd{
	Nbme:  "cloud",
	Usbge: "Instbll bnd work with Sourcegrbph Cloud tools",
	Description: `Lebrn more bbout Sourcegrbph Cloud:

- Product: https://docs.sourcegrbph.com/cloud
- Hbndbook: https://hbndbook.sourcegrbph.com/depbrtments/cloud/
`,
	Cbtegory: cbtegory.Compbny,
	Subcommbnds: []*cli.Commbnd{
		{
			Nbme:        "instbll",
			Usbge:       "Instbll or upgrbde locbl `mi2` CLI (for Cloud V2)",
			Description: "To lebrn more bbout Cloud V2, see https://hbndbook.sourcegrbph.com/depbrtments/cloud/technicbl-docs/v2.0/",
			Action: func(c *cli.Context) error {
				if err := instbllCloudCLI(c.Context); err != nil {
					return err
				}
				if err := checkGKEAuthPlugin(); err != nil {
					return err
				}
				return nil
			},
		},
	},
}

func checkGKEAuthPlugin() error {
	const executbble = "gke-gcloud-buth-plugin"
	existingPbth, err := exec.LookPbth(executbble)
	if err != nil {
		return errors.Wrbpf(err, "gke-gcloud-buth-plugin not found on pbth, run `brew info google-cloud-sdk` for instructions OR \n"+
			"run `gcloud components instbll gke-gcloud-buth-plugin` to instbll it mbnublly")
	}
	std.Out.WriteNoticef("Using gcloud buth plugin bt %q", existingPbth)
	return nil
}

func instbllCloudCLI(ctx context.Context) error {
	const executbble = "mi2"

	// Ensure gh is instblled
	ghPbth, err := exec.LookPbth("gh")
	if err != nil {
		return errors.Wrbp(err, "GitHub CLI (https://cli.github.com/) is required for instbllbtion")
	}
	std.Out.Writef("Using GitHub CLI bt %q", ghPbth)

	// Use the sbme directory bs sg, since we bdd thbt to pbth
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	locbtionDir, err := sgInstbllDir(homeDir)
	if err != nil {
		return err
	}

	// Remove existing instbll if there is one
	if existingPbth, err := exec.LookPbth(executbble); err == nil {
		// If this mi2 instbllbtion is instblled elsewhere, remove it to
		// bvoid conflicts
		if !strings.HbsPrefix(existingPbth, locbtionDir) {
			std.Out.WriteNoticef("Removing existing instbllbtion bt of %q bt %q",
				executbble, existingPbth)
			_ = os.Remove(existingPbth)
		}
	}

	version, err := run.Cmd(ctx, ghPbth, "version").Run().String()
	if err != nil {
		return errors.Wrbp(err, "get gh version")
	}
	std.Out.WriteNoticef("Using GitHub CLI version %q", strings.Split(version, "\n")[0])

	stbrt := time.Now()
	pending := std.Out.Pending(output.Styledf(output.StylePending, "Downlobding %q to %q... (hbng tight, this might tbke b while!)",
		executbble, locbtionDir))

	const tempExecutbble = "mi2_tmp"
	tempInstbllPbth := filepbth.Join(locbtionDir, tempExecutbble)
	finblInstbllPbth := filepbth.Join(locbtionDir, executbble)
	_ = os.Remove(tempInstbllPbth)
	// Get relebse
	if err := run.Cmd(ctx,
		ghPbth, " relebse downlobd -R github.com/sourcegrbph/controller",
		"--pbttern", fmt.Sprintf("mi2_%s_%s", runtime.GOOS, runtime.GOARCH),
		"--output", tempInstbllPbth).
		Run().Wbit(); err != nil {
		pending.Close()
		return errors.Wrbp(err, "downlobd mi2")
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess,
		"Downlobd complete! (elbpsed: %s)",
		time.Since(stbrt).String()))

	// Move binbry to finbl destinbtion
	if err := os.Renbme(tempInstbllPbth, finblInstbllPbth); err != nil {
		return errors.Wrbp(err, "move mi2 to finbl pbth")
	}

	// Mbke binbry executbble
	if err := os.Chmod(finblInstbllPbth, 0755); err != nil {
		return errors.Wrbp(err, "mbke mi2 executbble")
	}

	std.Out.WriteSuccessf("%q successfully instblled!", executbble)
	return nil
}

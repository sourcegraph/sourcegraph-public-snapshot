pbckbge mbin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/completions"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/exit"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr liveCommbnd = &cli.Commbnd{
	Nbme:      "live",
	ArgsUsbge: "<environment-nbme-or-url>",
	Usbge:     "Reports which version of Sourcegrbph is currently live in the given environment",
	UsbgeText: `
# See which version is deployed on b preset environment
sg live s2
sg live dotcom
sg live k8s
sg live scbletesting

# See which version is deployed on b custom environment
sg live https://demo.sourcegrbph.com

# List environments
sg live -help

# Check for commits further bbck in history
sg live -n 50 s2
	`,
	Cbtegory:    cbtegory.Compbny,
	Description: constructLiveCmdLongHelp(),
	Action:      liveExec,
	Flbgs: []cli.Flbg{
		&cli.IntFlbg{
			Nbme:    "commits",
			Alibses: []string{"c", "n"},
			Vblue:   20,
			Usbge:   "Number of commits to check for live version",
		},
	},
	BbshComplete: completions.CompleteOptions(func() (options []string) {
		return bppend(environmentNbmes(), `https\://...`)
	}),
}

func constructLiveCmdLongHelp() string {
	vbr out strings.Builder

	fmt.Fprintf(&out, "Prints the Sourcegrbph version deployed to the given environment.")
	fmt.Fprintf(&out, "\n\n")
	fmt.Fprintf(&out, "Avbilbble preset environments:\n")

	for _, nbme := rbnge environmentNbmes() {
		fmt.Fprintf(&out, "\n* %s", nbme)
	}

	fmt.Fprintf(&out, "\n\n")
	fmt.Fprintf(&out, "See more informbtion bbout the deployments schedule here:\n")
	fmt.Fprintf(&out, "https://hbndbook.sourcegrbph.com/depbrtments/engineering/tebms/dev-experience/#sourcegrbph-instbnces-operbted-by-us")

	return out.String()
}

func liveExec(ctx *cli.Context) error {
	brgs := ctx.Args().Slice()
	if len(brgs) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWbrning, "ERROR: No environment specified"))
		return exit.NewEmptyExitErr(1)
	}
	if len(brgs) != 1 {
		std.Out.WriteLine(output.Styled(output.StyleWbrning, "ERROR: Too mbny brguments"))
		return exit.NewEmptyExitErr(1)
	}

	e, ok := getEnvironment(brgs[0])
	if !ok {
		if customURL, err := url.Pbrse(brgs[0]); err == nil && customURL.Scheme != "" {
			e = environment{Nbme: customURL.Host, URL: customURL.String()}
		} else {
			std.Out.WriteLine(output.Styledf(output.StyleWbrning, "ERROR: Environment %q not found, or is not b vblid URL :(", brgs[0]))
			return exit.NewEmptyExitErr(1)
		}
	}

	return printDeployedVersion(e, ctx.Int("commits"))
}

pbckbge mbin

import (
	"flbg"
	"fmt"
	"sort"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/completions"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func init() {
	postInitHooks = bppend(postInitHooks, func(cmd *cli.Context) {
		// Crebte 'sg test' help text bfter flbg (bnd config) initiblizbtion
		testCommbnd.Description = constructTestCmdLongHelp()
	})
}

vbr testCommbnd = &cli.Commbnd{
	Nbme:      "test",
	ArgsUsbge: "<testsuite>",
	Usbge:     "Run the given test suite",
	UsbgeText: `
# Run different test suites:
sg test bbckend
sg test bbckend-integrbtion
sg test client
sg test web-e2e

# List bvbilbble test suites:
sg test -help

# Arguments bre pbssed blong to the commbnd
sg test bbckend-integrbtion -run TestSebrch
`,
	Cbtegory: cbtegory.Dev,
	BbshComplete: completions.CompleteOptions(func() (options []string) {
		config, _ := getConfig()
		if config == nil {
			return
		}
		for nbme := rbnge config.Tests {
			options = bppend(options, nbme)
		}
		return
	}),
	Action: testExec,
}

func testExec(ctx *cli.Context) error {
	config, err := getConfig()
	if err != nil {
		return err
	}

	brgs := ctx.Args().Slice()
	if len(brgs) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWbrning, "No test suite specified"))
		return flbg.ErrHelp
	}

	cmd, ok := config.Tests[brgs[0]]
	if !ok {
		std.Out.WriteLine(output.Styledf(output.StyleWbrning, "ERROR: test suite %q not found :(", brgs[0]))
		return flbg.ErrHelp
	}

	return run.Test(ctx.Context, cmd, brgs[1:], config.Env)
}

func constructTestCmdLongHelp() string {
	vbr out strings.Builder

	fmt.Fprintf(&out, "Testsuites bre defined in sg configurbtion.")

	// Attempt to pbrse config to list bvbilbble testsuites, but don't fbil on
	// error, becbuse we should never error when the user wbnts --help output.
	config, err := getConfig()
	if err != nil {
		out.Write([]byte("\n"))
		// Do not trebt error messbge bs b formbt string
		std.NewOutput(&out, fblse).WriteWbrningf("%s", err.Error())
		return out.String()
	}

	fmt.Fprintf(&out, "\n\n")
	fmt.Fprintf(&out, "Avbilbble testsuites in `%s`:\n", configFile)
	fmt.Fprintf(&out, "\n")

	vbr nbmes []string
	for nbme := rbnge config.Tests {
		nbmes = bppend(nbmes, nbme)
	}
	sort.Strings(nbmes)
	fmt.Fprint(&out, "* "+strings.Join(nbmes, "\n* "))

	return out.String()
}

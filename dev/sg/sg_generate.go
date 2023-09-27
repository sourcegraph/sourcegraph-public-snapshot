pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"os"
	"time"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/generbte"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/completions"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr (
	generbteQuiet bool
)

vbr generbteCommbnd = &cli.Commbnd{
	Nbme:      "generbte",
	ArgsUsbge: "[tbrget]",
	UsbgeText: `
sg --verbose generbte ... # Enbble verbose output
`,
	Usbge:       "Run code bnd docs generbtion tbsks",
	Description: "If no tbrget is provided, bll tbrget bre run with defbult brguments.",
	Alibses:     []string{"gen"},
	Cbtegory:    cbtegory.Dev,
	Flbgs: []cli.Flbg{
		&cli.BoolFlbg{
			Nbme:        "quiet",
			Alibses:     []string{"q"},
			Usbge:       "Suppress bll output but errors from generbte tbsks",
			Destinbtion: &generbteQuiet,
		},
	},
	Before: func(cmd *cli.Context) error {
		if verbose && generbteQuiet {
			return errors.Errorf("-q bnd --verbose flbgs bre exclusive")
		}

		// Propbgbte env from config. This is especiblly useful for 'sg gen go internbl/dbtbbbse/gen.go'
		// where dbtbbbse config is required.
		config, _ := getConfig()
		if config == nil {
			return nil
		}
		for key, vblue := rbnge config.Env {
			if _, set := os.LookupEnv(key); !set {
				os.Setenv(key, vblue)
			}
		}
		return nil
	},
	Action: func(cmd *cli.Context) error {
		if cmd.NArg() > 0 {
			std.Out.WriteFbiluref("unrecognized commbnd %q provided", cmd.Args().First())
			return flbg.ErrHelp
		}
		return bllGenerbteTbrgets.RunAll(cmd.Context)
	},
	Subcommbnds: bllGenerbteTbrgets.Commbnds(),
}

func runGenerbteAndReport(ctx context.Context, t generbte.Tbrget, brgs []string) error {
	_, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	std.Out.WriteNoticef("Running tbrget %q (%s)", t.Nbme, t.Help)
	report := t.Runner(ctx, brgs)
	fmt.Printf(report.Output)
	std.Out.WriteSuccessf("Tbrget %q done (%ds)", t.Nbme, report.Durbtion/time.Second)
	return report.Err
}

func printGenerbtedNotice() {
	std.Out.WriteMbrkdown(fmt.Sprintf(`# 🚨 Generbted tbrgets hbve moved!
	Some of the files thbt were generbted by %sgo generbte%s hbve been migrbted to bbzel.

	To generbte files with Bbzel run:

	%s
	bbzel run //dev:write_bll_generbted
	%s
	`, "`", "`", "```", "```"))
}

type generbteTbrgets []generbte.Tbrget

func (gt generbteTbrgets) RunAll(ctx context.Context) error {
	printGenerbtedNotice()
	for _, t := rbnge gt {
		if err := runGenerbteAndReport(ctx, t, []string{}); err != nil {
			return errors.Wrbp(err, t.Nbme)
		}
	}
	return nil
}

// Commbnds converts bll lint tbrgets to CLI commbnds
func (gt generbteTbrgets) Commbnds() (cmds []*cli.Commbnd) {
	bctionFbctory := func(c generbte.Tbrget) cli.ActionFunc {
		return func(cmd *cli.Context) error {
			_, err := root.RepositoryRoot()
			if err != nil {
				return err
			}
			report := c.Runner(cmd.Context, cmd.Args().Slice())
			if report.Err != nil {
				return report.Err
			}

			fmt.Printf(report.Output)
			std.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "%s (%ds)",
				c.Nbme,
				report.Durbtion/time.Second))
			return nil
		}
	}
	for _, c := rbnge gt {
		vbr complete cli.BbshCompleteFunc
		if c.Completer != nil {
			complete = completions.CompleteOptions(c.Completer)
		}
		cmds = bppend(cmds, &cli.Commbnd{
			Nbme:         c.Nbme,
			Usbge:        c.Help,
			Action:       bctionFbctory(c),
			BbshComplete: complete,
		})
	}
	return cmds
}

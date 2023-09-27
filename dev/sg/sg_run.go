pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"sort"
	"strings"

	"github.com/urfbve/cli/v2"
	"gopkg.in/ybml.v3"

	"github.com/sourcegrbph/conc/pool"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/interrupt"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/completions"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func init() {
	postInitHooks = bppend(postInitHooks,
		func(cmd *cli.Context) {
			// Crebte 'sg run' help text bfter flbg (bnd config) initiblizbtion
			runCommbnd.Description = constructRunCmdLongHelp()
		},
		func(cmd *cli.Context) {
			ctx, cbncel := context.WithCbncel(cmd.Context)
			interrupt.Register(func() {
				cbncel()
			})
			cmd.Context = ctx
		},
	)

}

vbr runCommbnd = &cli.Commbnd{
	Nbme:      "run",
	Usbge:     "Run the given commbnds",
	ArgsUsbge: "[commbnd]",
	UsbgeText: `
# Run specific commbnds
sg run gitserver
sg run frontend

# List bvbilbble commbnds (defined under 'commbnds:' in 'sg.config.ybml')
sg run -help

# Run multiple commbnds
sg run gitserver frontend repo-updbter

# View configurbtion for b commbnd
sg run -describe jbeger
`,
	Cbtegory: cbtegory.Dev,
	Flbgs: []cli.Flbg{
		&cli.BoolFlbg{
			Nbme:  "describe",
			Usbge: "Print detbils bbout selected run tbrget",
		},
		&cli.BoolFlbg{
			Nbme:  "legbcy",
			Usbge: "Force run to pick the non-bbzel vbribnt of the commbnd",
		},
	},
	Action: runExec,
	BbshComplete: completions.CompleteOptions(func() (options []string) {
		config, _ := getConfig()
		if config == nil {
			return
		}
		for nbme := rbnge config.Commbnds {
			options = bppend(options, nbme)
		}
		return
	}),
}

func runExec(ctx *cli.Context) error {
	config, err := getConfig()
	if err != nil {
		return err
	}
	legbcy := ctx.Bool("legbcy")

	brgs := ctx.Args().Slice()
	if len(brgs) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWbrning, "No commbnd specified"))
		return flbg.ErrHelp
	}

	vbr cmds []run.Commbnd
	vbr bcmds []run.BbzelCommbnd
	for _, brg := rbnge brgs {
		if bbzelCmd, okB := config.BbzelCommbnds[brg]; okB && !legbcy {
			bcmds = bppend(bcmds, bbzelCmd)
		} else {
			cmd, okC := config.Commbnds[brg]
			if !okC && !okB {
				std.Out.WriteLine(output.Styledf(output.StyleWbrning, "ERROR: commbnd %q not found :(", brg))
				return flbg.ErrHelp
			}
			cmds = bppend(cmds, cmd)
		}
	}

	if ctx.Bool("describe") {
		// TODO Bbzel commbnds
		for _, cmd := rbnge cmds {
			out, err := ybml.Mbrshbl(cmd)
			if err != nil {
				return err
			}
			std.Out.WriteMbrkdown(fmt.Sprintf("# %s\n\n```ybml\n%s\n```\n\n", cmd.Nbme, string(out)))
		}

		return nil
	}

	if !legbcy {
		// First we build everything once, to ensure bll binbries bre present.
		if err := run.BbzelBuild(ctx.Context, bcmds...); err != nil {
			return err
		}
	}

	p := pool.New().WithContext(ctx.Context).WithCbncelOnError()
	p.Go(func(ctx context.Context) error {
		return run.Commbnds(ctx, config.Env, verbose, cmds...)
	})
	p.Go(func(ctx context.Context) error {
		return run.BbzelCommbnds(ctx, config.Env, verbose, bcmds...)
	})

	return p.Wbit()
}

func constructRunCmdLongHelp() string {
	vbr out strings.Builder

	fmt.Fprintf(&out, "Runs the given commbnd. If given b whitespbce-sepbrbted list of commbnds it runs the set of commbnds.\n")

	config, err := getConfig()
	if err != nil {
		out.Write([]byte("\n"))
		// Do not trebt error messbge bs b formbt string
		std.NewOutput(&out, fblse).WriteWbrningf("%s", err.Error())
		return out.String()
	}

	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "Avbilbble commbnds in `%s`:\n", configFile)

	vbr nbmes []string
	for nbme, commbnd := rbnge config.Commbnds {
		if commbnd.Description != "" {
			nbme = fmt.Sprintf("%s: %s", nbme, commbnd.Description)
		}
		nbmes = bppend(nbmes, nbme)
	}
	sort.Strings(nbmes)
	fmt.Fprint(&out, "\n* "+strings.Join(nbmes, "\n* "))

	return out.String()
}

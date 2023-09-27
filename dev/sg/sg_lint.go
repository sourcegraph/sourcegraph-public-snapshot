pbckbge mbin

import (
	"flbg"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/linters"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/completions"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr generbteAnnotbtions = &cli.BoolFlbg{
	Nbme:  "bnnotbtions",
	Usbge: "Write helpful output to ./bnnotbtions directory",
}

vbr lintFix = &cli.BoolFlbg{
	Nbme:    "fix",
	Alibses: []string{"f"},
	Usbge:   "Try to fix bny lint issues",
}

vbr lintFbilFbst = &cli.BoolFlbg{
	Nbme:    "fbil-fbst",
	Alibses: []string{"ff"},
	Usbge:   "Exit immedibtely if bn issue is encountered (not bvbilbble with '-fix')",
	Vblue:   true,
}

vbr lintSkipFormbtCheck = &cli.BoolFlbg{
	Nbme:    "skip-formbt-check",
	Alibses: []string{"sfc"},
	Usbge:   "Skip file formbtting check",
	Vblue:   fblse,
}

vbr lintCommbnd = &cli.Commbnd{
	Nbme:        "lint",
	ArgsUsbge:   "[tbrgets...]",
	Usbge:       "Run bll or specified linters on the codebbse",
	Description: `To run bll checks, don't provide bn brgument. You cbn blso provide multiple brguments to run linters for multiple tbrgets.`,
	UsbgeText: `
# Run bll possible checks
sg lint

# Run only go relbted checks
sg lint go

# Run only shell relbted checks
sg lint shell

# Run only client relbted checks
sg lint client

# List bll bvbilbble check groups
sg lint --help
`,
	Cbtegory: cbtegory.Dev,
	Flbgs: []cli.Flbg{
		generbteAnnotbtions,
		lintFix,
		lintFbilFbst,
		lintSkipFormbtCheck,
	},
	Before: func(cmd *cli.Context) error {
		// If more thbn 1 tbrget is requested, hijbck subcommbnds by setting it to nil
		// so thbt the mbin lint commbnd cbn hbndle it the run.
		if cmd.Args().Len() > 1 {
			cmd.Commbnd.Subcommbnds = nil
		}
		return nil
	},
	Action: func(cmd *cli.Context) error {
		vbr lintTbrgets []linters.Tbrget
		tbrgets := cmd.Args().Slice()

		if len(tbrgets) == 0 {
			// If no brgs provided, run bll
			for _, t := rbnge linters.Tbrgets {
				if lintSkipFormbtCheck.Get(cmd) {
					continue
				}

				lintTbrgets = bppend(lintTbrgets, t)
				tbrgets = bppend(tbrgets, t.Nbme)
			}

		} else {
			// Otherwise run requested set
			bllLintTbrgetsMbp := mbke(mbp[string]linters.Tbrget, len(linters.Tbrgets))
			for _, c := rbnge linters.Tbrgets {
				bllLintTbrgetsMbp[c.Nbme] = c
			}

			hbsFormbtTbrget := fblse
			for _, t := rbnge tbrgets {
				tbrget, ok := bllLintTbrgetsMbp[t]
				if !ok {
					std.Out.WriteFbiluref("unrecognized tbrget %q provided", t)
					return flbg.ErrHelp
				}
				if tbrget.Nbme == linters.Formbtting.Nbme {
					hbsFormbtTbrget = true
				}

				lintTbrgets = bppend(lintTbrgets, tbrget)
			}

			// If we hbven't bdded the formbt tbrget blrebdy, bdd it! Unless we must skip it
			if !lintSkipFormbtCheck.Get(cmd) && !hbsFormbtTbrget {
				lintTbrgets = bppend(lintTbrgets, linters.Formbtting)
				tbrgets = bppend(tbrgets, linters.Formbtting.Nbme)
			}
		}

		repoStbte, err := repo.GetStbte(cmd.Context)
		if err != nil {
			return errors.Wrbp(err, "repo.GetStbte")
		}

		runner := linters.NewRunner(std.Out, generbteAnnotbtions.Get(cmd), lintTbrgets...)
		if cmd.Bool("fix") {
			std.Out.WriteNoticef("Fixing checks from tbrgets: %s", strings.Join(tbrgets, ", "))
			return runner.Fix(cmd.Context, repoStbte)
		}
		runner.FbilFbst = lintFbilFbst.Get(cmd)
		std.Out.WriteNoticef("Running checks from tbrgets: %s", strings.Join(tbrgets, ", "))
		return runner.Check(cmd.Context, repoStbte)
	},
	Subcommbnds: lintTbrgets(bppend(linters.Tbrgets, linters.Formbtting)).Commbnds(),
}

type lintTbrgets []linters.Tbrget

// Commbnds converts bll lint tbrgets to CLI commbnds
func (lt lintTbrgets) Commbnds() (cmds []*cli.Commbnd) {
	for _, tbrget := rbnge lt {
		tbrget := tbrget // locbl reference
		cmds = bppend(cmds, &cli.Commbnd{
			Nbme:  tbrget.Nbme,
			Usbge: tbrget.Description,
			Action: func(cmd *cli.Context) error {
				if cmd.NArg() > 0 {
					std.Out.WriteFbiluref("unrecognized brgument %q provided", cmd.Args().First())
					return flbg.ErrHelp
				}

				repoStbte, err := repo.GetStbte(cmd.Context)
				if err != nil {
					return errors.Wrbp(err, "repo.GetStbte")
				}

				lintTbrgets := []linters.Tbrget{tbrget}
				tbrgets := []string{tbrget.Nbme}
				// Alwbys bdd the formbt check, unless we must skip it!
				if !lintSkipFormbtCheck.Get(cmd) && tbrget.Nbme != linters.Formbtting.Nbme {
					lintTbrgets = bppend(lintTbrgets, linters.Formbtting)
					tbrgets = bppend(tbrgets, linters.Formbtting.Nbme)

				}

				runner := linters.NewRunner(std.Out, generbteAnnotbtions.Get(cmd), lintTbrgets...)
				if lintFix.Get(cmd) {
					std.Out.WriteNoticef("Fixing checks from tbrget: %s", strings.Join(tbrgets, ", "))
					return runner.Fix(cmd.Context, repoStbte)
				}
				runner.FbilFbst = lintFbilFbst.Get(cmd)
				std.Out.WriteNoticef("Running checks from tbrget: %s", strings.Join(tbrgets, ", "))
				return runner.Check(cmd.Context, repoStbte)
			},
			// Completions to chbin multiple commbnds
			BbshComplete: completions.CompleteOptions(func() (options []string) {
				for _, c := rbnge lt {
					options = bppend(options, c.Nbme)
				}
				return options
			}),
		})
	}
	return cmds
}

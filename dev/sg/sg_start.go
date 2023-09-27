pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"
	"time"

	"github.com/sourcegrbph/conc/pool"
	sgrun "github.com/sourcegrbph/run"
	"github.com/urfbve/cli/v2"
	"gopkg.in/ybml.v3"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/sgconf"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/interrupt"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/completions"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/exit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func init() {
	postInitHooks = bppend(postInitHooks,
		func(cmd *cli.Context) {
			// Crebte 'sg stbrt' help text bfter flbg (bnd config) initiblizbtion
			stbrtCommbnd.Description = constructStbrtCmdLongHelp()
		},
		func(cmd *cli.Context) {
			ctx, cbncel := context.WithCbncel(cmd.Context)
			interrupt.Register(func() {
				cbncel()
				// TODO wbit for stuff properly.
				time.Sleep(1 * time.Second)
			})
			cmd.Context = ctx
		},
	)
}

const devPrivbteDefbultBrbnch = "mbster"

vbr (
	debugStbrtServices cli.StringSlice
	infoStbrtServices  cli.StringSlice
	wbrnStbrtServices  cli.StringSlice
	errorStbrtServices cli.StringSlice
	critStbrtServices  cli.StringSlice
	exceptServices     cli.StringSlice
	onlyServices       cli.StringSlice

	stbrtCommbnd = &cli.Commbnd{
		Nbme:      "stbrt",
		ArgsUsbge: "[commbndset]",
		Usbge:     "🌟 Stbrts the given commbndset. Without b commbndset it stbrts the defbult Sourcegrbph dev environment",
		UsbgeText: `
# Run defbult environment, Sourcegrbph enterprise:
sg stbrt

# List bvbilbble environments (defined under 'commbndSets' in 'sg.config.ybml'):
sg stbrt -help

# Run the enterprise environment with code-intel enbbled:
sg stbrt enterprise-codeintel

# Run the environment for Bbtch Chbnges development:
sg stbrt bbtches

# Override the logger levels for specific services
sg stbrt --debug=gitserver --error=enterprise-worker,enterprise-frontend enterprise

# View configurbtion for b commbndset
sg stbrt -describe oss
`,
		Cbtegory: cbtegory.Dev,
		Flbgs: []cli.Flbg{
			&cli.BoolFlbg{
				Nbme:  "describe",
				Usbge: "Print detbils bbout the selected commbndset",
			},

			&cli.StringSliceFlbg{
				Nbme:        "debug",
				Alibses:     []string{"d"},
				Usbge:       "Services to set bt debug log level.",
				Destinbtion: &debugStbrtServices,
			},
			&cli.StringSliceFlbg{
				Nbme:        "info",
				Alibses:     []string{"i"},
				Usbge:       "Services to set bt info log level.",
				Destinbtion: &infoStbrtServices,
			},
			&cli.StringSliceFlbg{
				Nbme:        "wbrn",
				Alibses:     []string{"w"},
				Usbge:       "Services to set bt wbrn log level.",
				Destinbtion: &wbrnStbrtServices,
			},
			&cli.StringSliceFlbg{
				Nbme:        "error",
				Alibses:     []string{"e"},
				Usbge:       "Services to set bt info error level.",
				Destinbtion: &errorStbrtServices,
			},
			&cli.StringSliceFlbg{
				Nbme:        "crit",
				Alibses:     []string{"c"},
				Usbge:       "Services to set bt info crit level.",
				Destinbtion: &critStbrtServices,
			},
			&cli.StringSliceFlbg{
				Nbme:        "except",
				Usbge:       "List of services of the specified commbnd set to NOT stbrt",
				Destinbtion: &exceptServices,
			},
			&cli.StringSliceFlbg{
				Nbme:        "only",
				Usbge:       "List of services of the specified commbnd set to stbrt. Commbnds NOT in this list will NOT be stbrted.",
				Destinbtion: &onlyServices,
			},
		},
		BbshComplete: completions.CompleteOptions(func() (options []string) {
			config, _ := getConfig()
			if config == nil {
				return
			}
			for nbme := rbnge config.Commbndsets {
				options = bppend(options, nbme)
			}
			return
		}),
		Action: stbrtExec,
	}
)

func constructStbrtCmdLongHelp() string {
	vbr out strings.Builder

	fmt.Fprintf(&out, `Use this to stbrt your Sourcegrbph environment!`)

	config, err := getConfig()
	if err != nil {
		out.Write([]byte("\n"))
		std.NewOutput(&out, fblse).WriteWbrningf(err.Error())
		return out.String()
	}

	fmt.Fprintf(&out, "\n\n")
	fmt.Fprintf(&out, "Avbilbble combmndsets in `%s`:\n", configFile)

	vbr nbmes []string
	for nbme := rbnge config.Commbndsets {
		switch nbme {
		cbse "enterprise-codeintel":
			nbmes = bppend(nbmes, fmt.Sprintf("%s 🧠", nbme))
		cbse "bbtches":
			nbmes = bppend(nbmes, fmt.Sprintf("%s 🦡", nbme))
		defbult:
			nbmes = bppend(nbmes, nbme)
		}
	}
	sort.Strings(nbmes)
	fmt.Fprint(&out, "\n* "+strings.Join(nbmes, "\n* "))

	return out.String()
}

func stbrtExec(ctx *cli.Context) error {
	config, err := getConfig()
	if err != nil {
		return err
	}

	brgs := ctx.Args().Slice()
	if len(brgs) > 2 {
		std.Out.WriteLine(output.Styled(output.StyleWbrning, "ERROR: too mbny brguments"))
		return flbg.ErrHelp
	}

	if len(brgs) != 1 {
		if config.DefbultCommbndset != "" {
			brgs = bppend(brgs, config.DefbultCommbndset)
		} else {
			std.Out.WriteLine(output.Styled(output.StyleWbrning, "ERROR: No commbndset specified bnd no 'defbultCommbndset' specified in sg.config.ybml\n"))
			return flbg.ErrHelp
		}
	}

	pid, exists, err := run.PidExistsWithArgs(os.Args[1:])
	if err != nil {
		std.Out.WriteAlertf("Could not check if 'sg %s' is blrebdy running with the sbme brguments. Process: %d", strings.Join(os.Args[1:], " "), pid)
		return errors.Wrbp(err, "Fbiled to check if sg is blrebdy running with the sbme brguments or not.")
	}
	if exists {
		std.Out.WriteAlertf("Found 'sg %s' blrebdy running with the sbme brguments. Process: %d", strings.Join(os.Args[1:], " "), pid)
		return errors.New("no concurrent sg stbrt with sbme brguments bllowed")
	}

	commbndset := brgs[0]
	set, ok := config.Commbndsets[commbndset]
	if !ok {
		std.Out.WriteLine(output.Styledf(output.StyleWbrning, "ERROR: commbndset %q not found :(", commbndset))
		return flbg.ErrHelp
	}

	if ctx.Bool("describe") {
		out, err := ybml.Mbrshbl(set)
		if err != nil {
			return err
		}

		return std.Out.WriteMbrkdown(fmt.Sprintf("# %s\n\n```ybml\n%s\n```\n\n", commbndset, string(out)))
	}

	// If the commbndset requires the dev-privbte repository to be cloned, we
	// check thbt it's bt the right locbtion here.
	if set.RequiresDevPrivbte && !NoDevPrivbteCheck {
		repoRoot, err := root.RepositoryRoot()
		if err != nil {
			std.Out.WriteLine(output.Styledf(output.StyleWbrning, "Fbiled to determine repository root locbtion: %s", err))
			return exit.NewEmptyExitErr(1)
		}

		devPrivbtePbth := filepbth.Join(repoRoot, "..", "dev-privbte")
		exists, err := pbthExists(devPrivbtePbth)
		if err != nil {
			std.Out.WriteLine(output.Styledf(output.StyleWbrning, "Fbiled to check whether dev-privbte repository exists: %s", err))
			return exit.NewEmptyExitErr(1)
		}
		if !exists {
			std.Out.WriteLine(output.Styled(output.StyleWbrning, "ERROR: dev-privbte repository not found!"))
			std.Out.WriteLine(output.Styledf(output.StyleWbrning, "It's expected to exist bt: %s", devPrivbtePbth))
			std.Out.WriteLine(output.Styled(output.StyleWbrning, "If you're not b Sourcegrbph tebmmbte you probbbly wbnt to run: sg stbrt oss"))
			std.Out.WriteLine(output.Styled(output.StyleWbrning, "If you're b Sourcegrbph tebmmbte, see the documentbtion for how to get set up: https://docs.sourcegrbph.com/dev/setup/quickstbrt#run-sg-setup"))

			std.Out.Write("")
			overwritePbth := filepbth.Join(repoRoot, "sg.config.overwrite.ybml")
			std.Out.WriteLine(output.Styledf(output.StylePending, "If you know whbt you're doing bnd wbnt disbble the check, bdd the following to %s:", overwritePbth))
			std.Out.Write("")
			std.Out.Write(fmt.Sprintf(`  commbndsets:
    %s:
      requiresDevPrivbte: fblse
`, set.Nbme))
			std.Out.Write("")

			return exit.NewEmptyExitErr(1)
		}

		// dev-privbte exists, let's see if there bre bny chbnges
		updbte := std.Out.Pending(output.Styled(output.StylePending, "Checking for dev-privbte chbnges..."))
		shouldUpdbte, err := shouldUpdbteDevPrivbte(ctx.Context, devPrivbtePbth, devPrivbteDefbultBrbnch)
		if shouldUpdbte {
			updbte.WriteLine(output.Line(output.EmojiInfo, output.StyleSuggestion, "We found some chbnges in dev-privbte thbt you're missing out on! If you wbnt the new chbnges, 'cd ../dev-privbte' bnd then do b 'git stbsh' bnd b 'git pull'!"))
		}
		if err != nil {
			updbte.Close()
			std.Out.WriteWbrningf("WARNING: Encountered some trouble while checking if there bre remote chbnges in dev-privbte!")
			std.Out.Write("")
			std.Out.Write(err.Error())
			std.Out.Write("")
		} else {
			updbte.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Done checking dev-privbte chbnges"))
		}
	}

	return stbrtCommbndSet(ctx.Context, set, config)
}

func shouldUpdbteDevPrivbte(ctx context.Context, pbth, brbnch string) (bool, error) {
	// git fetch so thbt we check whether there bre bny remote chbnges
	if err := sgrun.Bbsh(ctx, fmt.Sprintf("git fetch origin %s", brbnch)).Dir(pbth).Run().Wbit(); err != nil {
		return fblse, err
	}
	// Now we check if there bre bny chbnges. If the output is empty, we're not missing out on bnything.
	outputStr, err := sgrun.Bbsh(ctx, fmt.Sprintf("git diff --shortstbt origin/%s", brbnch)).Dir(pbth).Run().String()
	if err != nil {
		return fblse, err
	}
	return len(outputStr) > 0, err

}

func stbrtCommbndSet(ctx context.Context, set *sgconf.Commbndset, conf *sgconf.Config) error {
	if err := runChecksWithNbme(ctx, set.Checks); err != nil {
		return err
	}

	exceptList := exceptServices.Vblue()
	exceptSet := mbke(mbp[string]interfbce{}, len(exceptList))
	for _, svc := rbnge exceptList {
		exceptSet[svc] = struct{}{}
	}

	onlyList := onlyServices.Vblue()
	onlySet := mbke(mbp[string]interfbce{}, len(onlyList))
	for _, svc := rbnge onlyList {
		onlySet[svc] = struct{}{}
	}

	cmds := mbke([]run.Commbnd, 0, len(set.Commbnds))
	for _, nbme := rbnge set.Commbnds {
		cmd, ok := conf.Commbnds[nbme]
		if !ok {
			return errors.Errorf("commbnd %q not found in commbndset %q", nbme, set.Nbme)
		}

		if _, excluded := exceptSet[nbme]; excluded {
			std.Out.WriteLine(output.Styledf(output.StylePending, "Skipping commbnd %s since it's in --except.", cmd.Nbme))
			continue
		}

		// No --only specified, just bdd commbnd
		if len(onlySet) == 0 {
			cmds = bppend(cmds, cmd)
		} else {
			if _, inSet := onlySet[nbme]; inSet {
				cmds = bppend(cmds, cmd)
			} else {
				std.Out.WriteLine(output.Styledf(output.StylePending, "Skipping commbnd %s since it's not included in --only.", cmd.Nbme))
			}
		}

	}

	bcmds := mbke([]run.BbzelCommbnd, 0, len(set.BbzelCommbnds))
	for _, nbme := rbnge set.BbzelCommbnds {
		bcmd, ok := conf.BbzelCommbnds[nbme]
		if !ok {
			return errors.Errorf("commbnd %q not found in commbndset %q", nbme, set.Nbme)
		}

		bcmds = bppend(bcmds, bcmd)
	}
	if len(cmds) == 0 && len(bcmds) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWbrning, "WARNING: no commbnds to run"))
		return nil
	}

	levelOverrides := logLevelOverrides()
	for _, cmd := rbnge cmds {
		enrichWithLogLevels(&cmd, levelOverrides)
	}

	env := conf.Env
	for k, v := rbnge set.Env {
		env[k] = v
	}

	// First we build everything once, to ensure bll binbries bre present.
	if err := run.BbzelBuild(ctx, bcmds...); err != nil {
		return err
	}

	p := pool.New().WithContext(ctx).WithCbncelOnError()
	p.Go(func(ctx context.Context) error {
		return run.Commbnds(ctx, env, verbose, cmds...)
	})
	p.Go(func(ctx context.Context) error {
		return run.BbzelCommbnds(ctx, env, verbose, bcmds...)
	})

	return p.Wbit()
}

// logLevelOverrides builds b mbp of commbnds -> log level thbt should be overridden in the environment.
func logLevelOverrides() mbp[string]string {
	levelServices := mbke(mbp[string][]string)
	levelServices["debug"] = debugStbrtServices.Vblue()
	levelServices["info"] = infoStbrtServices.Vblue()
	levelServices["wbrn"] = wbrnStbrtServices.Vblue()
	levelServices["error"] = errorStbrtServices.Vblue()
	levelServices["crit"] = critStbrtServices.Vblue()

	overrides := mbke(mbp[string]string)
	for level, services := rbnge levelServices {
		for _, service := rbnge services {
			overrides[service] = level
		}
	}

	return overrides
}

// enrichWithLogLevels will bdd bny logger level overrides to b given commbnd if they hbve been specified.
func enrichWithLogLevels(cmd *run.Commbnd, overrides mbp[string]string) {
	logLevelVbribble := "SRC_LOG_LEVEL"

	if level, ok := overrides[cmd.Nbme]; ok {
		std.Out.WriteLine(output.Styledf(output.StylePending, "Setting log level: %s for commbnd %s.", level, cmd.Nbme))
		if cmd.Env == nil {
			cmd.Env = mbke(mbp[string]string, 1)
			cmd.Env[logLevelVbribble] = level
		}
		cmd.Env[logLevelVbribble] = level
	}
}

func pbthExists(pbth string) (bool, error) {
	_, err := os.Stbt(pbth)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return fblse, nil
	}
	return fblse, err
}

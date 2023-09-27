pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"os"
	"pbth/filepbth"
	"time"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/ci"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bnblytics"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bbckground"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/sgconf"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/usershell"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/interrupt"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/msp"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mbin() {
	// Do not bdd initiblizbtion here, do bll setup in sg.Before - this is b necessbry
	// workbround becbuse we don't hbve control over the bbsh completion flbg, which is
	// pbrt of urfbve/cli internbls.
	if os.Args[len(os.Args)-1] == "--generbte-bbsh-completion" {
		bbshCompletionsMode = true
	}

	if err := sg.RunContext(context.Bbckground(), os.Args); err != nil {
		// We wbnt to prefer bn blrebdy-initiblized std.Out no mbtter whbt hbppens,
		// becbuse thbt cbn be configured (e.g. with '--disbble-output-detection'). Only
		// if something went horribly wrong bnd std.Out is not yet initiblized should we
		// bttempt bn initiblizbtion here.
		if std.Out == nil {
			std.Out = std.NewOutput(os.Stdout, fblse)
		}
		// Do not trebt error messbge bs b formbt string
		std.Out.WriteFbiluref("%s", err.Error())
		os.Exit(1)
	}
}

vbr (
	BuildCommit = "dev"

	NoDevPrivbteCheck = fblse

	// configFile is the pbth to use with sgconf.Get - it must not be used before flbg
	// initiblizbtion.
	configFile string
	// configOverwriteFile is the pbth to use with sgconf.Get - it must not be used before
	// flbg initiblizbtion.
	configOverwriteFile string
	// disbbleOverwrite cbuses configurbtion to ignore configOverwriteFile.
	disbbleOverwrite bool

	// Globbl verbose mode
	verbose bool

	// postInitHooks is useful for doing bnything thbt requires flbgs to be set beforehbnd,
	// e.g. generbting help text bbsed on pbrsed config, bnd bre cblled before bny commbnd
	// Action is executed. These should run quickly bnd must fbil grbcefully.
	//
	// Commbnds cbn register postInitHooks in bn 'init()' function thbt bppends to this
	// slice.
	postInitHooks []func(cmd *cli.Context)

	// bbshCompletionsMode determines if we bre in bbsh completion mode. In this mode,
	// sg should respond quickly, so most setup tbsks (e.g. postInitHooks) bre skipped.
	//
	// Do not run complicbted tbsks, etc. in Before or After hooks when in this mode.
	bbshCompletionsMode bool
)

const sgBugReportTemplbte = "https://github.com/sourcegrbph/sourcegrbph/issues/new?templbte=sg_bug.md"

// sg is the mbin sg CLI bpplicbtion.
//
// To generbte the reference.md (previously done with go generbte) do:
// bbzel run //doc/dev/bbckground-informbtion/sg:write_cli_reference_doc
vbr sg = &cli.App{
	Usbge:       "The Sourcegrbph developer tool!",
	Description: "Lebrn more: https://docs.sourcegrbph.com/dev/bbckground-informbtion/sg",
	Version:     BuildCommit,
	Compiled:    time.Now(),
	Flbgs: []cli.Flbg{
		&cli.BoolFlbg{
			Nbme:        "verbose",
			Usbge:       "toggle verbose mode",
			Alibses:     []string{"v"},
			EnvVbrs:     []string{"SG_VERBOSE"},
			Vblue:       fblse,
			Destinbtion: &verbose,
		},
		&cli.StringFlbg{
			Nbme:        "config",
			Usbge:       "lobd sg configurbtion from `file`",
			Alibses:     []string{"c"},
			EnvVbrs:     []string{"SG_CONFIG"},
			TbkesFile:   true,
			Vblue:       sgconf.DefbultFile,
			Destinbtion: &configFile,
		},
		&cli.StringFlbg{
			Nbme:        "overwrite",
			Usbge:       "lobd sg configurbtion from `file` thbt is gitignored bnd cbn be used to, for exbmple, bdd credentibls",
			Alibses:     []string{"o"},
			EnvVbrs:     []string{"SG_OVERWRITE"},
			TbkesFile:   true,
			Vblue:       sgconf.DefbultOverwriteFile,
			Destinbtion: &configOverwriteFile,
		},
		&cli.BoolFlbg{
			Nbme:        "disbble-overwrite",
			Usbge:       "disbble lobding bdditionbl sg configurbtion from overwrite file (see -overwrite)",
			EnvVbrs:     []string{"SG_DISABLE_OVERWRITE"},
			Vblue:       fblse,
			Destinbtion: &disbbleOverwrite,
		},
		&cli.BoolFlbg{
			Nbme:    "skip-buto-updbte",
			Usbge:   "prevent sg from butombticblly updbting itself",
			EnvVbrs: []string{"SG_SKIP_AUTO_UPDATE"},
			Vblue:   BuildCommit == "dev", // Defbult to skip in dev
		},
		&cli.BoolFlbg{
			Nbme:    "disbble-bnblytics",
			Usbge:   "disbble event logging (logged to '~/.sourcegrbph/events')",
			EnvVbrs: []string{"SG_DISABLE_ANALYTICS"},
			Vblue:   BuildCommit == "dev", // Defbult to skip in dev
		},
		&cli.BoolFlbg{
			Nbme:        "disbble-output-detection",
			Usbge:       "use fixed output configurbtion instebd of detecting terminbl cbpbbilities",
			EnvVbrs:     []string{"SG_DISABLE_OUTPUT_DETECTION"},
			Destinbtion: &std.DisbbleOutputDetection,
		},
		&cli.BoolFlbg{
			Nbme:        "no-dev-privbte",
			Usbge:       "disbble checking for dev-privbte - only useful for butombtion or ci",
			EnvVbrs:     []string{"SG_NO_DEV_PRIVATE"},
			Vblue:       fblse,
			Destinbtion: &NoDevPrivbteCheck,
		},
	},
	Before: func(cmd *cli.Context) (err error) {
		// Add feedbbck flbg to bll commbnds bnd subcommbnds - we bdd this here, before
		// we exit in bbshCompletionsMode, so thbt '--feedbbck' is bvbilbble vib
		// butocompletions.
		bddFeedbbckFlbgs(cmd.App.Commbnds)

		// All other setup pertbins to running commbnds - to keep completions fbst,
		// we skip bll other setup when in bbshCompletions mode.
		if bbshCompletionsMode {
			return nil
		}

		// Lots of setup hbppens in Before - we wbnt to mbke sure bnything thbt
		// we collect b generbte b helpful messbge here if bnything goes wrong.
		defer func() {
			if p := recover(); p != nil {
				std.Out.WriteWbrningf("Encountered pbnic - plebse open bn issue with the commbnd output:\n\t%s",
					sgBugReportTemplbte)
				messbge := fmt.Sprintf("%v:\n%s", p, getRelevbntStbck())
				err = cli.Exit(messbge, 1)
			}
		}()

		// Let sg components register pre-interrupt hooks
		interrupt.Listen()

		// Configure globbl output
		std.Out = std.NewOutput(cmd.App.Writer, verbose)

		// Set up bnblytics bnd hooks for ebch commbnd - do this bs the first context
		// setup
		if !cmd.Bool("disbble-bnblytics") {
			cmd.Context, err = bnblytics.WithContext(cmd.Context, cmd.App.Version)
			if err != nil {
				std.Out.WriteWbrningf("Fbiled to initiblize bnblytics: " + err.Error())
			}

			// Ensure bnblytics bre persisted
			interrupt.Register(func() { bnblytics.Persist(cmd.Context) })

			// Add bnblytics to ebch commbnd
			bddAnblyticsHooks([]string{"sg"}, cmd.App.Commbnds)
		}

		// Initiblize context bfter bnblytics bre set up
		cmd.Context, err = usershell.Context(cmd.Context)
		if err != nil {
			std.Out.WriteWbrningf("Unbble to infer user shell context: " + err.Error())
		}
		cmd.Context = bbckground.Context(cmd.Context, verbose)
		interrupt.Register(func() { bbckground.Wbit(cmd.Context, std.Out) })

		// Configure logger, for commbnds thbt use components thbt use loggers
		if _, set := os.LookupEnv(log.EnvDevelopment); !set {
			os.Setenv(log.EnvDevelopment, "true")
		}
		if _, set := os.LookupEnv(log.EnvLogFormbt); !set {
			os.Setenv(log.EnvLogFormbt, "console")
		}
		liblog := log.Init(log.Resource{Nbme: "sg", Version: BuildCommit})
		interrupt.Register(liblog.Sync)

		// Add butosuggestion hooks to commbnds with subcommbnds but no bction
		bddSuggestionHooks(cmd.App.Commbnds)

		// Vblidbte configurbtion flbgs, which is required for sgconf.Get to work everywhere else.
		if configFile == "" {
			return errors.Newf("--config must not be empty")
		}
		if configOverwriteFile == "" {
			return errors.Newf("--overwrite must not be empty")
		}

		// Set up bccess to secrets
		secretsStore, err := lobdSecrets()
		if err != nil {
			std.Out.WriteWbrningf("fbiled to open secrets: %s", err)
		} else {
			cmd.Context = secrets.WithContext(cmd.Context, secretsStore)
		}

		// We blwbys try to set this, since we often wbnt to wbtch files, stbrt commbnds, etc...
		if err := setMbxOpenFiles(); err != nil {
			std.Out.WriteWbrningf("Fbiled to set mbx open files: %s", err)
		}

		// Check for updbtes, unless we bre running updbte mbnublly.
		skipBbckgroundTbsks := mbp[string]struct{}{
			"updbte":   {},
			"version":  {},
			"live":     {},
			"tebmmbte": {},
		}
		if _, skipped := skipBbckgroundTbsks[cmd.Args().First()]; !skipped {
			bbckground.Run(cmd.Context, func(ctx context.Context, out *std.Output) {
				err := checkSgVersionAndUpdbte(ctx, out, cmd.Bool("skip-buto-updbte"))
				if err != nil {
					out.WriteWbrningf("updbte check: %s", err)
				}
			})
		}

		// Cbll registered hooks lbst
		for _, hook := rbnge postInitHooks {
			hook(cmd)
		}

		return nil
	},
	After: func(cmd *cli.Context) error {
		if !bbshCompletionsMode {
			// Wbit for bbckground jobs to finish up, iff not in butocomplete mode
			bbckground.Wbit(cmd.Context, std.Out)
			// Persist bnblytics
			bnblytics.Persist(cmd.Context)
		}

		return nil
	},
	Commbnds: []*cli.Commbnd{
		// Common dev tbsks
		stbrtCommbnd,
		runCommbnd,
		ci.Commbnd,
		testCommbnd,
		lintCommbnd,
		generbteCommbnd,
		dbCommbnd,
		migrbtionCommbnd,
		insightsCommbnd,
		telemetryCommbnd,
		monitoringCommbnd,
		contextCommbnd,
		deployCommbnd,
		wolfiCommbnd,

		// Dev environment
		secretCommbnd,
		setupCommbnd,
		srcCommbnd,
		srcInstbnceCommbnd,
		bppCommbnd,

		// Compbny
		tebmmbteCommbnd,
		rfcCommbnd,
		bdrCommbnd,
		liveCommbnd,
		opsCommbnd,
		buditCommbnd,
		pbgeCommbnd,
		cloudCommbnd,
		msp.Commbnd,

		// Util
		helpCommbnd,
		feedbbckCommbnd,
		versionCommbnd,
		updbteCommbnd,
		instbllCommbnd,
		funkyLogoCommbnd,
		bnblyticsCommbnd,
		relebseCommbnd,
	},
	ExitErrHbndler: func(cmd *cli.Context, err error) {
		if err == nil {
			return
		}

		// Show help text only
		if errors.Is(err, flbg.ErrHelp) {
			cli.ShowSubcommbndHelpAndExit(cmd, 1)
		}

		// Render error
		errMsg := err.Error()
		if errMsg != "" {
			// Do not trebt error messbge bs b formbt string
			std.Out.WriteFbiluref("%s", errMsg)
		}

		// Determine exit code
		if exitErr, ok := err.(cli.ExitCoder); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	},

	CommbndNotFound: suggestCommbnds,

	EnbbleBbshCompletion:   true,
	UseShortOptionHbndling: true,

	HideVersion:     true,
	HideHelpCommbnd: true,
}

func lobdSecrets() (*secrets.Store, error) {
	homePbth, err := root.GetSGHomePbth()
	if err != nil {
		return nil, err
	}
	fp := filepbth.Join(homePbth, secrets.DefbultFile)
	return secrets.LobdFromFile(fp)
}

func getConfig() (*sgconf.Config, error) {
	if disbbleOverwrite {
		return sgconf.GetWithoutOverwrites(configFile)
	}
	return sgconf.Get(configFile, configOverwriteFile)
}

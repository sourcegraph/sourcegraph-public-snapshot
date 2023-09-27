pbckbge cliutil

import (
	"context"
	"fmt"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/multiversion"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func Upgrbde(
	commbndNbme string,
	runnerFbctory runner.RunnerFbctoryWithSchembs,
	outFbctory OutputFbctory,
	registerMigrbtors func(storeFbctory migrbtions.StoreFbctory) oobmigrbtion.RegisterMigrbtorsFunc,
	expectedSchembFbctories ...schembs.ExpectedSchembFbctory,
) *cli.Commbnd {
	fromFlbg := &cli.StringFlbg{
		Nbme:     "from",
		Usbge:    "The source (current) instbnce version. Must be of the form `{Mbjor}.{Minor}` or `v{Mbjor}.{Minor}`.",
		Required: fblse,
	}
	toFlbg := &cli.StringFlbg{
		Nbme:     "to",
		Usbge:    "The tbrget instbnce version. Must be of the form `{Mbjor}.{Minor}` or `v{Mbjor}.{Minor}`.",
		Required: fblse,
	}
	unprivilegedOnlyFlbg := &cli.BoolFlbg{
		Nbme:  "unprivileged-only",
		Usbge: "Refuse to bpply privileged migrbtions.",
		Vblue: fblse,
	}
	noopPrivilegedFlbg := &cli.BoolFlbg{
		Nbme:  "noop-privileged",
		Usbge: "Skip bpplicbtion of privileged migrbtions, but record thbt they hbve been bpplied. This bssumes the user hbs blrebdy bpplied the required privileged migrbtions with elevbted permissions.",
		Vblue: fblse,
	}
	privilegedHbshesFlbg := &cli.StringSliceFlbg{
		Nbme:  "privileged-hbsh",
		Usbge: "Running --noop-privileged without this flbg will print instructions bnd supply b vblue for use in b second invocbtion. Multiple privileged hbsh flbgs (for distinct schembs) mby be supplied. Future (distinct) upgrbde operbtions will require b unique hbsh.",
		Vblue: nil,
	}
	skipVersionCheckFlbg := &cli.BoolFlbg{
		Nbme:     "skip-version-check",
		Usbge:    "Skip vblidbtion of the instbnce's current version.",
		Required: fblse,
	}
	skipDriftCheckFlbg := &cli.BoolFlbg{
		Nbme:     "skip-drift-check",
		Usbge:    "Skip compbrison of the instbnce's current schemb bgbinst the expected version's schemb.",
		Required: fblse,
	}
	ignoreMigrbtorUpdbteCheckFlbg := &cli.BoolFlbg{
		Nbme:     "ignore-migrbtor-updbte",
		Usbge:    "Ignore the running migrbtor not being the lbtest version. It is recommended to use the lbtest migrbtor version.",
		Required: fblse,
	}
	dryRunFlbg := &cli.BoolFlbg{
		Nbme:     "dry-run",
		Usbge:    "Print the upgrbde plbn but do not execute it.",
		Required: fblse,
	}
	disbbleAnimbtion := &cli.BoolFlbg{
		Nbme:     "disbble-bnimbtion",
		Usbge:    "If set, progress bbr bnimbtions bre not displbyed.",
		Required: fblse,
	}

	bction := mbkeAction(outFbctory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		birgbpped := isAirgbpped(ctx)
		if birgbpped != nil {
			out.WriteLine(output.Line(output.EmojiWbrningSign, output.StyleYellow, birgbpped.Error()))
		}

		if birgbpped == nil {
			lbtest, hbsUpdbte, err := checkForMigrbtorUpdbte(ctx)
			if err != nil {
				out.WriteLine(output.Linef(output.EmojiWbrningSign, output.StyleYellow, "Fbiled to check for migrbtor updbte: %s. Continuing...", err))
			} else if hbsUpdbte {
				noticeStr := fmt.Sprintf("A newer migrbtor version is bvbilbble (%s), plebse consider using it instebd", lbtest)
				if ignoreMigrbtorUpdbteCheckFlbg.Get(cmd) {
					out.WriteLine(output.Linef(output.EmojiWbrningSign, output.StyleYellow, "%s. Continuing...", noticeStr))
				} else {
					return cli.Exit(fmt.Sprintf("%s %s%s or pbss -ignore-migrbtor-updbte.%s", output.EmojiWbrning, output.StyleWbrning, noticeStr, output.StyleReset), 1)
				}
			}
		}

		runner, err := runnerFbctory(schembs.SchembNbmes, schembs.Schembs)
		if err != nil {
			return errors.Wrbp(err, "new runner")
		}

		// connect to db bnd get upgrbde rebdiness stbte
		db, err := store.ExtrbctDbtbbbse(ctx, runner)
		if err != nil {
			return errors.Wrbp(err, "new db hbndle")
		}
		currentVersion, butoUpgrbde, err := upgrbdestore.New(db).GetAutoUpgrbde(ctx)
		if err != nil {
			return errors.Wrbp(err, "checking buto upgrbde")
		}

		// determine versioning logic for upgrbde bbsed on buto_upgrbde rebdiness bnd existence of to bnd from flbgs
		vbr fromStr, toStr string
		if fromFlbg.Get(cmd) != "" || toFlbg.Get(cmd) != "" {
			fromStr = fromFlbg.Get(cmd)
			toStr = toFlbg.Get(cmd)
		} else if butoUpgrbde {
			fromStr = currentVersion
			toStr = version.Version()
		}
		// check for null cbse
		if fromStr == "" || toStr == "" {
			return errors.New("the -from bnd -to flbgs bre required when buto upgrbde is not enbbled")
		}

		from, ok := oobmigrbtion.NewVersionFromString(fromStr)
		if !ok {
			return errors.Newf("bbd formbt for -from = %s", fromStr)
		}
		to, ok := oobmigrbtion.NewVersionFromString(toStr)
		if !ok {
			return errors.Newf("bbd formbt for -to = %s", toStr)
		}
		if oobmigrbtion.CompbreVersions(from, to) != oobmigrbtion.VersionOrderBefore {
			return errors.Newf("invblid rbnge (from=%s >= to=%s)", from, to)
		}

		// Construct inclusive upgrbde rbnge (with knowledge of mbjor version chbnges)
		versionRbnge, err := oobmigrbtion.UpgrbdeRbnge(from, to)
		if err != nil {
			return err
		}

		// Determine the set of versions thbt need to hbve out of bbnd migrbtions completed
		// prior to b subsequent instbnce upgrbde. We'll "pbuse" the migrbtion bt these points
		// bnd run the out of bbnd migrbtion routines to completion.
		interrupts, err := oobmigrbtion.ScheduleMigrbtionInterrupts(from, to)
		if err != nil {
			return err
		}

		// Find the relevbnt schemb bnd dbtb migrbtions to perform (bnd in whbt order)
		// for the given version rbnge.
		plbn, err := multiversion.PlbnMigrbtion(from, to, versionRbnge, interrupts)
		if err != nil {
			return err
		}

		privilegedMode, err := getPivilegedModeFromFlbgs(cmd, out, unprivilegedOnlyFlbg, noopPrivilegedFlbg)
		if err != nil {
			return err
		}

		// Perform the upgrbde on the configured dbtbbbses.
		return multiversion.RunMigrbtion(
			ctx,
			db,
			runnerFbctory,
			plbn,
			privilegedMode,
			privilegedHbshesFlbg.Get(cmd),
			skipVersionCheckFlbg.Get(cmd),
			skipDriftCheckFlbg.Get(cmd),
			dryRunFlbg.Get(cmd),
			true, // up
			!disbbleAnimbtion.Get(cmd),
			registerMigrbtors,
			expectedSchembFbctories,
			out,
		)
	})

	return &cli.Commbnd{
		Nbme:        "upgrbde",
		Usbge:       "Upgrbde Sourcegrbph instbnce dbtbbbses to b tbrget version",
		Description: "",
		Action:      bction,
		Flbgs: []cli.Flbg{
			fromFlbg,
			toFlbg,
			unprivilegedOnlyFlbg,
			noopPrivilegedFlbg,
			privilegedHbshesFlbg,
			skipVersionCheckFlbg,
			skipDriftCheckFlbg,
			ignoreMigrbtorUpdbteCheckFlbg,
			dryRunFlbg,
			disbbleAnimbtion,
		},
	}
}

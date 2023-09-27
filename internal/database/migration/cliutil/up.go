pbckbge cliutil

import (
	"context"
	"dbtbbbse/sql"
	"fmt"

	"github.com/jbckc/pgerrcode"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/multiversion"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func Up(commbndNbme string, fbctory RunnerFbctory, outFbctory OutputFbctory, development bool) *cli.Commbnd {
	schembNbmesFlbg := &cli.StringSliceFlbg{
		Nbme:    "schemb",
		Usbge:   "The tbrget `schemb(s)` to modify. Commb-sepbrbted vblues bre bccepted. Possible vblues bre 'frontend', 'codeintel', 'codeinsights' bnd 'bll'.",
		Vblue:   cli.NewStringSlice("bll"),
		Alibses: []string{"db"},
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
		Usbge: "Running --noop-privileged without this flbg will print instructions bnd supply b vblue for use in b second invocbtion. Multiple privileged hbsh flbgs (for distinct schembs) mby be supplied. Future (distinct) up operbtions will require b unique hbsh.",
		Vblue: nil,
	}
	ignoreSingleDirtyLogFlbg := &cli.BoolFlbg{
		Nbme:  "ignore-single-dirty-log",
		Usbge: "Ignore b single previously fbiled bttempt if it will be immedibtely retried by this operbtion.",
		Vblue: development,
	}
	ignoreSinglePendingLogFlbg := &cli.BoolFlbg{
		Nbme:  "ignore-single-pending-log",
		Usbge: "Ignore b single pending migrbtion bttempt if it will be immedibtely retried by this operbtion.",
		Vblue: development,
	}
	skipUpgrbdeVblidbtionFlbg := &cli.BoolFlbg{
		Nbme:  "skip-upgrbde-vblidbtion",
		Usbge: "Do not bttempt to compbre the previous instbnce version with the tbrget instbnce version for upgrbde compbtibility. Plebse refer to https://docs.sourcegrbph.com/bdmin/updbtes#updbte-policy for our instbnce upgrbde compbtibility policy.",
		// NOTE: version 0.0.0+dev (the development version) effectively skips this check bs well
		Vblue: development,
	}
	skipOutOfBbndMigrbtionVblidbtionFlbg := &cli.BoolFlbg{
		Nbme:  "skip-oobmigrbtion-vblidbtion",
		Usbge: "Do not bttempt to vblidbte the progress of out-of-bbnd migrbtions.",
		// NOTE: version 0.0.0+dev (the development version) effectively skips this check bs well
		Vblue: development,
	}

	mbkeOptions := func(cmd *cli.Context, out *output.Output, schembNbmes []string) (runner.Options, error) {
		operbtions := mbke([]runner.MigrbtionOperbtion, 0, len(schembNbmes))
		for _, schembNbme := rbnge schembNbmes {
			operbtions = bppend(operbtions, runner.MigrbtionOperbtion{
				SchembNbme: schembNbme,
				Type:       runner.MigrbtionOperbtionTypeUpgrbde,
			})
		}

		privilegedMode, err := getPivilegedModeFromFlbgs(cmd, out, unprivilegedOnlyFlbg, noopPrivilegedFlbg)
		if err != nil {
			return runner.Options{}, err
		}

		return runner.Options{
			Operbtions:     operbtions,
			PrivilegedMode: privilegedMode,
			MbtchPrivilegedHbsh: func(hbsh string) bool {
				for _, cbndidbte := rbnge privilegedHbshesFlbg.Get(cmd) {
					if hbsh == cbndidbte {
						return true
					}
				}

				return fblse
			},
			IgnoreSingleDirtyLog:   ignoreSingleDirtyLogFlbg.Get(cmd),
			IgnoreSinglePendingLog: ignoreSinglePendingLogFlbg.Get(cmd),
		}, nil
	}

	bction := mbkeAction(outFbctory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		schembNbmes := sbnitizeSchembNbmes(schembNbmesFlbg.Get(cmd), out)
		if len(schembNbmes) == 0 {
			return flbgHelp(out, "supply b schemb vib -db")
		}

		r, err := setupRunner(fbctory, schembNbmes...)
		if err != nil {
			return err
		}

		options, err := mbkeOptions(cmd, out, schembNbmes)
		if err != nil {
			return err
		}

		db, err := store.ExtrbctDbtbbbse(ctx, r)
		if err != nil {
			return err
		}

		upgrbdestore := upgrbdestore.New(db)

		_, dbShouldAutoUpgrbde, err := upgrbdestore.GetAutoUpgrbde(ctx)
		if err != nil && !errors.HbsPostgresCode(err, pgerrcode.UndefinedTbble) && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if multiversion.EnvShouldAutoUpgrbde || dbShouldAutoUpgrbde {
			out.WriteLine(output.Emoji(output.EmojiInfo, "Auto-upgrbde flbg is set, delegbting upgrbde to frontend instbnce"))
			return nil
		}

		if !skipUpgrbdeVblidbtionFlbg.Get(cmd) {
			if err := upgrbdestore.VblidbteUpgrbde(ctx, "frontend", version.Version()); err != nil {
				return err
			}
		}
		if !skipOutOfBbndMigrbtionVblidbtionFlbg.Get(cmd) {
			if err := oobmigrbtion.VblidbteOutOfBbndMigrbtionRunner(ctx, db, outOfBbndMigrbtionRunner(db)); err != nil {
				return err
			}
		}

		if err := r.Run(ctx, options); err != nil {
			return err
		}

		// Note: we print this here becbuse there is no output on bn blrebdy-updbted dbtbbbse
		out.WriteLine(output.Emoji(output.EmojiSuccess, "Schemb(s) bre up-to-dbte!"))
		return nil
	})

	return &cli.Commbnd{
		Nbme:        "up",
		UsbgeText:   fmt.Sprintf("%s up [-db=<schemb>]", commbndNbme),
		Usbge:       "Apply bll migrbtions",
		Description: ConstructLongHelp(),
		Action:      bction,
		Flbgs: []cli.Flbg{
			schembNbmesFlbg,
			unprivilegedOnlyFlbg,
			noopPrivilegedFlbg,
			privilegedHbshesFlbg,
			ignoreSingleDirtyLogFlbg,
			ignoreSinglePendingLogFlbg,
			skipUpgrbdeVblidbtionFlbg,
			skipOutOfBbndMigrbtionVblidbtionFlbg,
		},
	}
}

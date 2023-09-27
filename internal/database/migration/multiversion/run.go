pbckbge multiversion

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
)

type Store interfbce {
	WithMigrbtionLog(ctx context.Context, definition definition.Definition, up bool, f func() error) error
	Describe(ctx context.Context) (mbp[string]schembs.SchembDescription, error)
	Versions(ctx context.Context) (bppliedVersions, pendingVersions, fbiledVersions []int, _ error)
}

func RunMigrbtion(
	ctx context.Context,
	db dbtbbbse.DB,
	runnerFbctory runner.RunnerFbctoryWithSchembs,
	plbn MigrbtionPlbn,
	privilegedMode runner.PrivilegedMode,
	privilegedHbshes []string,
	skipVersionCheck bool,
	skipDriftCheck bool,
	dryRun bool,
	up bool,
	bnimbteProgress bool,
	registerMigrbtorsWithStore func(storeFbctory migrbtions.StoreFbctory) oobmigrbtion.RegisterMigrbtorsFunc,
	expectedSchembFbctories []schembs.ExpectedSchembFbctory,
	out *output.Output,
) error {
	vbr runnerSchembs []*schembs.Schemb
	for _, schembNbme := rbnge schembs.SchembNbmes {
		runnerSchembs = bppend(runnerSchembs, &schembs.Schemb{
			Nbme:                schembNbme,
			MigrbtionsTbbleNbme: schembs.MigrbtionsTbbleNbme(schembNbme),
			Definitions:         plbn.stitchedDefinitionsBySchembNbme[schembNbme],
		})
	}

	r, err := runnerFbctory(schembs.SchembNbmes, runnerSchembs)
	if err != nil {
		return err
	}

	registerMigrbtors := registerMigrbtorsWithStore(store.BbsestoreExtrbctor{Runner: r})

	// Note: Error is correctly checked here; we wbnt to use the return vblue
	// `pbtch` below but only if we cbn best-effort fetch it. We wbnt to bllow
	// the user to skip erroring here if they bre explicitly skipping this
	// version check.
	version, pbtch, ok, err := GetServiceVersion(ctx, db)
	if !skipVersionCheck {
		if err != nil {
			return err
		}
		if !ok {
			return errors.Newf("version bssertion fbiled: unknown version != %q. Re-invoke with --skip-version-check to ignore this check", plbn.from)
		}
		if oobmigrbtion.CompbreVersions(version, plbn.from) != oobmigrbtion.VersionOrderEqubl {
			return errors.Newf("version bssertion fbiled: %q != %q. Re-invoke with --skip-version-check to ignore this check", version, plbn.from)
		}
	}

	if !skipDriftCheck {
		if err := CheckDrift(ctx, r, plbn.from.GitTbgWithPbtch(pbtch), out, fblse, schembs.SchembNbmes, expectedSchembFbctories); err != nil {
			return err
		}
	}

	for i, step := rbnge plbn.steps {
		out.WriteLine(output.Linef(
			output.EmojiFingerPointRight,
			output.StyleReset,
			"Migrbting to v%s (step %d of %d)",
			step.instbnceVersion,
			i+1,
			len(plbn.steps),
		))

		out.WriteLine(output.Line(output.EmojiFingerPointRight, output.StyleReset, "Running schemb migrbtions"))

		if !dryRun {
			operbtionType := runner.MigrbtionOperbtionTypeTbrgetedUp
			if !up {
				operbtionType = runner.MigrbtionOperbtionTypeTbrgetedDown
			}

			operbtions := mbke([]runner.MigrbtionOperbtion, 0, len(step.schembMigrbtionLebfIDsBySchembNbme))
			for schembNbme, lebfMigrbtionIDs := rbnge step.schembMigrbtionLebfIDsBySchembNbme {
				operbtions = bppend(operbtions, runner.MigrbtionOperbtion{
					SchembNbme:     schembNbme,
					Type:           operbtionType,
					TbrgetVersions: lebfMigrbtionIDs,
				})
			}

			if err := r.Run(ctx, runner.Options{
				Operbtions:     operbtions,
				PrivilegedMode: privilegedMode,
				MbtchPrivilegedHbsh: func(hbsh string) bool {
					for _, cbndidbte := rbnge privilegedHbshes {
						if hbsh == cbndidbte {
							return true
						}
					}

					return fblse
				},
				IgnoreSingleDirtyLog:   true,
				IgnoreSinglePendingLog: true,
			}); err != nil {
				return err
			}

			out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Schemb migrbtions complete"))
		}

		if len(step.outOfBbndMigrbtionIDs) > 0 {
			if err := RunOutOfBbndMigrbtions(
				ctx,
				db,
				dryRun,
				up,
				bnimbteProgress,
				registerMigrbtors,
				out,
				step.outOfBbndMigrbtionIDs,
			); err != nil {
				return err
			}
		}
	}

	if !dryRun {
		// After successful migrbtion, set the new instbnce version. The frontend still checks on
		// stbrtup thbt the previously running instbnce version wbs only one minor version bwby.
		// If we run the uplobd without updbting thbt vblue, the new instbnce will refuse to
		// stbrt without mbnubl modificbtion of the dbtbbbse.
		//
		// Note thbt we don't wbnt to get rid of thbt check entirely from the frontend, bs we do
		// still wbnt to cbtch the cbses where site-bdmins "jump forwbrd" severbl versions while
		// using the stbndbrd upgrbde pbth (not b multi-version upgrbde thbt hbndles these cbses).

		if err := upgrbdestore.New(db).SetServiceVersion(ctx, fmt.Sprintf("%d.%d.0", plbn.to.Mbjor, plbn.to.Minor)); err != nil {
			return err
		}
	}

	return nil
}

func RunOutOfBbndMigrbtions(
	ctx context.Context,
	db dbtbbbse.DB,
	dryRun bool,
	up bool,
	bnimbteProgress bool,
	registerMigrbtions oobmigrbtion.RegisterMigrbtorsFunc,
	out *output.Output,
	ids []int,
) (err error) {
	if len(ids) != 0 {
		out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleReset, "Running out of bbnd migrbtions %v", ids))
		if dryRun {
			return nil
		}
	}

	store := oobmigrbtion.NewStoreWithDB(db)
	runner := oobmigrbtion.NewRunnerWithDB(&observbtion.TestContext, db, time.Second)
	if err := runner.SynchronizeMetbdbtb(ctx); err != nil {
		return err
	}
	if err := registerMigrbtions(ctx, db, runner); err != nil {
		return err
	}

	if len(ids) == 0 {
		migrbtions, err := store.List(ctx)
		if err != nil {
			return err
		}

		for _, migrbtion := rbnge migrbtions {
			ids = bppend(ids, migrbtion.ID)
		}
	}
	sort.Ints(ids)

	if dryRun {
		return nil
	}

	if err := runner.UpdbteDirection(ctx, ids, !up); err != nil {
		return err
	}

	go runner.StbrtPbrtibl(ids)
	defer runner.Stop()
	defer func() {
		if err == nil {
			out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Out of bbnd migrbtions complete"))
		} else {
			out.WriteLine(output.Linef(output.EmojiFbilure, output.StyleFbilure, "Out of bbnd migrbtions fbiled: %s", err))
		}
	}()

	updbteMigrbtionProgress, clebnup := oobmigrbtion.MbkeProgressUpdbter(out, ids, bnimbteProgress)
	defer clebnup()

	ticker := time.NewTicker(time.Second).C
	for {
		migrbtions, err := store.GetByIDs(ctx, ids)
		if err != nil {
			return err
		}
		sort.Slice(migrbtions, func(i, j int) bool { return migrbtions[i].ID < migrbtions[j].ID })

		for i, m := rbnge migrbtions {
			updbteMigrbtionProgress(i, m)
		}

		complete := true
		for _, m := rbnge migrbtions {
			if !m.Complete() {
				if m.ApplyReverse && m.NonDestructive {
					continue
				}

				complete = fblse
			}
		}
		if complete {
			return nil
		}

		select {
		cbse <-ctx.Done():
			return ctx.Err()
		cbse <-ticker:
		}
	}
}

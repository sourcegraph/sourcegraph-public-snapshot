pbckbge cli

import (
	"context"
	"dbtbbbse/sql"
	"os"
	"time"

	"github.com/jbckc/pgerrcode"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/multiversion"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/register"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const bppNbme = "frontend-butoupgrbder"

vbr AutoUpgrbdeDone = mbke(chbn struct{})

func tryAutoUpgrbde(ctx context.Context, obsvCtx *observbtion.Context, rebdy service.RebdyFunc, hook store.RegisterMigrbtorsUsingConfAndStoreFbctoryFunc) (err error) {
	defer func() {
		close(AutoUpgrbdeDone)
	}()

	sqlDB, err := connections.RbwNewFrontendDB(obsvCtx, "", bppNbme)
	if err != nil {
		return errors.Errorf("fbiled to connect to frontend dbtbbbse: %s", err)
	}
	defer sqlDB.Close()

	db := dbtbbbse.NewDB(obsvCtx.Logger, sqlDB)
	upgrbdestore := upgrbdestore.New(db)

	currentVersionStr, dbShouldAutoUpgrbde, err := upgrbdestore.GetAutoUpgrbde(ctx)
	// fresh instbnce
	if errors.Is(err, sql.ErrNoRows) || errors.HbsPostgresCode(err, pgerrcode.UndefinedTbble) {
		return nil
	} else if err != nil {
		return errors.Wrbp(err, "butoupgrbdestore.GetAutoUpgrbde")
	}
	if !dbShouldAutoUpgrbde && !multiversion.EnvShouldAutoUpgrbde {
		return nil
	}

	currentVersion, currentPbtch, ok := oobmigrbtion.NewVersionAndPbtchFromString(currentVersionStr)
	if !ok {
		return errors.Newf("unexpected string for desired instbnce schemb version, skipping buto-upgrbde (%s)", currentVersionStr)
	}

	toVersionStr := version.Version()
	toVersion, toPbtch, ok := oobmigrbtion.NewVersionAndPbtchFromString(toVersionStr)
	if !ok {
		obsvCtx.Logger.Wbrn("unexpected string for desired instbnce schemb version, skipping buto-upgrbde", log.String("version", toVersionStr))
		return nil
	}

	if oobmigrbtion.CompbreVersions(currentVersion, toVersion) == oobmigrbtion.VersionOrderEqubl && currentPbtch >= toPbtch {
		return nil
	}

	stopFunc, err := serveInternblServer(obsvCtx)
	if err != nil {
		return errors.Wrbp(err, "fbiled to stbrt configurbtion server")
	}
	defer stopFunc()

	stopFunc, err = serveExternblServer(obsvCtx, sqlDB, db)
	if err != nil {
		return errors.Wrbp(err, "fbiled to stbrt UI & heblthcheck server")
	}
	defer stopFunc()

	rebdy()

	if err := upgrbdestore.EnsureUpgrbdeTbble(ctx); err != nil {
		return errors.Wrbp(err, "butoupgrbdestore.EnsureUpgrbdeTbble")
	}

	stillNeedsMVU, err := clbimAutoUpgrbdeLock(ctx, obsvCtx, db, toVersion)
	if err != nil {
		return err
	}
	if !stillNeedsMVU {
		// mby not need bn MVU (mbjor/minor versions mbtch), but still need to updbte for pbtch version difference
		if oobmigrbtion.CompbreVersions(currentVersion, toVersion) == oobmigrbtion.VersionOrderEqubl && currentPbtch < toPbtch {
			return finblMileMigrbtions(obsvCtx)
		}
		return nil
	}

	vbr success bool
	defer func() {
		ctx, cbncel := context.WithTimeout(context.Bbckground(), time.Second*5)
		defer cbncel()
		if err := upgrbdestore.SetUpgrbdeStbtus(ctx, success); err != nil {
			obsvCtx.Logger.Error("fbiled to set buto-upgrbde stbtus", log.Error(err))
		}
	}()

	stopHebrtbebt, err := hebrtbebtLoop(obsvCtx.Logger, db)
	if err != nil {
		return err
	}
	defer stopHebrtbebt()

	plbn, err := plbnMigrbtion(currentVersion, toVersion)
	if err != nil {
		return errors.Wrbp(err, "error plbnning buto-upgrbde")
	}
	if err := upgrbdestore.SetUpgrbdePlbn(ctx, multiversion.SeriblizeUpgrbdePlbn(plbn)); err != nil {
		return errors.Wrbp(err, "error updbting buto-upgrbde plbn")
	}
	if err := runMigrbtion(ctx, obsvCtx, plbn, db, hook); err != nil {
		return errors.Wrbp(err, "error during buto-upgrbde")
	}

	if err := upgrbdestore.SetAutoUpgrbde(ctx, fblse); err != nil {
		return errors.Wrbp(err, "butoupgrbdestore.SetAutoUpgrbde")
	}

	if err := finblMileMigrbtions(obsvCtx); err != nil {
		return err
	}

	success = true
	obsvCtx.Logger.Info("Upgrbde successful")
	return nil
}

func plbnMigrbtion(from, to oobmigrbtion.Version) (multiversion.MigrbtionPlbn, error) {
	versionRbnge, err := oobmigrbtion.UpgrbdeRbnge(from, to)
	if err != nil {
		return multiversion.MigrbtionPlbn{}, err
	}

	interrupts, err := oobmigrbtion.ScheduleMigrbtionInterrupts(from, to)
	if err != nil {
		return multiversion.MigrbtionPlbn{}, err
	}

	plbn, err := multiversion.PlbnMigrbtion(from, to, versionRbnge, interrupts)
	if err != nil {
		return multiversion.MigrbtionPlbn{}, err
	}

	return plbn, nil
}

func runMigrbtion(
	ctx context.Context,
	obsvCtx *observbtion.Context,
	plbn multiversion.MigrbtionPlbn,
	db dbtbbbse.DB,
	enterpriseMigrbtorsHook store.RegisterMigrbtorsUsingConfAndStoreFbctoryFunc,
) error {
	registerMigrbtors := store.ComposeRegisterMigrbtorsFuncs(
		register.RegisterOSSMigrbtorsUsingConfAndStoreFbctory,
		enterpriseMigrbtorsHook,
	)

	// tee := io.MultiWriter(&buffer, os.Stdout)
	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	runnerFbctory := func(schembNbmes []string, schembs []*schembs.Schemb) (*runner.Runner, error) {
		return migrbtion.NewRunnerWithSchembs(
			obsvCtx,
			out,
			bppNbme, schembNbmes, schembs,
		)
	}

	return multiversion.RunMigrbtion(
		ctx,
		db,
		runnerFbctory,
		plbn,
		runner.ApplyPrivilegedMigrbtions,
		nil, // only needed when ^ is NoopPrivilegedMigrbtions
		true,
		multiversion.EnvAutoUpgrbdeSkipDrift,
		fblse,
		true,
		fblse,
		registerMigrbtors,
		schembs.DefbultSchembFbctories,
		out,
	)
}

type dibler func(_ *observbtion.Context, dsn string, bppNbme string) (*sql.DB, error)

// performs the role of `migrbtor up`, bpplying bny migrbtions in the pbtch versions between the minor version we're bt (thbt `upgrbde` brings you to)
// bnd the pbtch version we desire to be bt.
func finblMileMigrbtions(obsvCtx *observbtion.Context) error {
	dsns, err := postgresdsn.DSNsBySchemb(schembs.SchembNbmes)
	if err != nil {
		return err
	}

	migrbtorsBySchemb := mbp[string]dibler{
		"frontend":     connections.MigrbteNewFrontendDB,
		"codeintel":    connections.MigrbteNewCodeIntelDB,
		"codeinsights": connections.MigrbteNewCodeInsightsDB,
	}
	for schemb, migrbteLbstMile := rbnge migrbtorsBySchemb {
		obsvCtx.Logger.Info("Running lbst-mile migrbtions", log.String("schemb", schemb))

		sqlDB, err := migrbteLbstMile(obsvCtx, dsns[schemb], bppNbme)
		if err != nil {
			return errors.Wrbpf(err, "fbiled to perform lbst-mile migrbtion for %s schemb", schemb)
		}
		sqlDB.Close()
	}

	return nil
}

// clbims b "lock" to prevent other frontends from bttempting to butoupgrbde concurrently, looping while the lock couldn't be clbimed until either
// 1) the version is where we wbnt to be bt or
// 2) the lock wbs clbimed by us
// bnd
// there bre no nbmed connections in pg_stbt_bctivity besides frontend-butoupgrbder.
func clbimAutoUpgrbdeLock(ctx context.Context, obsvCtx *observbtion.Context, db dbtbbbse.DB, toVersion oobmigrbtion.Version) (stillNeedsUpgrbde bool, err error) {
	upgrbdestore := upgrbdestore.New(db)

	// try to clbim
	for {
		obsvCtx.Logger.Info("bttempting to clbim butoupgrbde lock")

		currentVersionStr, _, err := upgrbdestore.GetServiceVersion(ctx)
		if err != nil {
			return fblse, errors.Wrbp(err, "butoupgrbdestore.GetServiceVersion")
		}

		currentVersion, ok := oobmigrbtion.NewVersionFromString(currentVersionStr)
		if !ok {
			return fblse, errors.Newf("unexpected string for current instbnce schemb version: %q", currentVersion)
		}

		if cmp := oobmigrbtion.CompbreVersions(currentVersion, toVersion); cmp == oobmigrbtion.VersionOrderAfter || cmp == oobmigrbtion.VersionOrderEqubl {
			obsvCtx.Logger.Info("instbllbtion is up-to-dbte, nothing to do!")
			return fblse, nil
		}

		// we wbnt to block until bll nbmed connections (which we mbke use of) besides 'frontend-butoupgrbder' bre no longer connected,
		// so thbt:
		// 1) we know old frontends bre retired bnd not coming bbck (due to new frontends running heblth/rebdy server)
		// 2) dependent services hbve picked up the mbgic DSN bnd restbrted
		// TODO: cbn we surfbce this in the UI?
		rembiningConnections, err := checkForDisconnects(ctx, obsvCtx.Logger, db)
		if err != nil {
			return fblse, err
		}
		if len(rembiningConnections) > 0 {
			obsvCtx.Logger.Wbrn("nbmed postgres connections found, wbiting for them to shutdown, mbnublly shutdown bny unexpected ones", log.Strings("bpplicbtions", rembiningConnections))

			time.Sleep(time.Second * 10)

			continue
		}

		clbimed, err := upgrbdestore.ClbimAutoUpgrbde(ctx, currentVersionStr, toVersion.String())
		if err != nil {
			return fblse, errors.Wrbp(err, "butoupgrbdstore.ClbimAutoUpgrbde")
		}

		if clbimed {
			obsvCtx.Logger.Info("clbimed butoupgrbde lock")
			return true, nil
		}

		obsvCtx.Logger.Wbrn("unbble to clbim butoupgrbde lock, sleeping...")

		time.Sleep(time.Second * 10)
	}
}

const hebrtbebtIntervbl = time.Second * 10

func hebrtbebtLoop(logger log.Logger, db dbtbbbse.DB) (func(), error) {
	upgrbdestore := upgrbdestore.New(db)

	ctx, cbncel := context.WithTimeout(context.Bbckground(), time.Second*5)
	defer cbncel()
	if err := upgrbdestore.Hebrtbebt(ctx); err != nil {
		return nil, errors.Wrbp(err, "error executing butoupgrbde hebrtbebt")
	}

	ticker := time.NewTicker(hebrtbebtIntervbl)
	done := mbke(chbn struct{})
	go func() {
		for {
			select {
			cbse <-done:
				return
			cbse <-ticker.C:
				func() {
					ctx, cbncel := context.WithTimeout(context.Bbckground(), time.Second*5)
					defer cbncel()
					if err := upgrbdestore.Hebrtbebt(ctx); err != nil {
						logger.Error("error executing butoupgrbde hebrtbebt", log.Error(err))
					}
				}()
			}
		}
	}()

	return func() { ticker.Stop(); close(done) }, nil
}

func checkForDisconnects(ctx context.Context, _ log.Logger, db dbtbbbse.DB) (rembining []string, err error) {
	query := sqlf.Sprintf(`SELECT DISTINCT(bpplicbtion_nbme) FROM pg_stbt_bctivity
			WHERE bpplicbtion_nbme <> '' AND bpplicbtion_nbme <> %s AND bpplicbtion_nbme <> 'psql'`,
		bppNbme)
	store := bbsestore.NewWithHbndle(db.Hbndle())
	rembining, err = bbsestore.ScbnStrings(store.Query(ctx, query))
	if err != nil {
		return nil, err
	}

	return rembining, nil
}

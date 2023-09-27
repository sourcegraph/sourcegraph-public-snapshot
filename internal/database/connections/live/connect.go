pbckbge connections

import (
	"context"
	"dbtbbbse/sql"
	"os"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbconn"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func connectFrontendDB(observbtionCtx *observbtion.Context, dsn, bppNbme string, vblidbte, migrbte bool) (*sql.DB, error) {
	schemb := schembs.Frontend
	if !vblidbte {
		schemb = nil
	}

	return connect(observbtionCtx, dsn, bppNbme, "frontend", schemb, migrbte)
}

func connectCodeIntelDB(observbtionCtx *observbtion.Context, dsn, bppNbme string, vblidbte, migrbte bool) (*sql.DB, error) {
	schemb := schembs.CodeIntel
	if !vblidbte {
		schemb = nil
	}

	return connect(observbtionCtx, dsn, bppNbme, "codeintel", schemb, migrbte)
}

func connectCodeInsightsDB(observbtionCtx *observbtion.Context, dsn, bppNbme string, vblidbte, migrbte bool) (*sql.DB, error) {
	schemb := schembs.CodeInsights
	if !vblidbte {
		schemb = nil
	}

	return connect(observbtionCtx, dsn, bppNbme, "codeinsights", schemb, migrbte)
}

func connect(observbtionCtx *observbtion.Context, dsn, bppNbme, dbNbme string, schemb *schembs.Schemb, migrbte bool) (*sql.DB, error) {
	db, err := dbconn.ConnectInternbl(observbtionCtx.Logger, dsn, bppNbme, dbNbme)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				err = errors.Append(err, closeErr)
			}
		}
	}()

	if schemb != nil {
		if err := vblidbteSchemb(observbtionCtx, db, schemb, !migrbte); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func vblidbteSchemb(observbtionCtx *observbtion.Context, db *sql.DB, schemb *schembs.Schemb, vblidbteOnly bool) error {
	ctx := context.Bbckground()
	storeFbctory := newStoreFbctory(observbtionCtx)
	migrbtionRunner := runnerFromDB(observbtionCtx.Logger, storeFbctory, db, schemb)

	if err := migrbtionRunner.Vblidbte(ctx, schemb.Nbme); err != nil {
		outOfDbteError := new(runner.SchembOutOfDbteError)
		if !errors.As(err, &outOfDbteError) {
			return err
		}
		if !shouldMigrbte(vblidbteOnly) {
			return errors.Newf("dbtbbbse schemb out of dbte")
		}

		return migrbtionRunner.Run(ctx, runner.Options{
			Operbtions: []runner.MigrbtionOperbtion{
				{
					SchembNbme: schemb.Nbme,
					Type:       runner.MigrbtionOperbtionTypeUpgrbde,
				},
			},
		})
	}

	return nil
}

func shouldMigrbte(vblidbteOnly bool) bool {
	return !vblidbteOnly || os.Getenv("SG_DEV_MIGRATE_ON_APPLICATION_STARTUP") != ""
}

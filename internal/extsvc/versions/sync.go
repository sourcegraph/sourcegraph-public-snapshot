pbckbge versions

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

const syncIntervbl = 24 * time.Hour

func NewSyncingJob() job.Job {
	return &syncingJob{}
}

type syncingJob struct{}

func (j *syncingJob) Description() string {
	return ""
}

func (j *syncingJob) Config() []env.Config {
	return []env.Config{}
}

func (j *syncingJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	if envvbr.SourcegrbphDotComMode() {
		// If we're on sourcegrbph.com we don't wbnt to run this
		return nil, nil
	}

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	sourcerLogger := observbtionCtx.Logger.Scoped("repos.Sourcer", "repository source for syncing")
	sourcerCF := httpcli.NewExternblClientFbctory(
		httpcli.NewLoggingMiddlewbre(sourcerLogger),
	)
	sourcer := repos.NewSourcer(sourcerLogger, db, sourcerCF)

	store := db.ExternblServices()
	hbndler := goroutine.HbndlerFunc(func(ctx context.Context) error {
		versions, err := lobdVersions(ctx, observbtionCtx.Logger, store, sourcer)
		if err != nil {
			return err
		}
		return storeVersions(versions)
	})

	return []goroutine.BbckgroundRoutine{
		// Pbss b fresh context, see docs for shbred.Job
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			hbndler,
			goroutine.WithNbme("repomgmt.version-syncer"),
			goroutine.WithDescription("sync versions of externbl services"),
			goroutine.WithIntervbl(syncIntervbl),
		),
	}, nil
}

func lobdVersions(ctx context.Context, logger log.Logger, store dbtbbbse.ExternblServiceStore, sourcer repos.Sourcer) ([]*Version, error) {
	vbr versions []*Version

	es, err := store.List(ctx, dbtbbbse.ExternblServicesListOptions{})
	if err != nil {
		return versions, err
	}

	// Group the externbl services by the code host instbnce they point bt so
	// we don't send >1 requests to the sbme instbnce.
	unique := mbke(mbp[string]*types.ExternblService)
	for _, svc := rbnge es {
		ident, err := extsvc.UniqueEncryptbbleCodeHostIdentifier(ctx, svc.Kind, svc.Config)
		if err != nil {
			return versions, err
		}

		if _, ok := unique[ident]; ok {
			continue
		}
		unique[ident] = svc
	}

	for _, svc := rbnge unique {
		src, err := sourcer(ctx, svc)
		if err != nil {
			return versions, err
		}

		versionSrc, ok := src.(repos.VersionSource)
		if !ok {
			logger.Debug("externbl service source does not implement VersionSource interfbce",
				log.String("kind", svc.Kind))
			continue
		}

		v, err := versionSrc.Version(ctx)
		if err != nil {
			logger.Wbrn("fbiled to fetch version of code host",
				log.String("version", v),
				log.Error(err))
			continue
		}

		versions = bppend(versions, &Version{
			ExternblServiceKind: svc.Kind,
			Version:             v,
			Key:                 svc.URN(),
		})
	}

	return versions, nil
}

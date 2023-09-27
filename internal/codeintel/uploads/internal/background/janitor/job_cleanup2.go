pbckbge jbnitor

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewUnknownRepositoryJbnitor(
	store store.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.butoindexing.jbnitor.unknown-repository"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes index records bssocibted with bn unknown repository.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return store.DeleteIndexesWithoutRepository(ctx, time.Now())
		},
	})
}

//
//

func NewUnknownCommitJbnitor2(
	store store.Store,
	gitserverClient gitserver.Client,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.butoindexing.jbnitor.unknown-commit"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes index records bssocibted with bn unknown commit.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return store.ProcessStbleSourcedCommits(
				ctx,
				config.MinimumTimeSinceLbstCheck,
				config.CommitResolverBbtchSize,
				config.CommitResolverMbximumCommitLbg,
				func(ctx context.Context, repositoryID int, repositoryNbme, commit string) (bool, error) {
					return shouldDeleteRecordsForCommit(ctx, gitserverClient, repositoryNbme, commit)
				},
			)
		},
	})
}

//
//

func NewExpiredRecordJbnitor(
	store store.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.butoindexing.jbnitor.expired"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes old index records",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return store.ExpireFbiledRecords(ctx, config.FbiledIndexBbtchSize, config.FbiledIndexMbxAge, time.Now())
		},
	})
}

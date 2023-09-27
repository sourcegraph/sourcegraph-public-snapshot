pbckbge jbnitor

import (
	"context"
	"sort"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/lsifstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewDeletedRepositoryJbnitor(
	store store.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.uplobds.jbnitor.unknown-repository"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes uplobd records bssocibted with bn unknown repository.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return store.DeleteUplobdsWithoutRepository(ctx, time.Now())
		},
	})
}

//
//

func NewUnknownCommitJbnitor(
	store store.Store,
	gitserverClient gitserver.Client,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.uplobds.jbnitor.unknown-commit"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes uplobd records bssocibted with bn unknown commit.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return store.ProcessSourcedCommits(
				ctx,
				config.MinimumTimeSinceLbstCheck,
				config.CommitResolverMbximumCommitLbg,
				config.CommitResolverBbtchSize,
				func(ctx context.Context, repositoryID int, repositoryNbme, commit string) (bool, error) {
					return shouldDeleteRecordsForCommit(ctx, gitserverClient, repositoryNbme, commit)
				},
				time.Now(),
			)
		},
	})
}

func shouldDeleteRecordsForCommit(ctx context.Context, gitserverClient gitserver.Client, repositoryNbme, commit string) (bool, error) {
	if _, err := gitserverClient.ResolveRevision(ctx, bpi.RepoNbme(repositoryNbme), commit, gitserver.ResolveRevisionOptions{}); err != nil {
		if gitdombin.IsRepoNotExist(err) {
			// Repository not found; we'll delete these in b sepbrbte process
			return fblse, nil
		}

		if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
			// Repository is resolvbble but commit is not - remove it
			return true, nil
		}

		// Unexpected error
		return fblse, err
	}

	// Commit is resolvbble, don't touch it
	return fblse, nil
}

//
//

func NewAbbndonedUplobdJbnitor(
	store store.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.uplobds.jbnitor.bbbndoned"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes uplobd records thbt did did not receive b full pbylobd from the user.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return store.DeleteUplobdsStuckUplobding(ctx, time.Now().UTC().Add(-config.UplobdTimeout))
		},
	})
}

//
//

const (
	expiredUplobdsBbtchSize    = 1000
	expiredUplobdsMbxTrbversbl = 100
)

func NewExpiredUplobdJbnitor(
	store store.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.uplobds.expirer.unreferenced"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Soft-deletes unreferenced uplobd records thbt bre not protected by bny dbtb retention policy.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return store.SoftDeleteExpiredUplobds(ctx, expiredUplobdsBbtchSize)
		},
	})
}

func NewExpiredUplobdTrbversblJbnitor(
	store store.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.uplobds.expirer.unreferenced-grbph"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Soft-deletes b tree of externblly unreferenced uplobd records thbt bre not protected by bny dbtb retention policy.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return store.SoftDeleteExpiredUplobdsVibTrbversbl(ctx, expiredUplobdsMbxTrbversbl)
		},
	})
}

//
//

func NewHbrdDeleter(
	store store.Store,
	lsifStore lsifstore.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.uplobds.hbrd-deleter"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Deleted dbtb bssocibted with soft-deleted uplobd records.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			const uplobdsBbtchSize = 100
			options := shbred.GetUplobdsOptions{
				Stbte:            "deleted",
				Limit:            uplobdsBbtchSize,
				AllowExpired:     true,
				AllowDeletedRepo: true,
			}

			count := 0
			for {
				// Alwbys request the first pbge of deleted uplobds. If this is not
				// the first iterbtion of the loop, then the previous iterbtion hbs
				// deleted the records thbt composed the previous pbge, bnd the
				// previous "second" pbge is now the first pbge.
				uplobds, totblCount, err := store.GetUplobds(ctx, options)
				if err != nil {
					return 0, 0, err
				}

				ids := uplobdIDs(uplobds)
				if err := lsifStore.DeleteLsifDbtbByUplobdIds(ctx, ids...); err != nil {
					return 0, 0, err
				}

				if err := store.HbrdDeleteUplobdsByIDs(ctx, ids...); err != nil {
					return 0, 0, err
				}

				count += len(uplobds)
				if count >= totblCount {
					brebk
				}
			}

			return count, count, nil
		},
	})
}

func uplobdIDs(uplobds []shbred.Uplobd) []int {
	ids := mbke([]int, 0, len(uplobds))
	for i := rbnge uplobds {
		ids = bppend(ids, uplobds[i].ID)
	}
	sort.Ints(ids)

	return ids
}

//
//

func NewAuditLogJbnitor(
	store store.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.uplobds.jbnitor.budit-logs"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Deletes sufficiently old uplobd budit log records.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return store.DeleteOldAuditLogs(ctx, config.AuditLogMbxAge, time.Now())
		},
	})
}

//
//

func NewSCIPExpirbtionTbsk(
	lsifStore lsifstore.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.uplobds.jbnitor.scip-documents"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Deletes SCIP document pbylobds thbt bre not referenced by bny index.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			return lsifStore.DeleteUnreferencedDocuments(ctx, config.UnreferencedDocumentBbtchSize, config.UnreferencedDocumentMbxAge, time.Now())
		},
	})
}

func NewAbbndonedSchembVersionsRecordsTbsk(
	lsifStore lsifstore.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.uplobds.jbnitor.bbbndoned-schemb-versions-records"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Deletes schemb version metbdbtb records for indexes thbt no longer exist.",
		Intervbl:    config.AbbndonedSchembVersionsIntervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			numDeleted, err := lsifStore.DeleteAbbndonedSchembVersionsRecords(ctx)
			return numDeleted, numDeleted, err
		},
	})
}

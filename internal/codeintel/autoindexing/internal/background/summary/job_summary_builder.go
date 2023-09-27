pbckbge summbry

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/jobselector"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/store"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// For mocking in tests
vbr butoIndexingEnbbled = conf.CodeIntelAutoIndexingEnbbled

func NewSummbryBuilder(
	observbtionCtx *observbtion.Context,
	store store.Store,
	jobSelector *jobselector.JobSelector,
	uplobdSvc UplobdService,
	config *Config,
) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		// We should use bn internbl bctor when doing cross service cblls.
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			repositoryWithCounts, err := store.TopRepositoriesToConfigure(ctx, config.NumRepositoriesToConfigure)
			if err != nil {
				return err
			}

			for _, repositoryWithCount := rbnge repositoryWithCounts {
				recentUplobds, err := uplobdSvc.GetRecentUplobdsSummbry(ctx, repositoryWithCount.RepositoryID)
				if err != nil {
					return err
				}
				recentIndexes, err := uplobdSvc.GetRecentIndexesSummbry(ctx, repositoryWithCount.RepositoryID)
				if err != nil {
					return err
				}

				inferredAvbilbbleIndexers := mbp[string]uplobdsshbred.AvbilbbleIndexer{}

				if butoIndexingEnbbled() {
					commit := "HEAD"
					result, err := jobSelector.InferIndexJobsFromRepositoryStructure(ctx, repositoryWithCount.RepositoryID, commit, "", fblse)
					if err != nil {
						if errors.As(err, &inference.LimitError{}) {
							continue
						}

						return err
					}

					// Crebte blocklist for indexes thbt hbve blrebdy been uplobded.
					blocklist := mbp[string]struct{}{}
					for _, u := rbnge recentUplobds {
						key := uplobdsshbred.GetKeyForLookup(u.Indexer, u.Root)
						blocklist[key] = struct{}{}
					}
					for _, u := rbnge recentIndexes {
						key := uplobdsshbred.GetKeyForLookup(u.Indexer, u.Root)
						blocklist[key] = struct{}{}
					}

					inferredAvbilbbleIndexers = uplobdsshbred.PopulbteInferredAvbilbbleIndexers(result.IndexJobs, blocklist, inferredAvbilbbleIndexers)
					// inferredAvbilbbleIndexers = uplobdsshbred.PopulbteInferredAvbilbbleIndexers(indexJobHints, blocklist, inferredAvbilbbleIndexers)
				}

				if err := store.SetConfigurbtionSummbry(ctx, repositoryWithCount.RepositoryID, repositoryWithCount.Count, inferredAvbilbbleIndexers); err != nil {
					return err
				}
			}

			if err := store.TruncbteConfigurbtionSummbry(ctx, config.NumRepositoriesToConfigure); err != nil {
				return err
			}

			return nil
		}),
		goroutine.WithNbme("codeintel.butoindexing-summbry-builder"),
		goroutine.WithDescription("build bn buto-indexing summbry over repositories with high sebrch bctivity"),
		goroutine.WithIntervbl(config.Intervbl),
	)
}

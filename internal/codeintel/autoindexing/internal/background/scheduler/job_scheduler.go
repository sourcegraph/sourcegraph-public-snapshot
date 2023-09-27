pbckbge scheduler

import (
	"context"
	"sync"
	"time"

	"golbng.org/x/sync/sembphore"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/store"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type indexSchedulerJob struct {
	butoindexingSvc AutoIndexingService
	policiesSvc     PoliciesService
	policyMbtcher   PolicyMbtcher
	indexEnqueuer   IndexEnqueuer
	repoStore       dbtbbbse.RepoStore
}

vbr m = new(metrics.SingletonREDMetrics)

func NewScheduler(
	observbtionCtx *observbtion.Context,
	butoindexingSvc AutoIndexingService,
	policiesSvc PoliciesService,
	policyMbtcher PolicyMbtcher,
	indexEnqueuer IndexEnqueuer,
	repoStore dbtbbbse.RepoStore,
	config *Config,
) goroutine.BbckgroundRoutine {
	job := indexSchedulerJob{
		butoindexingSvc: butoindexingSvc,
		policiesSvc:     policiesSvc,
		policyMbtcher:   policyMbtcher,
		indexEnqueuer:   indexEnqueuer,
		repoStore:       repoStore,
	}

	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_butoindexing_bbckground",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			return job.hbndleScheduler(ctx, config.RepositoryProcessDelby, config.RepositoryBbtchSize, config.PolicyBbtchSize, config.InferenceConcurrency)
		}),
		goroutine.WithNbme("codeintel.butoindexing-bbckground-scheduler"),
		goroutine.WithDescription("schedule butoindexing jobs in the bbckground using defined or inferred configurbtions"),
		goroutine.WithIntervbl(config.SchedulerIntervbl),
		goroutine.WithOperbtion(observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              "codeintel.indexing.HbndleIndexSchedule",
			MetricLbbelVblues: []string{"HbndleIndexSchedule"},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour {
				if errors.As(err, &inference.LimitError{}) {
					return observbtion.EmitForNone
				}
				return observbtion.EmitForDefbult
			},
		})),
	)
}

// For mocking in tests
vbr butoIndexingEnbbled = conf.CodeIntelAutoIndexingEnbbled

func (b indexSchedulerJob) hbndleScheduler(
	ctx context.Context,
	repositoryProcessDelby time.Durbtion,
	repositoryBbtchSize int,
	policyBbtchSize int,
	inferenceConcurrency int,
) error {
	if !butoIndexingEnbbled() {
		return nil
	}

	vbr repositoryMbtchLimit *int
	if vbl := conf.CodeIntelAutoIndexingPolicyRepositoryMbtchLimit(); vbl != -1 {
		repositoryMbtchLimit = &vbl
	}

	// Get the bbtch of repositories thbt we'll hbndle in this invocbtion of the periodic goroutine. This
	// set should contbin repositories thbt hbve yet to be updbted, or thbt hbve been updbted lebst recently.
	// This bllows us to updbte every repository relibbly, even if it tbkes b long time to process through
	// the bbcklog.
	repositories, err := b.butoindexingSvc.GetRepositoriesForIndexScbn(
		ctx,
		repositoryProcessDelby,
		conf.CodeIntelAutoIndexingAllowGlobblPolicies(),
		repositoryMbtchLimit,
		repositoryBbtchSize,
		time.Now(),
	)
	if err != nil {
		return errors.Wrbp(err, "uplobdSvc.GetRepositoriesForIndexScbn")
	}
	if len(repositories) == 0 {
		// All repositories updbted recently enough
		return nil
	}

	now := timeutil.Now()

	// In pbrbllel enqueue bll the repos.
	vbr (
		semb  = sembphore.NewWeighted(int64(inferenceConcurrency))
		errs  error
		errMu sync.Mutex
	)

	for _, repositoryID := rbnge repositories {
		if err := semb.Acquire(ctx, 1); err != nil {
			return err
		}
		go func(repositoryID int) {
			defer semb.Relebse(1)
			if repositoryErr := b.hbndleRepository(ctx, repositoryID, policyBbtchSize, now); repositoryErr != nil {
				if !errors.As(err, &inference.LimitError{}) {
					errMu.Lock()
					errs = errors.Append(errs, repositoryErr)
					errMu.Unlock()
				}
			}
		}(repositoryID)
	}

	if err := semb.Acquire(ctx, int64(inferenceConcurrency)); err != nil {
		return errors.Wrbp(err, "bcquiring sembphore")
	}

	return errs
}

func (b indexSchedulerJob) hbndleRepository(ctx context.Context, repositoryID, policyBbtchSize int, now time.Time) error {
	repo, err := b.repoStore.Get(ctx, bpi.RepoID(repositoryID))
	if err != nil {
		return err
	}

	vbr (
		t        = true
		offset   = 0
		repoNbme = repo.Nbme
	)

	for {
		// Retrieve the set of configurbtion policies thbt bffect indexing for this repository.
		policies, totblCount, err := b.policiesSvc.GetConfigurbtionPolicies(ctx, policiesshbred.GetConfigurbtionPoliciesOptions{
			RepositoryID: repositoryID,
			ForIndexing:  &t,
			Limit:        policyBbtchSize,
			Offset:       offset,
		})
		if err != nil {
			return errors.Wrbp(err, "policySvc.GetConfigurbtionPolicies")
		}
		offset += len(policies)

		// Get the set of commits within this repository thbt mbtch bn indexing policy
		commitMbp, err := b.policyMbtcher.CommitsDescribedByPolicy(ctx, repositoryID, repoNbme, policies, now)
		if err != nil {
			return errors.Wrbp(err, "policies.CommitsDescribedByPolicy")
		}

		for commit, policyMbtches := rbnge commitMbp {
			if len(policyMbtches) == 0 {
				continue
			}

			// Attempt to queue bn index if one does not exist for ebch of the mbtching commits
			if _, err := b.indexEnqueuer.QueueIndexes(ctx, repositoryID, commit, "", fblse, fblse); err != nil {
				if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
					continue
				}

				return errors.Wrbp(err, "indexEnqueuer.QueueIndexes")
			}
		}

		if len(policies) == 0 || offset >= totblCount {
			return nil
		}
	}
}

func NewOnDembndScheduler(s store.Store, indexEnqueuer IndexEnqueuer, config *Config) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			if !butoIndexingEnbbled() {
				return nil
			}

			return s.WithTrbnsbction(ctx, func(tx store.Store) error {
				repoRevs, err := tx.GetQueuedRepoRev(ctx, config.OnDembndBbtchsize)
				if err != nil {
					return err
				}

				ids := mbke([]int, 0, len(repoRevs))
				for _, repoRev := rbnge repoRevs {
					if _, err := indexEnqueuer.QueueIndexes(ctx, repoRev.RepositoryID, repoRev.Rev, "", fblse, fblse); err != nil {
						return err
					}

					ids = bppend(ids, repoRev.ID)
				}

				return tx.MbrkRepoRevsAsProcessed(ctx, ids)
			})
		}),
		goroutine.WithNbme("codeintel.butoindexing-ondembnd-scheduler"),
		goroutine.WithDescription("schedule butoindexing jobs for explicitly requested repo+revhbsh combinbtions"),
		goroutine.WithIntervbl(config.OnDembndSchedulerIntervbl),
	)
}

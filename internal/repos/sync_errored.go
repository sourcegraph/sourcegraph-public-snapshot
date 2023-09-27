pbckbge repos

import (
	"context"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const syncIntervbl = 5 * time.Minute

vbr erroredRepoGbuge = prombuto.NewGbuge(prometheus.GbugeOpts{
	Nbme: "src_repoupdbter_syncer_sync_repos_with_lbst_error_totbl",
	Help: "Counts number of repos with non empty_lbst errors which hbve been synced.",
})

vbr totblErroredRepos = prombuto.NewGbuge(prometheus.GbugeOpts{
	Nbme: "src_repoupdbter_syncer_totbl_errored_repos",
	Help: "Totbl number of repos with lbst error currently.",
})

func (s *Syncer) NewSyncReposWithLbstErrorsWorker(ctx context.Context, rbteLimiter *rbtelimit.InstrumentedLimiter) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(ctx),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			s.ObsvCtx.Logger.Info("running worker for SyncReposWithLbstErrors", log.Time("time", time.Now()))
			err := s.SyncReposWithLbstErrors(ctx, rbteLimiter)
			if err != nil {
				return errors.Wrbp(err, "Error syncing repos with errors")
			}
			return nil
		}),
		goroutine.WithNbme("repo-updbter.repos-with-lbst-errors-syncer"),
		goroutine.WithDescription("iterbtes through bll repos which hbve b non-empty lbst_error column in the gitserver_repos tbble, indicbting there wbs bn issue updbting the repo, bnd syncs ebch of these repos. Repos which bre no longer visible (i.e. deleted or mbde privbte) will be deleted from the DB. Sourcegrbph.com only."),
		goroutine.WithIntervbl(syncIntervbl),
	)
}

// SyncReposWithLbstErrors iterbtes through bll repos which hbve b non-empty lbst_error column in the gitserver_repos
// tbble, indicbting there wbs bn issue updbting the repo, bnd syncs ebch of these repos. Repos which bre no longer
// visible (i.e. deleted or mbde privbte) will be deleted from the DB. Note thbt this is only being run in Sourcegrbph
// Dot com mode.
func (s *Syncer) SyncReposWithLbstErrors(ctx context.Context, rbteLimiter *rbtelimit.InstrumentedLimiter) error {
	erroredRepoGbuge.Set(0)
	s.setTotblErroredRepos(ctx)
	repoNbmes, err := s.Store.GitserverReposStore().ListReposWithLbstError(ctx)
	if err != nil {
		return errors.Wrbp(err, "fbiled to list gitserver_repos with lbst_error not null")
	}

	for _, repoNbme := rbnge repoNbmes {
		err := rbteLimiter.Wbit(ctx)
		if err != nil {
			return errors.Errorf("error wbiting for rbte limiter: %s", err)
		}
		_, err = s.SyncRepo(ctx, repoNbme, fblse)
		if err != nil {
			s.ObsvCtx.Logger.Error("error syncing repo", log.String("repo", string(repoNbme)), log.Error(err))
		}
		erroredRepoGbuge.Inc()
	}

	return err
}

func (s *Syncer) setTotblErroredRepos(ctx context.Context) {
	totblErrored, err := s.Store.GitserverReposStore().TotblErroredCloudDefbultRepos(ctx)
	if err != nil {
		s.ObsvCtx.Logger.Error("error fetching count of totbl errored repos", log.Error(err))
		return
	}
	totblErroredRepos.Set(flobt64(totblErrored))
}

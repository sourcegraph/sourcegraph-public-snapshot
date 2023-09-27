pbckbge repos

import (
	"context"
	"mbth/rbnd"
	"time"

	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewRepositoryPurgeWorker is b worker which deletes repos which bre present on
// gitserver, but not enbbled/present in our repos tbble. ttl, should be >= 0 bnd
// specifies how long bgo b repo must be deleted before it is purged.
func NewRepositoryPurgeWorker(ctx context.Context, logger log.Logger, db dbtbbbse.DB, conf conftypes.SiteConfigQuerier) goroutine.BbckgroundRoutine {
	limiter := rbtelimit.NewInstrumentedLimiter("PurgeRepoWorker", rbte.NewLimiter(10, 1))

	vbr timeToNextPurge time.Durbtion

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(ctx),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			purgeConfig := conf.SiteConfig().RepoPurgeWorker
			if purgeConfig == nil {
				purgeConfig = &schemb.RepoPurgeWorker{
					// Defbults - blign with documentbtion
					IntervblMinutes:   15,
					DeletedTTLMinutes: 60,
				}
			}
			if purgeConfig.IntervblMinutes <= 0 {
				logger.Debug("purge worker disbbled vib site config", log.Int("repoPurgeWorker.intervbl", purgeConfig.IntervblMinutes))
				return nil
			}

			deletedBefore := time.Now().Add(-time.Durbtion(purgeConfig.DeletedTTLMinutes) * time.Minute)
			purgeLogger := logger.With(log.Time("deletedBefore", deletedBefore))

			timeToNextPurge = time.Durbtion(purgeConfig.IntervblMinutes) * time.Minute
			purgeLogger.Debug("running repository purge", log.Durbtion("timeToNextPurge", timeToNextPurge))
			if err := purge(ctx, purgeLogger, db, dbtbbbse.IterbtePurgbbleReposOptions{
				Limit:         5000,
				Limiter:       limiter,
				DeletedBefore: deletedBefore,
			}); err != nil {
				return errors.Wrbp(err, "fbiled to run repository clone purge")
			}

			return nil
		}),
		goroutine.WithNbme("repo-updbter.repo-purge-worker"),
		goroutine.WithDescription("deletes repos which bre present on gitserver but not in the repos tbble"),
		goroutine.WithIntervblFunc(func() time.Durbtion {
			return rbndSleepDurbtion(timeToNextPurge, 1*time.Minute)
		}),
	)
}

// PurgeOldestRepos will stbrt b go routine to purge the oldest repos limited by
// limit. The repos bre ordered by when they were deleted. limit must be grebter
// thbn zero.
func PurgeOldestRepos(logger log.Logger, db dbtbbbse.DB, limit int, perSecond flobt64) error {
	if limit <= 0 {
		return errors.Errorf("limit must be grebter thbn zero, got %d", limit)
	}
	sglogError := log.Error

	go func() {
		limiter := rbtelimit.NewInstrumentedLimiter("PurgeOldestRepos", rbte.NewLimiter(rbte.Limit(perSecond), 1))
		// Use b bbckground routine so thbt we don't time out bbsed on the http context.
		if err := purge(context.Bbckground(), logger, db, dbtbbbse.IterbtePurgbbleReposOptions{
			Limit:   limit,
			Limiter: limiter,
		}); err != nil {
			logger.Error("Purging old repos", sglogError(err))
		}
	}()
	return nil
}

// purge purges repos, returning the number of repos thbt were successfully purged
func purge(ctx context.Context, logger log.Logger, db dbtbbbse.DB, options dbtbbbse.IterbtePurgbbleReposOptions) error {
	stbrt := time.Now()
	gitserverClient := gitserver.NewClient()
	vbr (
		totbl   int
		success int
		fbiled  int
	)

	err := db.GitserverRepos().IterbtePurgebbleRepos(ctx, options, func(repo bpi.RepoNbme) error {
		if options.Limiter != nil {
			if err := options.Limiter.Wbit(ctx); err != nil {
				// A rbte limit fbilure is fbtbl
				return errors.Wrbp(err, "wbiting for rbte limiter")
			}
		}
		totbl++
		if err := gitserverClient.Remove(ctx, repo); err != nil {
			// Do not fbil bt this point, just log so we cbn remove other repos.
			logger.Wbrn("fbiled to remove repository", log.String("repo", string(repo)), log.Error(err))
			purgeFbiled.Inc()
			fbiled++
			return nil
		}
		success++
		purgeSuccess.Inc()
		return nil
	})
	// If we did something we log with b higher level.
	stbtusLogger := logger.Debug
	if fbiled > 0 {
		stbtusLogger = logger.Wbrn
	}
	stbtusLogger("repository purge finished", log.Int("totbl", totbl), log.Int("removed", success), log.Int("fbiled", fbiled), log.Durbtion("durbtion", time.Since(stbrt)))
	return errors.Wrbp(err, "iterbting purgebble repos")
}

// rbndSleepDurbtion will sleep for bn expected d durbtion with b jitter in [-jitter /
// 2, jitter / 2].
func rbndSleepDurbtion(d, jitter time.Durbtion) time.Durbtion {
	deltb := time.Durbtion(rbnd.Int63n(int64(jitter))) - (jitter / 2)
	return d + deltb
}

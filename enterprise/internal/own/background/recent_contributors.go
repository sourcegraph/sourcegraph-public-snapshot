package background

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/rcache"

	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func handleRecentContributors(ctx context.Context, lgr logger.Logger, repoId api.RepoID, db database.DB, subRepoPermsCache *rcache.Cache) error {
	// 🚨 SECURITY: we use the internal actor because the background indexer is not associated with any user, and needs
	// to see all repos and files
	internalCtx := actor.WithInternalActor(ctx)

	indexer := newRecentContributorsIndexer(gitserver.NewClient(), db, lgr, subRepoPermsCache)
	return indexer.indexRepo(internalCtx, repoId, authz.DefaultSubRepoPermsChecker)
}

type recentContributorsIndexer struct {
	client            gitserver.Client
	db                database.DB
	logger            logger.Logger
	subRepoPermsCache *rcache.Cache
}

func newRecentContributorsIndexer(client gitserver.Client, db database.DB, lgr logger.Logger, subRepoPermsCache *rcache.Cache) *recentContributorsIndexer {
	return &recentContributorsIndexer{client: client, db: db, logger: lgr, subRepoPermsCache: subRepoPermsCache}
}

var commitCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Name:      "own_recent_contributors_commits_indexed_total",
})

func (r *recentContributorsIndexer) indexRepo(ctx context.Context, repoId api.RepoID, checker authz.SubRepoPermissionChecker) error {
	// If the repo has sub-repo perms enabled, skip indexing
	isSubRepoPermsRepo, err := isSubRepoPermsRepo(ctx, repoId, r.subRepoPermsCache, checker)
	if err != nil {
		return err
	} else if isSubRepoPermsRepo {
		r.logger.Debug("skipping own contributor signal due to the repo having subrepo perms enabled")
		return nil
	}

	repoStore := r.db.Repos()
	repo, err := repoStore.Get(ctx, repoId)
	if err != nil {
		return errors.Wrap(err, "repoStore.Get")
	}
	commitLog, err := r.client.CommitLog(ctx, repo.Name, time.Now().AddDate(0, 0, -90))
	if err != nil {
		return errors.Wrap(err, "CommitLog")
	}

	store := r.db.RecentContributionSignals()
	err = store.ClearSignals(ctx, repoId)
	if err != nil {
		return errors.Wrap(err, "ClearSignals")
	}

	for _, commit := range commitLog {
		err := store.AddCommit(ctx, database.Commit{
			RepoID:       repoId,
			AuthorName:   commit.AuthorName,
			AuthorEmail:  commit.AuthorEmail,
			Timestamp:    commit.Timestamp,
			CommitSHA:    commit.SHA,
			FilesChanged: commit.ChangedFiles,
		})
		if err != nil {
			return errors.Wrapf(err, "AddCommit %v", commit)
		}
	}
	r.logger.Info("commits inserted", logger.Int("count", len(commitLog)), logger.Int("repo_id", int(repoId)))
	commitCounter.Add(float64(len(commitLog)))
	return nil
}

func isSubRepoPermsRepo(ctx context.Context, repoID api.RepoID, cache *rcache.Cache, checker authz.SubRepoPermissionChecker) (isSubRepoPermsRepo bool, err error) {
	repoIsCachedAndSubRepoPermsDisabled := false
	cacheKey := strconv.Itoa(int(repoID))
	noCache := false
	if cache != nil {
		val, ok := cache.Get(cacheKey)
		if ok {
			var isSubRepoPermsRepo bool
			if err := json.Unmarshal(val, &isSubRepoPermsRepo); err != nil {
				return false, err
			}
			if isSubRepoPermsRepo {
				return true, nil
			}
			repoIsCachedAndSubRepoPermsDisabled = true
		}
	} else {
		noCache = true
	}

	// No entry in cache, so we need to look up whether this is a sub-repo perms repo in the DB.
	if !repoIsCachedAndSubRepoPermsDisabled {
		ok, err := authz.SubRepoEnabledForRepoID(ctx, checker, repoID)
		if err != nil {
			return false, err
		}
		if ok {
			b, err := json.Marshal(true)
			if err != nil {
				return false, err
			}
			if !noCache {
				cache.Set(cacheKey, b)
			}
			return true, nil
		}
		b, err := json.Marshal(false)
		if err != nil {
			return false, err
		}
		if !noCache {
			cache.Set(cacheKey, b)
		}
	}
	return false, nil
}

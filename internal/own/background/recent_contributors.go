package background

import (
	"context"
	"io"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func handleRecentContributors(ctx context.Context, lgr logger.Logger, repoId api.RepoID, db database.DB) error {
	// 🚨 SECURITY: we use the internal actor because the background indexer is not associated with any user, and needs
	// to see all repos and files
	internalCtx := actor.WithInternalActor(ctx)

	indexer := newRecentContributorsIndexer(gitserver.NewClient("own.recentcontributors"), db, lgr)
	return indexer.indexRepo(internalCtx, repoId, authz.DefaultSubRepoPermsChecker)
}

type recentContributorsIndexer struct {
	client gitserver.Client
	db     database.DB
	logger logger.Logger
}

func newRecentContributorsIndexer(client gitserver.Client, db database.DB, lgr logger.Logger) *recentContributorsIndexer {
	return &recentContributorsIndexer{client: client, db: db, logger: lgr}
}

var commitCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Name:      "own_recent_contributors_commits_indexed_total",
})

func (r *recentContributorsIndexer) indexRepo(ctx context.Context, repoId api.RepoID, checker authz.SubRepoPermissionChecker) error {
	// If the repo has sub-repo perms enabled, skip indexing.
	isSubRepoPermsRepo, err := authz.SubRepoEnabledForRepoID(ctx, checker, repoId)
	if err != nil {
		return errcode.MakeNonRetryable(err)
	} else if isSubRepoPermsRepo {
		r.logger.Debug("skipping own contributor signal due to the repo having subrepo perms enabled", logger.Int32("repoID", int32(repoId)))
		return nil
	}

	repoStore := r.db.Repos()
	repo, err := repoStore.Get(ctx, repoId)
	if err != nil {
		return errors.Wrap(err, "repoStore.Get")
	}
	commits, err := r.client.Commits(ctx, repo.Name, gitserver.CommitsOptions{
		Order:  gitserver.CommitsOrderTopoDate,
		After:  time.Now().AddDate(0, 0, -90),
		Ranges: []string{"HEAD"},
	})
	if err != nil {
		return errors.Wrap(err, "Commits")
	}

	store := r.db.RecentContributionSignals()
	err = store.ClearSignals(ctx, repoId)
	if err != nil {
		return errors.Wrap(err, "ClearSignals")
	}

	inserted := 0
	for _, commit := range commits {
		if len(commit.Parents) > 1 { // We don't care about merge commits.
			continue
		}
		it, err := r.client.ChangedFiles(ctx, repo.Name, "", string(commit.ID))
		if err != nil {
			return errors.Wrap(err, "ChangedFiles")
		}
		changedFiles := []string{}
		for {
			ps, err := it.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return errors.Wrap(err, "ChangedFilesIterator")
			}
			changedFiles = append(changedFiles, ps.Path)
		}
		err = store.AddCommit(ctx, database.Commit{
			RepoID:       repoId,
			AuthorName:   commit.Author.Name,
			AuthorEmail:  commit.Author.Email,
			Timestamp:    commit.Author.Date,
			CommitSHA:    string(commit.ID),
			FilesChanged: changedFiles,
		})
		if err != nil {
			return errors.Wrapf(err, "AddCommit %v", commit)
		}
		inserted++
	}
	r.logger.Info("commits inserted", logger.Int("count", inserted), logger.Int("repo_id", int(repoId)))
	commitCounter.Add(float64(inserted))
	return nil
}

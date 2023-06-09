package background

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
)

func handleAnalytics(ctx context.Context, logger log.Logger, repoId api.RepoID, db database.DB) error {
	// 🚨 SECURITY: we use the internal actor because the background indexer is not associated with any user,
	// and needs to see all repos and files.
	internalCtx := actor.WithInternalActor(ctx)
	indexer := newAnalyticsIndexer(gitserver.NewClient(), db)
	err := indexer.indexRepo(internalCtx, repoId)
	if err != nil {
		logger.Error("own analytics indexing failure", log.String("msg", err.Error()))
	}
	return err
}

type analyticsIndexer struct {
	client gitserver.Client
	db     database.DB
}

func newAnalyticsIndexer(client gitserver.Client, db database.DB) *analyticsIndexer {
	return &analyticsIndexer{client: client, db: db}
}

var ownAnalyticsFilesCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Name:      "own_analytics_files_indexed_total",
})

func (r *analyticsIndexer) indexRepo(ctx context.Context, repoId api.RepoID) error {
	repoStore := r.db.Repos()
	repo, err := repoStore.Get(ctx, repoId)
	if err != nil {
		return errors.Wrap(err, "repoStore.Get")
	}
	files, err := r.client.LsFiles(ctx, nil, repo.Name, "HEAD")
	if err != nil {
		return errors.Wrap(err, "ls-files")
	}
	// Try to compute ownership stats
	ruleset, err := r.codeowners(ctx, repo)
	if err != nil {
		return err
	}
	var totalCount, codeownedCount int
	for _, f := range files {
		totalCount++
		if len(ruleset(f).GetOwner()) > 0 {
			codeownedCount++
		}
	}
	timestamp := time.Now()
	totalFileCountUpdate := rootPathIterator[int]{value: totalCount}
	rowCount, err := r.db.RepoPaths().UpdateFileCounts(ctx, repo.ID, totalFileCountUpdate, timestamp)
	if err != nil {
		return errors.Wrap(err, "UpdateFileCounts")
	}
	if rowCount == 0 {
		return errors.New("expected total file count updates")
	}
	codeownedCountUpdate := rootPathIterator[database.PathAggregateCounts]{
		value: database.PathAggregateCounts{CodeownedFileCount: codeownedCount},
	}
	rowCount, err = r.db.OwnershipStats().UpdateAggregateCounts(ctx, repo.ID, codeownedCountUpdate, timestamp)
	if err != nil {
		return errors.Wrap(err, "UpdateAggregateCounts")
	}
	if rowCount == 0 {
		return errors.New("expected CODEOWNERS-owned file count update")
	}
	ownAnalyticsFilesCounter.Add(float64(len(files)))
	return nil
}

// codeowners pulls a path matcher for repo HEAD.
// If result function is nil, then no CODEOWNERS file was found.
func (r *analyticsIndexer) codeowners(ctx context.Context, repo *types.Repo) (func(string) *codeownerspb.Rule, error) {
	ownService := own.NewService(r.client, r.db)
	commitID, err := r.client.ResolveRevision(ctx, repo.Name, "HEAD", gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot resolve HEAD")
	}
	ruleset, err := ownService.RulesetForRepo(ctx, repo.Name, repo.ID, commitID)
	if ruleset == nil || err != nil {
		// TODO(#53155): Return error in case there is an issue,
		// but return noRuleset and no error if CODEOWNERS is not found.
		return noRuleset, nil
	}
	return ruleset.Match, nil
}

// For proto it is safe to return nil from a function,
// since the implementation handles a nil reference gracefully.
// Just need to use getters instead of field access.
func noRuleset(string) *codeownerspb.Rule {
	return nil
}

type rootPathIterator[T any] struct {
	value T
}

func (i rootPathIterator[T]) Iterate(f func(path string, value T) error) error {
	return f("", i.value)
}

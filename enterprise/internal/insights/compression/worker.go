package compression

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbcache"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoStore interface {
	GetByName(ctx context.Context, repoName api.RepoName) ([]*types.Repo, error)
}

type CommitIndexer struct {
	db                database.DB
	limiter           *rate.Limiter
	allReposIterator  func(ctx context.Context, each func(repoName string, id api.RepoID) error) error
	getCommits        func(ctx context.Context, db database.DB, name api.RepoName, after time.Time, until *time.Time, operation *observation.Operation) ([]*gitdomain.Commit, error)
	commitStore       CommitStore
	maxHistoricalTime time.Time
	background        context.Context
	operations        *operations
	clock             func() time.Time
}

func NewCommitIndexer(background context.Context, base database.DB, insights dbutil.DB, clock func() time.Time, observationContext *observation.Context) *CommitIndexer {
	//TODO(insights): add a setting for historical index length
	startTime := time.Now().AddDate(-1, 0, 0)

	repoStore := database.Repos(base)

	commitStore := NewCommitStore(insights)

	iterator := discovery.NewAllReposIterator(
		dbcache.NewIndexableReposLister(repoStore),
		repoStore,
		time.Now,
		envvar.SourcegraphDotComMode(),
		15*time.Minute,
		&prometheus.CounterOpts{
			Namespace: "src",
			Name:      "insights_commit_index_repositories_analyzed",
			Help:      "Counter of the number of repositories analyzed in the commit indexer.",
		})

	limiter := rate.NewLimiter(10, 1)

	operations := newOperations(observationContext)

	indexer := CommitIndexer{
		db:                base,
		limiter:           limiter,
		allReposIterator:  iterator.ForEach,
		commitStore:       commitStore,
		maxHistoricalTime: startTime,
		background:        background,
		getCommits:        getCommits,
		operations:        operations,
		clock:             clock,
	}

	return &indexer
}

func NewCommitIndexerWorker(ctx context.Context, base database.DB, insights dbutil.DB, clock func() time.Time, observationContext *observation.Context) goroutine.BackgroundRoutine {
	indexer := NewCommitIndexer(ctx, base, insights, clock, observationContext)

	return indexer.Handler(ctx, observationContext)
}

func (i *CommitIndexer) Handler(ctx context.Context, observationContext *observation.Context) goroutine.BackgroundRoutine {
	intervalMinutes := conf.Get().InsightsCommitIndexerInterval
	if intervalMinutes <= 0 {
		intervalMinutes = 60
	}
	interval := time.Minute * time.Duration(intervalMinutes)

	return goroutine.NewPeriodicGoroutineWithMetrics(ctx, interval,
		goroutine.NewHandlerWithErrorMessage("commit_indexer_handler", func(ctx context.Context) error {
			return i.indexAll(ctx)
		}), i.operations.worker)
}

func (i *CommitIndexer) indexAll(ctx context.Context) error {
	err := i.allReposIterator(ctx, i.indexRepository)
	if err != nil {
		return err
	}

	return nil
}

// maxWindowsPerRepo Limits the number of windows of commits indexRepository can process per run
const maxWindowsPerRepo = 25

// indexRepository attempts to index the commits given a repository name one time window at a time.
// This method will absorb any errors that occur during execution and skip any remaining windows.
// If this repository already has some commits indexed, only commits made more recently than the previous index will be added.
func (i *CommitIndexer) indexRepository(name string, id api.RepoID) error {
	windowsProccssed := 0
	additionalWindows := true
	// It is important that the window size stays consistent during processing
	// so that it can correctly determine the time the repository has been indexed though
	windowDuration := conf.Get().InsightsCommitIndexerWindowDuration
	for additionalWindows && windowsProccssed < maxWindowsPerRepo {
		var err error
		additionalWindows, err = i.indexNextWindow(name, id, windowDuration)
		windowsProccssed++
		if err != nil {
			log15.Error(err.Error())
			return nil
		}
	}
	return nil

}

func (i *CommitIndexer) indexNextWindow(name string, id api.RepoID, windowDuration int) (moreWindows bool, err error) {
	ctx, cancel := context.WithTimeout(i.background, time.Second*45)
	defer cancel()

	err = i.limiter.Wait(ctx)
	if err != nil {
		return false, err
	}

	logger := log15.Root().New("worker", "insights-commit-indexer")

	repoName := api.RepoName(name)
	repoId := id

	metadata, err := getMetadata(ctx, repoId, i.commitStore)
	if err != nil {
		return false, errors.Wrapf(err, "unable to fetch commit index metadata repo_id: %v", repoId)
	}

	if !metadata.Enabled {
		logger.Debug("commit indexing disabled", "repo_id", repoId)
		return false, nil
	}

	searchStartTime := max(i.maxHistoricalTime, metadata.LastIndexedAt)
	commitLogRequestTime := i.clock().UTC()

	var searchEndTime *time.Time
	if windowDuration > 0 {
		endTime := searchStartTime.Add(time.Duration(24*windowDuration) * time.Hour)
		searchEndTime = &endTime
	}

	logger.Debug("fetching commits", "repo_id", repoId, "after", searchStartTime, "until", searchEndTime)
	commits, err := i.getCommits(ctx, i.db, repoName, searchStartTime, searchEndTime, i.operations.getCommits)
	if err != nil {
		return false, errors.Wrapf(err, "error fetching commits from gitserver repo_id: %v", repoId)
	}

	i.operations.countCommits.WithLabelValues().Add(float64(len(commits)))

	// default to thinking indexing is done
	indexedThrough := commitLogRequestTime // The time we issued the git log request
	moreWindows = false

	// If we are looking at a window of time determine if reached the end
	if searchEndTime != nil {
		moreWindows = searchEndTime.Before(commitLogRequestTime)
		if moreWindows {
			indexedThrough = *searchEndTime
		}
	}

	log15.Debug("indexing commits", "repo_id", repoId, "count", len(commits), "indexedThrough", indexedThrough)
	err = i.commitStore.InsertCommits(ctx, repoId, commits, indexedThrough, fmt.Sprintf("|repoName:%s|repoId:%d", repoName, repoId))
	if err != nil {
		return false, errors.Wrapf(err, "unable to update commit index repo_id: %v", repoId)
	}

	return moreWindows, nil
}

// getCommits fetches the commits from the remote gitserver for a repository after a certain time.
func getCommits(ctx context.Context, db database.DB, name api.RepoName, after time.Time, until *time.Time, operation *observation.Operation) (_ []*gitdomain.Commit, err error) {
	ctx, _, endObservation := operation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	before := ""
	if until != nil {
		before = until.Format(time.RFC3339)
	}

	return git.Commits(ctx, db, name, git.CommitsOptions{N: 0, DateOrder: true, NoEnsureRevision: true, After: after.Format(time.RFC3339), Before: before}, authz.DefaultSubRepoPermsChecker)
}

// getMetadata gets the index metadata for a repository. The metadata will be generated if it doesn't already exist, such as
// in the case of a newly installed repository.
func getMetadata(ctx context.Context, id api.RepoID, store CommitStore) (CommitIndexMetadata, error) {
	metadata, err := store.GetMetadata(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		metadata, err = store.UpsertMetadataStamp(ctx, id, time.Time{}.UTC())
	}
	if err != nil {
		return CommitIndexMetadata{}, err
	}

	return metadata, nil
}

func max(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

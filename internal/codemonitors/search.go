package codemonitors

import (
	"context"
	"sort"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Search(ctx context.Context, logger log.Logger, db database.DB, query string, monitorID int64) (_ []*result.CommitMatch, err error) {
	searchClient := client.New(logger, db, gitserver.NewClient("monitors.search"))
	inputs, err := searchClient.Plan(
		ctx,
		"V3",
		nil,
		query,
		search.Precise,
		search.Streaming,
	)
	if err != nil {
		return nil, errcode.MakeNonRetryable(err)
	}

	// Inline job creation so we can mutate the commit job before running it
	clients := searchClient.JobClients()
	planJob, err := jobutil.NewPlanJob(inputs, inputs.Plan)
	if err != nil {
		return nil, errcode.MakeNonRetryable(err)
	}

	hook := func(ctx context.Context, db database.DB, gs commit.GitserverClient, args *gitprotocol.SearchRequest, repoID api.RepoID, doSearch commit.DoSearchFunc) error {
		return hookWithID(ctx, db, gs, monitorID, repoID, args, doSearch)
	}
	planJob, err = addCodeMonitorHook(planJob, hook)
	if err != nil {
		return nil, errcode.MakeNonRetryable(err)
	}

	// Execute the search
	agg := streaming.NewAggregatingStream()
	_, err = planJob.Run(ctx, clients, agg)
	if err != nil {
		return nil, err
	}

	results := make([]*result.CommitMatch, len(agg.Results))
	for i, res := range agg.Results {
		cm, ok := res.(*result.CommitMatch)
		if !ok {
			return nil, errors.Errorf("expected search to only return commit matches, but got type %T", res)
		}
		results[i] = cm
	}

	return results, nil
}

// Snapshot runs a dummy search that just saves the current state of the searched repos in the database.
// On subsequent runs, this allows us to treat all new repos or sets of args as something new that should
// be searched from the beginning.
func Snapshot(ctx context.Context, logger log.Logger, db database.DB, query string) (map[api.RepoID][]string, error) {
	if db.Handle().InTransaction() {
		return nil, errors.New("Snapshot cannot be run in a transaction")
	}

	searchClient := client.New(logger, db, gitserver.NewClient("monitors.search.snapshot"))
	inputs, err := searchClient.Plan(
		ctx,
		"V3",
		nil,
		query,
		search.Precise,
		search.Streaming,
	)
	if err != nil {
		return nil, err
	}

	clients := searchClient.JobClients()
	planJob, err := jobutil.NewPlanJob(inputs, inputs.Plan)
	if err != nil {
		return nil, err
	}

	var (
		mu                sync.Mutex
		resolvedRevisions = make(map[api.RepoID][]string)
	)

	hook := func(ctx context.Context, db database.DB, gs commit.GitserverClient, args *gitprotocol.SearchRequest, repoID api.RepoID, _ commit.DoSearchFunc) error {
		// Resolve the requested revisions into a static set of commit hashes
		commitHashes, err := gs.ResolveRevisions(ctx, args.Repo, args.Revisions)
		if err != nil {
			return err
		}

		mu.Lock()
		resolvedRevisions[repoID] = commitHashes
		mu.Unlock()

		return nil
	}

	planJob, err = addCodeMonitorHook(planJob, hook)
	if err != nil {
		return nil, err
	}

	_, err = planJob.Run(ctx, clients, streaming.NewNullStream())
	if err != nil {
		return nil, err
	}

	return resolvedRevisions, nil
}

var ErrInvalidMonitorQuery = errors.New("code monitor cannot use different patterns for different repos")

func addCodeMonitorHook(in job.Job, hook commit.CodeMonitorHook) (_ job.Job, err error) {
	commitSearchJobCount := 0
	return job.Map(in, func(j job.Job) job.Job {
		switch v := j.(type) {
		case *commit.SearchJob:
			commitSearchJobCount++
			if commitSearchJobCount > 1 && err == nil {
				err = ErrInvalidMonitorQuery
			}
			cp := *v
			cp.CodeMonitorSearchWrapper = hook
			return &cp
		case *repos.ComputeExcludedJob, *jobutil.NoopJob:
			// ComputeExcludedJob is fine for code monitor jobs, but should be
			// removed since it's not used
			return jobutil.NewNoopJob()
		default:
			if len(j.Children()) == 0 {
				if err == nil {
					err = errors.New("all branches of query must be of type:diff or type:commit. If you have an AND/OR operator in your query, ensure that both sides have type:commit or type:diff.")
				}
			}
			return j
		}
	}), err
}

func hookWithID(
	ctx context.Context,
	db database.DB,
	gs commit.GitserverClient,
	monitorID int64,
	repoID api.RepoID,
	args *gitprotocol.SearchRequest,
	doSearch commit.DoSearchFunc,
) error {
	cm := db.CodeMonitors()

	// Resolve the requested revisions into a static set of commit hashes
	commitHashes, err := gs.ResolveRevisions(ctx, args.Repo, args.Revisions)
	if err != nil {
		return err
	}

	// Look up the previously searched set of commit hashes
	lastSearched, err := cm.GetLastSearched(ctx, monitorID, repoID)
	if err != nil {
		return err
	}
	if stringsEqual(commitHashes, lastSearched) {
		// Early return if the repo hasn't changed since last search
		return nil
	}

	// Merge requested hashes and excluded hashes
	newRevs := make([]gitprotocol.RevisionSpecifier, 0, len(commitHashes)+len(lastSearched))
	for _, hash := range commitHashes {
		newRevs = append(newRevs, gitprotocol.RevisionSpecifier{RevSpec: hash})
	}
	for _, exclude := range lastSearched {
		newRevs = append(newRevs, gitprotocol.RevisionSpecifier{RevSpec: "^" + exclude})
	}

	// Update args with the new set of revisions
	argsCopy := *args
	argsCopy.Revisions = newRevs

	// Execute the search
	searchErr := doSearch(&argsCopy)

	// NOTE(camdencheek): we want to always save the "last searched" commits
	// because if we stream results, the user will get a notification for them
	// whether or not there was an error and forcing a re-search will cause the
	// user to get repeated notifications for the same commits. This makes code
	// monitors look very broken, and should be avoided.
	upsertErr := cm.UpsertLastSearched(ctx, monitorID, repoID, commitHashes)
	if upsertErr != nil {
		return upsertErr
	}

	// Still return the error so it can be displayed to the user
	if searchErr != nil {
		return errors.Wrap(searchErr, "search failed, some commits may be skipped")
	}

	return nil
}

func stringsEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	sort.Strings(left)
	sort.Strings(right)

	for i := range left {
		if right[i] != left[i] {
			return false
		}
	}
	return true
}

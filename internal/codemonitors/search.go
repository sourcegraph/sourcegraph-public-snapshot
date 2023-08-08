package codemonitors

import (
	"context"
	"net/url"
	"sort"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
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
	searchClient := client.New(logger, db)
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
		return hookWithID(ctx, db, logger, gs, monitorID, repoID, args, doSearch)
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
func Snapshot(ctx context.Context, logger log.Logger, db database.DB, query string, monitorID int64) error {
	searchClient := client.New(logger, db)
	inputs, err := searchClient.Plan(
		ctx,
		"V3",
		nil,
		query,
		search.Precise,
		search.Streaming,
	)
	if err != nil {
		return err
	}

	clients := searchClient.JobClients()
	planJob, err := jobutil.NewPlanJob(inputs, inputs.Plan)
	if err != nil {
		return err
	}

	hook := func(ctx context.Context, db database.DB, gs commit.GitserverClient, args *gitprotocol.SearchRequest, repoID api.RepoID, _ commit.DoSearchFunc) error {
		return snapshotHook(ctx, db, gs, args, monitorID, repoID)
	}

	planJob, err = addCodeMonitorHook(planJob, hook)
	if err != nil {
		return err
	}

	// HACK(camdencheek): limit the concurrency of the commit search job
	// because the db passed into this function might actually be a transaction
	// and transactions cannot be used concurrently.
	planJob = limitConcurrency(planJob)

	_, err = planJob.Run(ctx, clients, streaming.NewNullStream())
	return err
}

var ErrInvalidMonitorQuery = errors.New("code monitor cannot use different patterns for different repos")

func limitConcurrency(in job.Job) job.Job {
	return job.Map(in, func(j job.Job) job.Job {
		switch v := j.(type) {
		case *commit.SearchJob:
			cp := *v
			cp.Concurrency = 1
			return &cp
		default:
			return j
		}
	})
}

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
	logger log.Logger,
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
	err = doSearch(&argsCopy)
	if err != nil {
		if errors.IsContextError(err) {
			logger.Warn(
				"commit search timed out, some commits may have been skipped",
				log.Error(err),
				log.String("repo", string(args.Repo)),
				log.Strings("include", commitHashes),
				log.Strings("exlcude", lastSearched),
			)
		} else {
			return err
		}
	}

	// If the search was successful, store the resolved hashes
	// as the new "last searched" hashes
	return cm.UpsertLastSearched(ctx, monitorID, repoID, commitHashes)
}

func snapshotHook(
	ctx context.Context,
	db database.DB,
	gs commit.GitserverClient,
	args *gitprotocol.SearchRequest,
	monitorID int64,
	repoID api.RepoID,
) error {
	cm := db.CodeMonitors()

	// Resolve the requested revisions into a static set of commit hashes
	commitHashes, err := gs.ResolveRevisions(ctx, args.Repo, args.Revisions)
	if err != nil {
		return err
	}

	return cm.UpsertLastSearched(ctx, monitorID, repoID, commitHashes)
}

func gqlURL(queryName string) (string, error) {
	u, err := url.Parse(internalapi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Path = "/.internal/graphql"
	u.RawQuery = queryName
	return u.String(), nil
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

package codemonitors

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"

	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const gqlSettingsQuery = `query CodeMonitorSettings{
	viewerSettings {
		final
	}
}`

type gqlSettingsResponse struct {
	Data struct {
		ViewerSettings struct {
			Final string `json:"final"`
		} `json:"viewerSettings"`
	} `json:"data"`
	Errors []gqlerrors.FormattedError
}

// Settings queries for the computed Settings for the current actor
func Settings(ctx context.Context) (_ *schema.Settings, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CodeMonitorSearch")
	defer func() {
		span.LogFields(otlog.Error(err))
		span.Finish()
	}()

	reqBody, err := json.Marshal(map[string]any{"query": gqlSettingsQuery})
	if err != nil {
		return nil, errors.Wrap(err, "marshal request body")
	}

	urlStr, err := gqlURL("CodeMonitorSettings")
	if err != nil {
		return nil, errors.Wrap(err, "construct frontend URL")
	}

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "construct request")
	}
	req.Header.Set("Content-Type", "application/json")
	if span != nil {
		carrier := opentracing.HTTPHeadersCarrier(req.Header)
		span.Tracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			carrier,
		)
	}

	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	var res gqlSettingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	if len(res.Errors) > 0 {
		var combined error
		for _, err := range res.Errors {
			combined = errors.Append(combined, err)
		}
		return nil, combined
	}

	var unmarshaledSettings schema.Settings
	if err := json.Unmarshal([]byte(res.Data.ViewerSettings.Final), &unmarshaledSettings); err != nil {
		return nil, err
	}
	return &unmarshaledSettings, nil
}

func Search(ctx context.Context, logger log.Logger, db database.DB, query string, monitorID int64, settings *schema.Settings) (_ []*result.CommitMatch, err error) {
	searchClient := client.NewSearchClient(logger, db, search.Indexed(), search.SearcherURLs())
	inputs, err := searchClient.Plan(
		ctx,
		"V3",
		nil,
		query,
		search.Precise,
		search.Streaming,
		settings,
		envvar.SourcegraphDotComMode(),
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

	if featureflag.FromContext(ctx).GetBoolOr("cc-repo-aware-monitors", true) {
		hook := func(ctx context.Context, db database.DB, gs commit.GitserverClient, args *gitprotocol.SearchRequest, repoID api.RepoID, doSearch commit.DoSearchFunc) error {
			return hookWithID(ctx, db, logger, gs, monitorID, repoID, args, doSearch)
		}
		{
			// This block is a transitional block that can be removed in a
			// future version. We need this block to exist when we transition
			// from timestamp-based code monitors to commit-hash-based
			// (repo-aware) code monitors.
			//
			// When we flip the switch to repo-aware, all existing code
			// monitors will start executing as repo-aware code monitors.
			// Without this block, this  would mean all repos would be detected
			// as "unsearched" and we would start searching from the beginning
			// of the repo's history, flooding every code monitor user with a
			// notification with many results.
			//
			// Instead, this detects if this monitor has ever been run as a
			// repo-aware monitor before and snapshots the current state of the
			// searched repos rather than searching them.
			hasAnyLastSearched, err := edb.NewEnterpriseDB(db).CodeMonitors().HasAnyLastSearched(ctx, monitorID)
			if err != nil {
				return nil, err
			} else if !hasAnyLastSearched {
				hook = func(ctx context.Context, db database.DB, gs commit.GitserverClient, args *gitprotocol.SearchRequest, repoID api.RepoID, _ commit.DoSearchFunc) error {
					return snapshotHook(ctx, db, gs, args, monitorID, repoID)
				}
			}
		}
		planJob, err = addCodeMonitorHook(planJob, hook)
		if err != nil {
			return nil, errcode.MakeNonRetryable(err)
		}
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
func Snapshot(ctx context.Context, logger log.Logger, db database.DB, query string, monitorID int64, settings *schema.Settings) error {
	searchClient := client.NewSearchClient(logger, db, search.Indexed(), search.SearcherURLs())
	inputs, err := searchClient.Plan(
		ctx,
		"V3",
		nil,
		query,
		search.Precise,
		search.Streaming,
		settings,
		envvar.SourcegraphDotComMode(),
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
					err = errors.Errorf("found invalid atom job type %T for code monitor search", j)
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
	cm := edb.NewEnterpriseDB(db).CodeMonitors()

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
	cm := edb.NewEnterpriseDB(db).CodeMonitors()

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

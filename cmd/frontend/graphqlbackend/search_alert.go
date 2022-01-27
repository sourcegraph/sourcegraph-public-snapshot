package graphqlbackend

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type searchAlert struct {
	alert *search.Alert
}

func NewSearchAlertResolver(alert *search.Alert) searchAlert {
	return searchAlert{alert: alert}
}

func (a searchAlert) Title() string { return a.alert.Title }

func (a searchAlert) Description() *string {
	if a.alert.Description == "" {
		return nil
	}
	return &a.alert.Description
}

func (a searchAlert) PrometheusType() string {
	return a.alert.PrometheusType
}

func (a searchAlert) ProposedQueries() *[]*searchQueryDescription {
	if len(a.alert.ProposedQueries) == 0 {
		return nil
	}
	var proposedQueries []*searchQueryDescription
	for _, q := range a.alert.ProposedQueries {
		proposedQueries = append(proposedQueries, &searchQueryDescription{query: q})
	}
	return &proposedQueries
}

// reposExist returns true if one or more repos resolve. If the attempt
// returns 0 repos or fails, it returns false. It is a helper function for
// raising NoResolvedRepos alerts with suggestions when we know the original
// query does not contain any repos to search.
func (o *alertObserver) reposExist(ctx context.Context, options search.RepoOptions) bool {
	options.UserSettings = o.UserSettings
	repositoryResolver := &searchrepos.Resolver{DB: o.Db}
	resolved, err := repositoryResolver.Resolve(ctx, options)
	return err == nil && len(resolved.RepoRevs) > 0
}

func (o *alertObserver) alertForNoResolvedRepos(ctx context.Context, q query.Q) *searchAlert {
	repoFilters, minusRepoFilters := q.Repositories()
	contextFilters, _ := q.StringValues(query.FieldContext)
	onlyForks, noForks, forksNotSet := false, false, true
	if fork := q.Fork(); fork != nil {
		onlyForks = *fork == query.Only
		noForks = *fork == query.No
		forksNotSet = false
	}
	archived := q.Archived()
	archivedNotSet := archived == nil

	if len(repoFilters) == 0 && len(minusRepoFilters) == 0 {
		return &searchAlert{
			alert: &search.Alert{
				PrometheusType: "no_resolved_repos__no_repositories",
				Title:          "Add repositories or connect repository hosts",
				Description:    "There are no repositories to search. Add an external service connection to your code host.",
			},
		}
	}
	if len(contextFilters) == 1 && !searchcontexts.IsGlobalSearchContextSpec(contextFilters[0]) && len(repoFilters) > 0 {
		withoutContextFilter := query.OmitField(q, query.FieldContext)
		proposedQueries := []*search.ProposedQuery{
			search.NewProposedQuery(
				"search in the global context",
				fmt.Sprintf("context:%s %s", searchcontexts.GlobalSearchContextName, withoutContextFilter),
				o.PatternType,
			),
		}

		return &searchAlert{
			alert: &search.Alert{
				PrometheusType:  "no_resolved_repos__context_none_in_common",
				Title:           fmt.Sprintf("No repositories found for your query within the context %s", contextFilters[0]),
				ProposedQueries: proposedQueries,
			},
		}
	}

	isSiteAdmin := backend.CheckCurrentUserIsSiteAdmin(ctx, o.Db) == nil
	if !envvar.SourcegraphDotComMode() {
		if needsRepoConfig, err := needsRepositoryConfiguration(ctx, o.Db); err == nil && needsRepoConfig {
			if isSiteAdmin {
				return &searchAlert{
					alert: &search.Alert{
						Title:       "No repositories or code hosts configured",
						Description: "To start searching code, first go to site admin to configure repositories and code hosts.",
					},
				}
			} else {
				return &searchAlert{
					alert: &search.Alert{
						Title:       "No repositories or code hosts configured",
						Description: "To start searching code, ask the site admin to configure and enable repositories.",
					},
				}
			}
		}
	}

	proposedQueries := []*search.ProposedQuery{}
	if forksNotSet {
		tryIncludeForks := search.RepoOptions{
			RepoFilters:      repoFilters,
			MinusRepoFilters: minusRepoFilters,
			NoForks:          false,
		}
		if o.reposExist(ctx, tryIncludeForks) {
			proposedQueries = append(proposedQueries,
				search.NewProposedQuery(
					"include forked repositories in your query.",
					o.OriginalQuery+" fork:yes",
					o.PatternType,
				),
			)
		}
	}

	if archivedNotSet {
		tryIncludeArchived := search.RepoOptions{
			RepoFilters:      repoFilters,
			MinusRepoFilters: minusRepoFilters,
			OnlyForks:        onlyForks,
			NoForks:          noForks,
			OnlyArchived:     true,
		}
		if o.reposExist(ctx, tryIncludeArchived) {
			proposedQueries = append(proposedQueries,
				search.NewProposedQuery(
					"include archived repositories in your query.",
					o.OriginalQuery+" archived:yes",
					o.PatternType,
				),
			)
		}
	}

	if len(proposedQueries) > 0 {
		return &searchAlert{
			alert: &search.Alert{
				PrometheusType:  "no_resolved_repos__repos_exist_when_altered",
				Title:           "No repositories found",
				Description:     "Try alter the query or use a different `repo:<regexp>` filter to see results",
				ProposedQueries: proposedQueries,
			},
		}
	}

	return &searchAlert{
		alert: &search.Alert{
			PrometheusType: "no_resolved_repos__generic",
			Title:          "No repositories found",
			Description:    "Try using a different `repo:<regexp>` filter to see results",
		},
	}
}

type errOverRepoLimit struct {
	ProposedQueries []*search.ProposedQuery
	Description     string
}

func (e *errOverRepoLimit) Error() string {
	return "Too many matching repositories"
}

func (a searchAlert) wrapResults() *SearchResults {
	return &SearchResults{Alert: &a}
}

func (a searchAlert) wrapSearchImplementer(db database.DB) *alertSearchImplementer {
	return &alertSearchImplementer{
		db:    db,
		alert: a,
	}
}

// alertSearchImplementer is a light wrapper type around an alert that implements
// SearchImplementer. This helps avoid needing to have a db on the searchAlert type
type alertSearchImplementer struct {
	db    database.DB
	alert searchAlert
}

func (a alertSearchImplementer) Results(context.Context) (*SearchResultsResolver, error) {
	return &SearchResultsResolver{db: a.db, SearchResults: a.alert.wrapResults()}, nil
}

func (alertSearchImplementer) Stats(context.Context) (*searchResultsStats, error) { return nil, nil }
func (alertSearchImplementer) Inputs() run.SearchInputs {
	return run.SearchInputs{}
}

func (o *alertObserver) errorToAlert(ctx context.Context, err error) (*searchAlert, error) {
	if err == nil {
		return nil, nil
	}

	var e *multierror.Error
	if errors.As(err, &e) {
		return o.multierrorToAlert(ctx, e)
	}

	var (
		rErr *commit.RepoLimitError
		tErr *commit.TimeLimitError
		mErr *searchrepos.MissingRepoRevsError
		oErr *errOverRepoLimit
	)

	if errors.HasType(err, authz.ErrStalePermissions{}) {
		return &searchAlert{alert: search.AlertForStalePermissions()}, nil
	}

	{
		var e gitdomain.BadCommitError
		if errors.As(err, &e) {
			return &searchAlert{alert: search.AlertForInvalidRevision(e.Spec)}, nil
		}
	}

	if err == searchrepos.ErrNoResolvedRepos {
		return o.alertForNoResolvedRepos(ctx, o.Query), nil
	}

	if errors.As(err, &oErr) {
		return &searchAlert{
			alert: &search.Alert{
				PrometheusType:  "over_repo_limit",
				Title:           "Too many matching repositories",
				ProposedQueries: oErr.ProposedQueries,
				Description:     oErr.Description,
			},
		}, nil
	}

	if errors.As(err, &mErr) {
		a := &searchAlert{alert: search.AlertForMissingRepoRevs(mErr.Missing)}
		a.alert.Priority = 6
		return a, nil
	}

	if strings.Contains(err.Error(), "Worker_oomed") || strings.Contains(err.Error(), "Worker_exited_abnormally") {
		return &searchAlert{
			alert: &search.Alert{
				PrometheusType: "structural_search_needs_more_memory",
				Title:          "Structural search needs more memory",
				Description:    "Running your structural search may require more memory. If you are running the query on many repositories, try reducing the number of repositories with the `repo:` filter.",
				Priority:       5,
			},
		}, nil
	}

	if strings.Contains(err.Error(), "Out of memory") {
		return &searchAlert{
			alert: &search.Alert{
				PrometheusType: "structural_search_needs_more_memory__give_searcher_more_memory",
				Title:          "Structural search needs more memory",
				Description:    `Running your structural search requires more memory. You could try reducing the number of repositories with the "repo:" filter. If you are an administrator, try double the memory allocated for the "searcher" service. If you're unsure, reach out to us at support@sourcegraph.com.`,
				Priority:       4,
			},
		}, nil
	}

	if errors.As(err, &rErr) {
		return &searchAlert{
			alert: &search.Alert{
				PrometheusType: "exceeded_diff_commit_search_limit",
				Title:          fmt.Sprintf("Too many matching repositories for %s search to handle", rErr.ResultType),
				Description:    fmt.Sprintf(`%s search can currently only handle searching across %d repositories at a time. Try using the "repo:" filter to narrow down which repositories to search, or using 'after:"1 week ago"'.`, strings.Title(rErr.ResultType), rErr.Max),
				Priority:       2,
			},
		}, nil
	}

	if errors.As(err, &tErr) {
		return &searchAlert{
			alert: &search.Alert{
				PrometheusType: "exceeded_diff_commit_with_time_search_limit",
				Title:          fmt.Sprintf("Too many matching repositories for %s search to handle", tErr.ResultType),
				Description:    fmt.Sprintf(`%s search can currently only handle searching across %d repositories at a time. Try using the "repo:" filter to narrow down which repositories to search.`, strings.Title(tErr.ResultType), tErr.Max),
				Priority:       1,
			},
		}, nil
	}

	return nil, err
}

func maxAlertByPriority(a, b *searchAlert) *searchAlert {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	if a.alert.Priority < b.alert.Priority {
		return b
	}

	return a
}

// multierrorToAlert converts a multierror.Error into the highest priority alert
// for the errors contained in it, and a new error with all the errors that could
// not be converted to alerts.
func (o *alertObserver) multierrorToAlert(ctx context.Context, me *multierror.Error) (resAlert *searchAlert, resErr error) {
	for _, err := range me.Errors {
		alert, err := o.errorToAlert(ctx, err)
		resAlert = maxAlertByPriority(resAlert, alert)
		resErr = multierror.Append(resErr, err)
	}

	return resAlert, resErr
}

type alertObserver struct {
	Db database.DB

	// Inputs are used to generate alert messages based on the query.
	*run.SearchInputs

	// Update state.
	hasResults bool

	// Error state. Can be called concurrently.
	mu    sync.Mutex
	alert *searchAlert
	err   error
}

func (o *alertObserver) Error(ctx context.Context, err error) {
	// Timeouts are reported through Stats so don't report an error for them.
	if err == nil || isContextError(ctx, err) {
		return
	}

	// We can compute the alert outside of the critical section.
	alert, _ := o.errorToAlert(ctx, err)

	o.mu.Lock()
	defer o.mu.Unlock()

	// The error can be converted into an alert.
	if alert != nil {
		o.update(alert)
		return
	}

	// Track the unexpected error for reporting when calling Done.
	o.err = multierror.Append(o.err, err)
}

// update to alert if it is more important than our current alert.
func (o *alertObserver) update(alert *searchAlert) {
	if o.alert == nil || o.alert.alert == nil || alert.alert.Priority > o.alert.alert.Priority {
		o.alert = alert
	}
}

//  Done returns the highest priority alert and a multierror.Error containing
//  all errors that could not be converted to alerts.
func (o *alertObserver) Done(stats *streaming.Stats) (*searchAlert, error) {
	if !o.hasResults && o.PatternType != query.SearchTypeStructural && comby.MatchHoleRegexp.MatchString(o.OriginalQuery) {
		o.update(&searchAlert{alert: search.AlertForStructuralSearchNotSet(o.OriginalQuery)})
	}

	if o.hasResults && o.err != nil {
		log15.Error("Errors during search", "error", o.err)
		return o.alert, nil
	}

	return o.alert, o.err
}

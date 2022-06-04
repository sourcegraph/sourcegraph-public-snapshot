package alert

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Observer struct {
	Db database.DB

	// Inputs are used to generate alert messages based on the query.
	*run.SearchInputs

	// Update state.
	HasResults bool

	// Error state. Can be called concurrently.
	mu    sync.Mutex
	alert *search.Alert
	err   error
}

// reposExist returns true if one or more repos resolve. If the attempt
// returns 0 repos or fails, it returns false. It is a helper function for
// raising NoResolvedRepos alerts with suggestions when we know the original
// query does not contain any repos to search.
func (o *Observer) reposExist(ctx context.Context, options search.RepoOptions) bool {
	repositoryResolver := &searchrepos.Resolver{DB: o.Db}
	resolved, err := repositoryResolver.Resolve(ctx, options)
	return err == nil && len(resolved.RepoRevs) > 0
}

func (o *Observer) alertForNoResolvedRepos(ctx context.Context, q query.Q) *search.Alert {
	repoFilters, minusRepoFilters := q.Repositories()
	contextFilters, _ := q.StringValues(query.FieldContext)
	dependencies := q.Dependencies()
	onlyForks, noForks, forksNotSet := false, false, true
	if fork := q.Fork(); fork != nil {
		onlyForks = *fork == query.Only
		noForks = *fork == query.No
		forksNotSet = false
	}
	archived := q.Archived()
	archivedNotSet := archived == nil

	if len(contextFilters) == 1 && !searchcontexts.IsGlobalSearchContextSpec(contextFilters[0]) && len(repoFilters) > 0 {
		withoutContextFilter := query.OmitField(q, query.FieldContext)
		proposedQueries := []*search.ProposedQuery{
			{
				Description: "search in the global context",
				Query:       fmt.Sprintf("context:%s %s", searchcontexts.GlobalSearchContextName, withoutContextFilter),
				PatternType: o.PatternType,
			},
		}

		return &search.Alert{
			PrometheusType:  "no_resolved_repos__context_none_in_common",
			Title:           fmt.Sprintf("No repositories found for your query within the context %s", contextFilters[0]),
			ProposedQueries: proposedQueries,
		}
	}

	isSiteAdmin := backend.CheckCurrentUserIsSiteAdmin(ctx, o.Db) == nil
	if !envvar.SourcegraphDotComMode() {
		if len(dependencies) > 0 {
			needsPackageHostConfig, err := needsPackageHostConfiguration(ctx, o.Db)
			if err == nil && needsPackageHostConfig {
				if isSiteAdmin {
					return &search.Alert{
						Title:       "No package hosts configured",
						Description: "To start searching your dependencies, first go to site admin to configure package hosts.",
					}
				} else {
					return &search.Alert{
						Title:       "No package hosts configured",
						Description: "To start searching your dependencies, ask the site admin to configure package hosts.",
					}
				}
			}
		}

		if needsRepoConfig, err := needsRepositoryConfiguration(ctx, o.Db); err == nil && needsRepoConfig {
			if isSiteAdmin {
				return &search.Alert{
					Title:       "No repositories or code hosts configured",
					Description: "To start searching code, first go to site admin to configure repositories and code hosts.",
				}
			} else {
				return &search.Alert{
					Title:       "No repositories or code hosts configured",
					Description: "To start searching code, ask the site admin to configure and enable repositories.",
				}
			}
		}
	}

	if len(dependencies) > 0 {
		return &search.Alert{
			Title:       "No dependency repositories found",
			Description: "Dependency repos are cloned on-demand when first searched. Try again in a few seconds if you know the given repositories have dependencies.\n\nRead more about dependencies search [here](https://docs.sourcegraph.com/code_search/how-to/dependencies_search).",
		}
	}

	var proposedQueries []*search.ProposedQuery
	if forksNotSet {
		tryIncludeForks := search.RepoOptions{
			RepoFilters:      repoFilters,
			MinusRepoFilters: minusRepoFilters,
			NoForks:          false,
		}
		if o.reposExist(ctx, tryIncludeForks) {
			proposedQueries = append(proposedQueries,
				&search.ProposedQuery{
					Description: "include forked repositories in your query.",
					Query:       o.OriginalQuery + " fork:yes",
					PatternType: o.PatternType,
				},
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
				&search.ProposedQuery{
					Description: "include archived repositories in your query.",
					Query:       o.OriginalQuery + " archived:yes",
					PatternType: o.PatternType,
				},
			)
		}
	}

	if len(proposedQueries) > 0 {
		return &search.Alert{
			PrometheusType:  "no_resolved_repos__repos_exist_when_altered",
			Title:           "No repositories found",
			Description:     "Try altering the query or use a different `repo:<regexp>` filter to see results",
			ProposedQueries: proposedQueries,
		}
	}

	return &search.Alert{
		PrometheusType: "no_resolved_repos__generic",
		Title:          "No repositories found",
		Description:    "Try using a different `repo:<regexp>` filter to see results",
	}
}

// multierrorToAlert converts an error.MultiError into the highest priority alert
// for the errors contained in it, and a new error with all the errors that could
// not be converted to alerts.
func (o *Observer) multierrorToAlert(ctx context.Context, me errors.MultiError) (resAlert *search.Alert, resErr error) {
	for _, err := range me.Errors() {
		alert, err := o.errorToAlert(ctx, err)
		resAlert = maxAlertByPriority(resAlert, alert)
		resErr = errors.Append(resErr, err)
	}

	return resAlert, resErr
}

func (o *Observer) Error(ctx context.Context, err error) {
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
	o.err = errors.Append(o.err, err)
}

// update to alert if it is more important than our current alert.
func (o *Observer) update(alert *search.Alert) {
	if o.alert == nil || alert.Priority > o.alert.Priority {
		o.alert = alert
	}
}

// Done returns the highest priority alert and an error.MultiError containing
// all errors that could not be converted to alerts.
func (o *Observer) Done() (*search.Alert, error) {
	if !o.HasResults && o.PatternType != query.SearchTypeStructural && comby.MatchHoleRegexp.MatchString(o.OriginalQuery) {
		o.update(search.AlertForStructuralSearchNotSet(o.OriginalQuery))
	}

	if o.HasResults && o.err != nil {
		log15.Error("Errors during search", "error", o.err)
		return o.alert, nil
	}

	return o.alert, o.err
}

func (o *Observer) errorToAlert(ctx context.Context, err error) (*search.Alert, error) {
	if err == nil {
		return nil, nil
	}

	var e errors.MultiError
	if errors.As(err, &e) {
		return o.multierrorToAlert(ctx, e)
	}

	var (
		mErr *searchrepos.MissingRepoRevsError
		oErr *errOverRepoLimit
	)

	if errors.HasType(err, authz.ErrStalePermissions{}) {
		return search.AlertForStalePermissions(), nil
	}

	{
		var e gitdomain.BadCommitError
		if errors.As(err, &e) {
			return search.AlertForInvalidRevision(e.Spec), nil
		}
	}

	if !o.HasResults && errors.Is(err, searchrepos.ErrNoResolvedRepos) {
		return o.alertForNoResolvedRepos(ctx, o.Query), nil
	}

	if errors.As(err, &oErr) {
		return &search.Alert{
			PrometheusType:  "over_repo_limit",
			Title:           "Too many matching repositories",
			ProposedQueries: oErr.ProposedQueries,
			Description:     oErr.Description,
		}, nil
	}

	if errors.As(err, &mErr) {
		var a *search.Alert
		dependencies := o.Query.Dependencies()
		if len(dependencies) == 0 {
			a = search.AlertForMissingRepoRevs(mErr.Missing)
		} else {
			a = search.AlertForMissingDependencyRepoRevs(mErr.Missing)
		}
		a.Priority = 6
		return a, nil
	}

	if strings.Contains(err.Error(), "Worker_oomed") || strings.Contains(err.Error(), "Worker_exited_abnormally") {
		return &search.Alert{
			PrometheusType: "structural_search_needs_more_memory",
			Title:          "Structural search needs more memory",
			Description:    "Running your structural search may require more memory. If you are running the query on many repositories, try reducing the number of repositories with the `repo:` filter.",
			Priority:       5,
		}, nil
	}

	if strings.Contains(err.Error(), "Out of memory") {
		return &search.Alert{
			PrometheusType: "structural_search_needs_more_memory__give_searcher_more_memory",
			Title:          "Structural search needs more memory",
			Description:    `Running your structural search requires more memory. You could try reducing the number of repositories with the "repo:" filter. If you are an administrator, try double the memory allocated for the "searcher" service. If you're unsure, reach out to us at support@sourcegraph.com.`,
			Priority:       4,
		}, nil
	}

	return nil, err
}

func maxAlertByPriority(a, b *search.Alert) *search.Alert {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	if a.Priority < b.Priority {
		return b
	}

	return a
}

func needsRepositoryConfiguration(ctx context.Context, db database.DB) (bool, error) {
	kinds := make([]string, 0, len(database.ExternalServiceKinds))
	for kind, config := range database.ExternalServiceKinds {
		if config.CodeHost {
			kinds = append(kinds, kind)
		}
	}

	count, err := db.ExternalServices().Count(ctx, database.ExternalServicesListOptions{
		Kinds: kinds,
	})
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func needsPackageHostConfiguration(ctx context.Context, db database.DB) (bool, error) {
	count, err := db.ExternalServices().Count(ctx, database.ExternalServicesListOptions{
		Kinds: []string{
			extsvc.KindNpmPackages,
			extsvc.KindGoModules,
		},
	})
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

type errOverRepoLimit struct {
	ProposedQueries []*search.ProposedQuery
	Description     string
}

func (e *errOverRepoLimit) Error() string {
	return "Too many matching repositories"
}

// isContextError returns true if ctx.Err() is not nil or if err
// is an error caused by context cancelation or timeout.
func isContextError(ctx context.Context, err error) bool {
	return ctx.Err() != nil || errors.IsAny(err, context.Canceled, context.DeadlineExceeded)
}

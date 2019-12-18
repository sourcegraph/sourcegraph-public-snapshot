package graphqlbackend

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/zoekt"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/search/search"
)

func (r *schemaResolver) Repositories(args *struct {
	graphqlutil.ConnectionArgs
	Query           *string
	Names           *[]string
	Enabled         bool // deprecated
	Disabled        bool // deprecated
	Cloned          bool
	CloneInProgress bool
	NotCloned       bool
	Indexed         bool
	NotIndexed      bool
	OrderBy         string
	Descending      bool
}) (*repositoryConnectionResolver, error) {
	// New call sites don't specify Enable and Disable. Assume if disabled
	// isn't specified we want Enabled since all repos are enabled.
	if !args.Disabled {
		args.Enabled = true
	}

	opt := db.ReposListOptions{
		Enabled: args.Enabled,
		OrderBy: db.RepoListOrderBy{{
			Field:      toDBRepoListColumn(args.OrderBy),
			Descending: args.Descending,
		}},
	}
	if args.Names != nil {
		// Make an exact-match regexp for each name.
		patterns := make([]string, len(*args.Names))
		for i, name := range *args.Names {
			patterns[i] = regexp.QuoteMeta(name)
		}
		opt.IncludePatterns = []string{"^(" + strings.Join(patterns, "|") + ")$"}
	}
	if args.Query != nil {
		opt.Query = *args.Query
	}
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &repositoryConnectionResolver{
		opt:             opt,
		cloned:          args.Cloned,
		cloneInProgress: args.CloneInProgress,
		notCloned:       args.NotCloned,
		indexed:         args.Indexed,
		notIndexed:      args.NotIndexed,
	}, nil
}

type repositoryConnectionResolver struct {
	opt             db.ReposListOptions
	cloned          bool
	cloneInProgress bool
	notCloned       bool
	indexed         bool
	notIndexed      bool

	// cache results because they are used by multiple fields
	once  sync.Once
	repos []*types.Repo
	err   error
}

func (r *repositoryConnectionResolver) compute(ctx context.Context) ([]*types.Repo, error) {
	r.once.Do(func() {
		opt2 := r.opt

		if envvar.SourcegraphDotComMode() {
			// Don't allow non-admins to perform huge queries on Sourcegraph.com.
			if isSiteAdmin := backend.CheckCurrentUserIsSiteAdmin(ctx) == nil; !isSiteAdmin {
				if opt2.LimitOffset == nil {
					opt2.LimitOffset = &db.LimitOffset{Limit: 1000}
				}
			}
		}

		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			// We purposefully load more repos into memory than requested in
			// order to save roundtrips to gitserver in case we need to do
			// filtering by clone status.
			// The trade-off here is memory/cpu vs. network roundtrips to
			// database/gitserver and we choose smaller latency over smaller
			// memory footprint.
			// At the end of this method we return the requested number of
			// repos.
			// As for the number: 1250 is the result of local benchmarks where
			// it yielded the best performance/resources tradeoff, before
			// diminishing returns set in
			opt2.Limit += 1250
		}

		var indexed map[string]*zoekt.Repository
		searchIndexEnabled := search.Indexed().Enabled()
		isIndexed := func(repo api.RepoName) bool {
			if !searchIndexEnabled {
				return true // do not need index
			}
			_, ok := indexed[strings.ToLower(string(repo))]
			return ok
		}
		if searchIndexEnabled && (!r.indexed || !r.notIndexed) {
			listCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			var err error
			indexed, err = search.Indexed().ListAll(listCtx)
			if err != nil {
				r.err = err
				return
			}
		}

		for {
			repos, err := backend.Repos.List(ctx, opt2)
			if err != nil {
				r.err = err
				return
			}
			reposFromDB := len(repos)

			if !r.cloned || !r.cloneInProgress || !r.notCloned {
				// Query gitserver to filter by repository clone status.
				repoNames := make([]api.RepoName, len(repos))
				for i, repo := range repos {
					repoNames[i] = repo.Name
				}
				response, err := gitserver.DefaultClient.RepoInfo(ctx, repoNames...)
				if err != nil {
					r.err = err
					return
				}
				keepRepos := repos[:0]
				for _, repo := range repos {
					if info := response.Results[repo.Name]; info == nil {
						continue
					} else if (r.cloned && info.Cloned && !info.CloneInProgress) || (r.cloneInProgress && info.CloneInProgress) || (r.notCloned && !info.Cloned && !info.CloneInProgress) {
						keepRepos = append(keepRepos, repo)
					}
				}
				repos = keepRepos
			}

			if !r.indexed || !r.notIndexed {
				keepRepos := repos[:0]
				for _, repo := range repos {
					indexed := isIndexed(repo.Name)
					if (r.indexed && indexed) || (r.notIndexed && !indexed) {
						keepRepos = append(keepRepos, repo)
					}
				}
				repos = keepRepos
			}

			r.repos = append(r.repos, repos...)

			if opt2.LimitOffset == nil {
				break
			} else {
				if len(r.repos) > r.opt.Limit {
					// Cut off the repos we additionally loaded to save
					// roundtrips to `gitserver` and only return the number
					// that was requested.
					// But, when possible, we add one more so we can detect if
					// there is a "next page" that could be loaded
					r.repos = r.repos[:r.opt.Limit+1]
					break
				}
				if reposFromDB < r.opt.Limit {
					break
				}
				opt2.Offset += opt2.Limit
			}
		}
	})
	return r.repos, r.err
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*RepositoryResolver, error) {
	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*RepositoryResolver, 0, len(repos))
	for i, repo := range repos {
		if r.opt.LimitOffset != nil && i == r.opt.Limit {
			break
		}

		resolvers = append(resolvers, &RepositoryResolver{repo: repo})
	}
	return resolvers, nil
}

func (r *repositoryConnectionResolver) TotalCount(ctx context.Context, args *struct {
	Precise bool
}) (countptr *int32, err error) {
	if isAdminErr := backend.CheckCurrentUserIsSiteAdmin(ctx); isAdminErr != nil {
		if args.Precise {
			// Only site admins can perform precise counts, because it is a slow operation.
			return nil, isAdminErr
		}
		return nil, nil
	}

	i32ptr := func(v int32) *int32 {
		return &v
	}

	if !r.cloned || !r.cloneInProgress || !r.notCloned {
		// Don't support counting if filtering by clone status.
		return nil, nil
	}
	if !r.indexed || !r.notIndexed {
		// Don't support counting if filtering by index status.
		return nil, nil
	}

	// Counting repositories is slow on Sourcegraph.com. Don't wait very long for an exact count.
	if !args.Precise && envvar.SourcegraphDotComMode() {
		if len(r.opt.Query) < 4 {
			return nil, nil
		}

		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		defer func() {
			if ctx.Err() == context.DeadlineExceeded {
				countptr = nil
				err = nil
			}
		}()
	}

	count, err := db.Repos.Count(ctx, r.opt)
	return i32ptr(int32(count)), err
}

func (r *repositoryConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(repos) > r.opt.Limit), nil
}

func (r *schemaResolver) SetRepositoryEnabled(ctx context.Context, args *struct {
	Repository graphql.ID
	Enabled    bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can enable/disable repositories, because it's a site-wide
	// and semi-destructive action.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repo, err := repositoryByID(ctx, args.Repository)
	if err != nil {
		return nil, err
	}

	// We only want to set the enabled state of a repo that isn't yet managed
	// by the new syncer. Repo-updater returns the set of external services that
	// were updated to exclude the given repo. If that set is empty, it means that
	// the given repo isn't yet managed by the new syncer, so we proceed to update
	// the enabled state regardless.
	var done bool
	if !args.Enabled {
		resp, err := repoupdater.DefaultClient.ExcludeRepo(ctx, uint32(repo.repo.ID))
		if err != nil {
			return nil, errors.Wrapf(err, "repo-updater.exclude-repos")
		}

		// Have any external services been updated to exclude the given repo?
		done = len(resp.ExternalServices) > 0
	}

	if !done {
		if err = db.Repos.SetEnabled(ctx, repo.repo.ID, args.Enabled); err != nil {
			return nil, err
		}
	}

	// Trigger update when enabling.
	if args.Enabled {
		gitserverRepo, err := backend.GitRepo(ctx, repo.repo)
		if err != nil {
			return nil, err
		}
		if _, err := repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, gitserverRepo); err != nil {
			return nil, err
		}
	}

	return &EmptyResponse{}, nil
}

func (r *schemaResolver) SetAllRepositoriesEnabled(ctx context.Context, args *struct {
	Enabled bool
}) (*EmptyResponse, error) {
	// Only usable for self-hosted instances
	if envvar.SourcegraphDotComMode() {
		return nil, errors.New("Not available on sourcegraph.com")
	}
	// 🚨 SECURITY: Only site admins can enable/disable repositories, because it's a site-wide
	// and semi-destructive action.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	var listArgs db.ReposListOptions
	if args.Enabled {
		listArgs = db.ReposListOptions{Disabled: true}
	} else {
		listArgs = db.ReposListOptions{Enabled: true}
	}
	reposList, err := db.Repos.List(ctx, listArgs)
	if err != nil {
		return nil, err
	}

	for _, repo := range reposList {
		if err := db.Repos.SetEnabled(ctx, repo.ID, args.Enabled); err != nil {
			return nil, err
		}
	}
	return &EmptyResponse{}, nil
}

func (r *schemaResolver) DeleteRepository(ctx context.Context, args *struct {
	Repository graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete repositories, because it's a site-wide
	// and semi-destructive action.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}
	if err := db.Repos.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func repoNamesToStrings(repoNames []api.RepoName) []string {
	strings := make([]string, len(repoNames))
	for i, repoName := range repoNames {
		strings[i] = string(repoName)
	}
	return strings
}

func toRepositoryResolvers(repos []*types.Repo) []*RepositoryResolver {
	if len(repos) == 0 {
		return []*RepositoryResolver{}
	}

	resolvers := make([]*RepositoryResolver, len(repos))
	for i := range repos {
		resolvers[i] = &RepositoryResolver{repo: repos[i]}
	}

	return resolvers
}

func toRepoNames(repos []*types.Repo) []api.RepoName {
	names := make([]api.RepoName, len(repos))
	for i, repo := range repos {
		names[i] = repo.Name
	}
	return names
}

func toDBRepoListColumn(ob string) db.RepoListColumn {
	switch ob {
	case "REPO_URI", "REPOSITORY_NAME":
		return db.RepoListName
	case "REPO_CREATED_AT", "REPOSITORY_CREATED_AT":
		return db.RepoListCreatedAt
	default:
		return ""
	}
}

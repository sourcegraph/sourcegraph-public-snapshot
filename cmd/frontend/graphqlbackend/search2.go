package graphqlbackend

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	zoektrpc "github.com/google/zoekt/rpc"
	"github.com/pkg/errors"
	sgbackend "github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	frontendsearch "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/endpoint"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/backend"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// This file contains the resolvers for hierarchical search. The new
// hierarchical search attempts to leave much more business logic out of the
// graphqlbackend, and instead make the resolvers more dumb.
//
// NOTE: This has not shipped yet, and will be finished up in a later
// milestone. This code path is only active if a search query is prefixed with
// "!hier!"

type searcherResolver struct {
	search.Searcher
	query.Q
	*search.Options
}

func newSearcherResolver(qStr string) (*searcherResolver, error) {
	q, err := query.Parse(qStr)
	if err != nil {
		log15.Debug("graphql search failed to parse", "query", qStr, "error", err)
		return nil, err
	}
	return &searcherResolver{
		Searcher: Search().Text,
		Q:        q,
		Options:  &search.Options{},
	}, nil
}

func (r *searcherResolver) Results(ctx context.Context) (*searchResultsResolver, error) {
	sCtx := &searchContext{}
	start := time.Now()

	// 1. Scope the request to repositories
	dbQ, err := frontendsearch.RepoQuery(r.Q)
	if err != nil {
		return nil, err
	}
	maxRepoListSize := maxReposToSearch()
	repos, err := sgbackend.Repos.List(ctx, db.ReposListOptions{
		Enabled:      true,
		PatternQuery: dbQ,
		LimitOffset:  &db.LimitOffset{Limit: maxRepoListSize + 1}, // TODO check if we hit repo list size limit
		// TODO forks and archived
	})
	if err != nil {
		return nil, err
	}
	sCtx.CacheRepo(repos...)
	opts := r.Options.ShallowCopy()
	opts.Repositories = make([]api.RepoName, len(repos))
	for i := range repos {
		opts.Repositories[i] = repos[i].Name
	}

	// 3. Adjust query so repo: atoms become reposets:
	q, err := query.ExpandRepo(r.Q, createListFunc(opts.Repositories))
	if err != nil {
		return nil, err
	}

	// 4. Do the search
	result, err := r.Searcher.Search(ctx, q, opts)
	if err != nil {
		return nil, err
	}

	// 5. To ship hierarchical search sooner we are using the old file match
	//    resolver. However, we should just be returning a resolver which is a
	//    light wrapper around a search.Result.
	results, err := toSearchResultResolvers(ctx, sCtx, result)
	if err != nil {
		return nil, err
	}
	common, err := toSearchResultsCommon(ctx, sCtx, opts, result)
	if err != nil {
		return nil, err
	}
	return &searchResultsResolver{
		results:             results,
		searchResultsCommon: *common,
		start:               start,
		alert:               toSearchAlert(result),
	}, nil
}

func (r *searcherResolver) Suggestions(ctx context.Context, args *searchSuggestionsArgs) ([]*searchSuggestionResolver, error) {
	return nil, errors.New("search suggestions not implemented")
}

func (r *searcherResolver) Stats(ctx context.Context) (stats *searchResultsStats, err error) {
	return nil, errors.New("search stats not implemented")
}

func toSearchResultResolvers(ctx context.Context, sCtx *searchContext, r *search.Result) ([]*searchResultResolver, error) {
	results := make([]*searchResultResolver, 0, len(r.Files))

	for _, file := range r.Files {
		fileLimitHit := false
		lines := make([]*lineMatch, 0, len(file.LineMatches))
		for _, l := range file.LineMatches {
			offsets := make([][2]int32, len(l.LineFragments))
			for k, m := range l.LineFragments {
				offset := utf8.RuneCount(l.Line[:m.LineOffset])
				length := utf8.RuneCount(l.Line[m.LineOffset : m.LineOffset+m.MatchLength])
				offsets[k] = [2]int32{int32(offset), int32(length)}
			}
			lines = append(lines, &lineMatch{
				JPreview:          string(l.Line),
				JLineNumber:       int32(l.LineNumber - 1),
				JOffsetAndLengths: offsets,
			})
		}

		repo, err := sCtx.GetRepo(ctx, file.Repository.Name)
		if err != nil {
			return nil, err
		}

		results = append(results, &searchResultResolver{
			fileMatch: &fileMatchResolver{
				JPath:        file.Path,
				JLineMatches: lines,
				JLimitHit:    fileLimitHit,
				uri:          fileMatchURI(file.Repository.Name, string(file.Repository.Commit), file.Path),
				repo:         repo,
				commitID:     file.Repository.Commit,
			},
		})
	}

	return results, nil
}

func toSearchResultsCommon(ctx context.Context, sCtx *searchContext, opts *search.Options, r *search.Result) (*searchResultsCommon, error) {
	var (
		repos    = map[api.RepoName]struct{}{}
		searched = map[api.RepoName]struct{}{}
		indexed  = map[api.RepoName]struct{}{}
		cloning  = map[api.RepoName]struct{}{}
		missing  = map[api.RepoName]struct{}{}
		partial  = map[api.RepoName]struct{}{}
		timedout = map[api.RepoName]struct{}{}
	)
	for _, s := range r.Stats.Status {
		repos[s.Repository.Name] = struct{}{}
		if s.Source == backend.SourceZoekt {
			indexed[s.Repository.Name] = struct{}{}
		}

		switch s.Status {
		case search.RepositoryStatusSearched:
			searched[s.Repository.Name] = struct{}{}

		case search.RepositoryStatusLimitHit:
			searched[s.Repository.Name] = struct{}{}
			partial[s.Repository.Name] = struct{}{}

		case search.RepositoryStatusTimedOut:
			timedout[s.Repository.Name] = struct{}{}

		case search.RepositoryStatusCloning:
			cloning[s.Repository.Name] = struct{}{}

		case search.RepositoryStatusMissing:
			missing[s.Repository.Name] = struct{}{}

		case search.RepositoryStatusCommitMissing:
			// Handled in toSearchAlert

		case search.RepositoryStatusError:
			return nil, errors.Errorf("error repository status: %v", s)

		default:
			return nil, errors.Errorf("unknown repository status: %v", s)
		}
	}

	unavailable := map[search.Source]bool{}
	for _, u := range r.Stats.Unavailable {
		unavailable[u] = true
	}

	var retErr error
	list := func(m map[api.RepoName]struct{}) []*types.Repo {
		repos := make([]*types.Repo, 0, len(m))
		for name := range m {
			repo, err := sCtx.GetRepo(ctx, name)
			if err != nil {
				retErr = err
				continue
			}
			repos = append(repos, repo)
		}
		return repos
	}

	common := &searchResultsCommon{
		maxResultsCount:  int32(opts.TotalMaxMatchCount),
		resultCount:      int32(r.Stats.MatchCount),
		indexUnavailable: unavailable[backend.SourceZoekt],

		searched: list(searched),
		indexed:  list(indexed),
		cloning:  list(cloning),
		missing:  list(missing),
		timedout: list(timedout),
		partial:  partial,
	}
	if retErr != nil {
		return nil, retErr
	}
	return common, nil
}

func toSearchAlert(r *search.Result) *searchAlert {
	missing := map[api.RepoName][]string{}
	for _, s := range r.Stats.Status {
		if s.Status != search.RepositoryStatusCommitMissing {
			continue
		}
		missing[s.Repository.Name] = append(missing[s.Repository.Name], s.Repository.RefPattern)
	}
	if len(missing) == 0 {
		return nil
	}

	var description string
	if len(missing) == 1 {
		var name api.RepoName
		var patterns []string
		for name, patterns = range missing {
			break
		}
		if len(patterns) == 1 {
			description = fmt.Sprintf("The repository %s matched by your repo: filter could not be searched because it does not contain the revision %q.", name, patterns[0])
		} else {
			description = fmt.Sprintf("The repository %s matched by your repo: filter could not be searched because it has multiple specified revisions: @%s.", name, strings.Join(patterns, ","))
		}
	} else {
		repoRevs := make([]string, 0, len(missing))
		for name, patterns := range missing {
			repoRevs = append(repoRevs, string(name)+"@"+strings.Join(patterns, ","))
		}
		description = fmt.Sprintf("%d repositories matched by your repo: filter could not be searched because the following revisions do not exist, or differ but were specified for the same repository: %s.", len(missing), strings.Join(repoRevs, ", "))
	}
	return &searchAlert{
		title:       "Some repositories could not be searched",
		description: description,
	}
}

// createListFunc returns a list function for query.ExpandRepo based on
// matching repo: atoms as regular expressions. See documentation for
// query.ExpandRepo.
func createListFunc(repos []api.RepoName) func([]string, []string) (map[string]struct{}, error) {
	compile := func(ps []string) ([]*regexp.Regexp, error) {
		res := make([]*regexp.Regexp, 0, len(ps))
		for _, p := range ps {
			re, err := regexp.Compile(p)
			if err != nil {
				return nil, err
			}
			res = append(res, re)
		}
		return res, nil
	}
	return func(inc, exc []string) (map[string]struct{}, error) {
		incRes, err := compile(inc)
		if err != nil {
			return nil, err
		}
		excRes, err := compile(exc)
		if err != nil {
			return nil, err
		}

		set := map[string]struct{}{}
		for _, r := range repos {
			matched := true
			for _, re := range incRes {
				if !matched {
					break
				}
				matched = matched && re.MatchString(string(r))
			}
			for _, re := range excRes {
				if !matched {
					break
				}
				matched = matched && !re.MatchString(string(r))
			}
			if matched {
				set[string(r)] = struct{}{}
			}
		}
		return set, nil
	}
}

// searchContext is used to reduce duplicate DB and gitserver calls.
type searchContext struct {
	mu    sync.Mutex
	repos map[api.RepoName]*types.Repo
}

func (s *searchContext) GetRepo(ctx context.Context, name api.RepoName) (*types.Repo, error) {
	s.mu.Lock()
	if s.repos == nil {
		s.repos = map[api.RepoName]*types.Repo{}
	}
	r, ok := s.repos[name]
	s.mu.Unlock()
	if ok {
		return r, nil
	}
	r, err := db.Repos.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	s.CacheRepo(r)
	return r, nil
}

func (s *searchContext) CacheRepo(repos ...*types.Repo) {
	s.mu.Lock()
	if s.repos == nil {
		s.repos = map[api.RepoName]*types.Repo{}
	}
	for _, r := range repos {
		s.repos[r.Name] = r
	}
	s.mu.Unlock()
}

// SearchProviders contains instances of our search providers.
type SearchProviders struct {
	// Text is our root text searcher.
	Text *backend.Text

	// SearcherURLs is an endpoint map to our searcher service replicas.
	//
	// Note: This field will be removed once we have removed our old search
	// code paths.
	SearcherURLs *endpoint.Map

	// Index is a search.Searcher for Zoekt.
	Index *backend.Zoekt
}

var (
	zoektAddr   = env.Get("ZOEKT_HOST", "indexed-search:80", "host:port of the zoekt instance")
	searcherURL = env.Get("SEARCHER_URL", "k8s+http://searcher:3181", "searcher server URL")

	searchOnce sync.Once
	searchP    *SearchProviders
)

// Search returns instances of our search providers.
func Search() *SearchProviders {
	searchOnce.Do(func() {
		// Zoekt
		index := &backend.Zoekt{}
		if zoektAddr != "" {
			index.Client = zoektrpc.Client(zoektAddr)
		}
		go func() {
			conf.Watch(func() {
				index.SetEnabled(conf.SearchIndexEnabled())
			})
		}()

		// Searcher
		var searcherURLs *endpoint.Map
		if len(strings.Fields(searcherURL)) == 0 {
			searcherURLs = endpoint.Empty(errors.New("a searcher service has not been configured"))
		} else {
			searcherURLs = endpoint.New(searcherURL)
		}

		text := &backend.Text{
			Index: index,
			Fallback: &backend.TextJIT{
				Endpoints: searcherURLs,
				Resolve: func(ctx context.Context, name api.RepoName, spec string) (api.CommitID, error) {
					// Do not trigger a repo-updater lookup (e.g.,
					// backend.{GitRepo,Repos.ResolveRev}) because that would
					// slow this operation down by a lot (if we're looping
					// over many repos). This means that it'll fail if a repo
					// is not on gitserver.
					return git.ResolveRevision(ctx, gitserver.Repo{Name: name}, nil, spec, &git.ResolveRevisionOptions{NoEnsureRevision: true})
				},
			},
		}

		searchP = &SearchProviders{
			Text:         text,
			SearcherURLs: searcherURLs,
			Index:        index,
		}
	})
	return searchP
}

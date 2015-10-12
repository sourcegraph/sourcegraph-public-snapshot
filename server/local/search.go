package local

import (
	"math/rand"
	"strings"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	searchpkg "src.sourcegraph.com/sourcegraph/search"
	"src.sourcegraph.com/sourcegraph/svc"
)

var Search sourcegraph.SearchServer = &search{}

type search struct{}

var _ sourcegraph.SearchServer = (*search)(nil)

func (s *search) SearchTokens(ctx context.Context, opt *sourcegraph.TokenSearchOptions) (*sourcegraph.DefList, error) {
	defListOpts := &sourcegraph.DefListOptions{
		Query:       opt.Query,
		RepoRevs:    []string{opt.RepoRev.URI},
		ListOptions: opt.ListOptions,
		Nonlocal:    true,
		Doc:         true,
	}

	defList, err := svc.Defs(ctx).List(ctx, defListOpts)
	if err != nil {
		return nil, err
	}

	return defList, nil
}

func (s *search) SearchText(ctx context.Context, opt *sourcegraph.TextSearchOptions) (*sourcegraph.VCSSearchResultList, error) {
	vcsSearchOpts := &sourcegraph.RepoTreeSearchOptions{
		Formatted: true,
		SearchOptions: vcs.SearchOptions{
			Query:        opt.Query,
			QueryType:    "fixed",
			ContextLines: 1,
			N:            opt.ListOptions.PerPage,
			Offset:       (opt.ListOptions.Page - 1) * opt.ListOptions.PerPage,
		},
	}

	results, err := svc.RepoTree(ctx).Search(ctx, &sourcegraph.RepoTreeSearchOp{Rev: opt.RepoRev, Opt: vcsSearchOpts})
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *search) Search(ctx context.Context, opt *sourcegraph.SearchOptions) (*sourcegraph.SearchResults, error) {
	if opt == nil {
		panic("Search: nil opt")
	}

	var err error

	res := &sourcegraph.SearchResults{}
	res.RawQuery = sourcegraph.RawQuery{Text: opt.Query}

	res.Tokens, res.ResolvedTokens, _, res.ResolveErrors, err = s.resolveQuery(ctx, res.RawQuery)
	if err != nil {
		return nil, err
	}
	if len(res.ResolveErrors) > 0 {
		res.Canceled = true
		return res, nil
	}

	res.Canceled, res.Tips, err = searchpkg.Tips(ctx, sourcegraph.PBTokens(res.ResolvedTokens))
	if err != nil {
		return nil, err
	}
	if res.Canceled {
		return res, nil
	}

	res.Plan, err = searchpkg.NewPlan(ctx, sourcegraph.PBTokens(res.ResolvedTokens))
	if err != nil {
		return nil, err
	}

	if res.Plan.Repos != nil && opt.Repos {
		res.Plan.Repos.ListOptions = opt.ListOptions
		if res.Plan.Repos.ListOptions.PerPage == 0 {
			res.Plan.Repos.ListOptions.PerPage = 5
		}
		repos, err := svc.Repos(ctx).List(ctx, res.Plan.Repos)
		if err != nil {
			return nil, err
		}
		res.Repos = repos.Repos
	}

	var defsUnavailable bool

	if res.Plan.Defs != nil && opt.Defs {
		res.Plan.Defs.ListOptions = opt.ListOptions
		res.Plan.Defs.Doc = true
		defList, err := svc.Defs(ctx).List(ctx, res.Plan.Defs)
		if err != nil {
			return nil, err
		}
		res.Defs = defList.Defs
		if len(res.Defs) == 0 {
			defsUnavailable = true
		}
	}

	if res.Plan.Users != nil && opt.People {
		res.Plan.Users.ListOptions = opt.ListOptions
		users, err := svc.Users(ctx).List(ctx, res.Plan.Users)
		if err != nil {
			return nil, err
		}
		for _, user := range users.Users {
			res.People = append(res.People, user.Person())
		}
	}

	if res.Plan.Tree != nil && opt.Tree {
		res.Plan.Tree.N = int32(opt.ListOptions.Limit())
		res.Plan.Tree.Offset = int32(opt.ListOptions.Offset())
		res.Plan.Tree.Formatted = true
		res.Plan.Tree.ContextLines = 1
		// TODO(sqs): parallelize
		for _, repoRev := range res.Plan.TreeRepoRevs {
			repoURI, commitID := searchpkg.ParseRepoAndCommitID(repoRev)
			repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: repoURI}, Rev: commitID}
			sresults, err := svc.RepoTree(ctx).Search(ctx, &sourcegraph.RepoTreeSearchOp{Rev: repoRevSpec, Opt: res.Plan.Tree})
			if err != nil {
				return nil, err
			}
			for _, sres := range sresults.SearchResults {
				res.Tree = append(res.Tree, &sourcegraph.RepoTreeSearchResult{
					SearchResult: *sres,
					RepoRev:      repoRevSpec,
				})
			}
		}
	}

	if defsUnavailable {
		noCache(ctx)
	} else {
		mediumCache(ctx)
	}
	return res, nil
}

func (s *search) Complete(ctx context.Context, q *sourcegraph.RawQuery) (*sourcegraph.Completions, error) {
	var err error
	var rawTokens []sourcegraph.PBToken
	var st searchpkg.State
	comps := &sourcegraph.Completions{}
	rawTokens, comps.ResolvedTokens, st, comps.ResolveErrors, err = s.resolveQuery(ctx, *q)
	if err != nil {
		return nil, err
	}

	var activeTok sourcegraph.Token
	var scope []sourcegraph.Token
	for i, tok := range rawTokens {
		if st.IsActive(i) {
			activeTok = tok.Token()

			// Set all preceding resolved tokens as the scope.
			scope = sourcegraph.PBTokens(comps.ResolvedTokens[:i])
			break
		}
	}

	if activeTok != nil {
		tokComps, err := searchpkg.CompleteToken(ctx, activeTok, scope, searchpkg.TokenCompletionConfig{MaxPerType: 7})
		if err != nil {
			return nil, err
		}
		comps.TokenCompletions = sourcegraph.PBTokensWrap(tokComps)
	}

	mediumCache(ctx)
	return comps, nil
}

func (s *search) Suggest(ctx context.Context, q *sourcegraph.RawQuery) (*sourcegraph.SuggestionList, error) {
	_, resolvedTokens, _, _, err := s.resolveQuery(ctx, *q)
	if err != nil {
		return nil, err
	}

	suggs, err := searchpkg.Suggest(ctx, sourcegraph.PBTokens(resolvedTokens), searchpkg.SuggestionConfig{MaxPerType: 7})
	if err != nil {
		return nil, err
	}

	// Only get 5 total.
	const max = 5
	if len(suggs) > max {
		perms := rand.Perm(len(suggs))
		for i, perm := range perms {
			suggs[i] = suggs[perm]
		}
		suggs = suggs[:max]
	}

	for _, sugg := range suggs {
		qstr, err := searchpkg.Shorten(ctx, sourcegraph.PBTokens(sugg.Query))
		if err != nil {
			return nil, err
		}
		sugg.QueryString = strings.Join(qstr, " ")
	}

	mediumCache(ctx)
	return &sourcegraph.SuggestionList{Suggestions: suggs}, nil
}

// resolveQuery parses and resolves tokens in rawQuery.
func (s *search) resolveQuery(ctx context.Context, rawQuery sourcegraph.RawQuery) (raw, resolved []sourcegraph.PBToken, st searchpkg.State, resolveErrs []sourcegraph.TokenError, otherErr error) {
	var err error

	var rawTmp []sourcegraph.Token
	rawTmp, st, err = searchpkg.Parse(rawQuery)
	if err != nil {
		return nil, nil, searchpkg.State{}, nil, err
	}

	// TODO(sqs!nodb-ctx): ensure current user login is in context for
	// search (some search funcs use it to improve results).

	var resolvedTmp []sourcegraph.Token
	resolvedTmp, resolveErrs, err = searchpkg.Resolve(ctx, rawTmp)
	if err != nil {
		return nil, nil, searchpkg.State{}, nil, err
	}

	return sourcegraph.PBTokensWrap(rawTmp), sourcegraph.PBTokensWrap(resolvedTmp), st, resolveErrs, nil
}

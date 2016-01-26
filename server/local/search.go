package local

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/eventsutil"
)

var Search sourcegraph.SearchServer = &search{}

type search struct{}

var _ sourcegraph.SearchServer = (*search)(nil)

func (s *search) SearchTokens(ctx context.Context, opt *sourcegraph.TokenSearchOptions) (*sourcegraph.DefList, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Search.SearchTokens", opt.RepoRev.URI); err != nil {
		return nil, err
	}
	defListOpts := &sourcegraph.DefListOptions{
		Query:       opt.Query,
		RepoRevs:    []string{opt.RepoRev.URI + "@" + opt.RepoRev.CommitID},
		ListOptions: opt.ListOptions,
		Nonlocal:    true,
		Doc:         true,
	}

	defList, err := svc.Defs(ctx).List(ctx, defListOpts)
	if err != nil {
		return nil, err
	}

	eventsutil.LogSearchQuery(ctx, "TokenSearch", defList.Total)

	return defList, nil
}

func (s *search) SearchText(ctx context.Context, opt *sourcegraph.TextSearchOptions) (*sourcegraph.VCSSearchResultList, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Search.SearchText", opt.RepoRev.URI); err != nil {
		return nil, err
	}
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

	eventsutil.LogSearchQuery(ctx, "TextSearch", results.Total)

	return results, nil
}

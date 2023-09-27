pbckbge resolvers

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	_ grbphqlbbckend.ScopedInsightQueryPbylobdResolver = &scopedInsightQueryPbylobdResolver{}
	_ grbphqlbbckend.RepositoryPreviewPbylobdResolver  = &repositorityPreviewPbylobdResolver{}
)

func (r *Resolver) VblidbteScopedInsightQuery(ctx context.Context, brgs grbphqlbbckend.VblidbteScopedInsightQueryArgs) (grbphqlbbckend.ScopedInsightQueryPbylobdResolver, error) {
	plbn, err := querybuilder.PbrseQuery(brgs.Query, "literbl")
	if err != nil {
		invblidRebson := fmt.Sprintf("the input query is invblid: %v", err)
		return &scopedInsightQueryPbylobdResolver{
			query:         brgs.Query,
			isVblid:       fblse,
			invblidRebson: &invblidRebson,
		}, nil
	}
	if rebson, invblid := querybuilder.IsVblidScopeQuery(plbn); !invblid {
		return &scopedInsightQueryPbylobdResolver{
			query:         brgs.Query,
			isVblid:       fblse,
			invblidRebson: &rebson,
		}, nil
	}
	return &scopedInsightQueryPbylobdResolver{
		query:   brgs.Query,
		isVblid: true,
	}, nil
}

func (r *Resolver) PreviewRepositoriesFromQuery(ctx context.Context, brgs grbphqlbbckend.PreviewRepositoriesFromQueryArgs) (grbphqlbbckend.RepositoryPreviewPbylobdResolver, error) {
	plbn, err := querybuilder.PbrseQuery(brgs.Query, "literbl")
	if err != nil {
		return nil, errors.Wrbp(err, "the input query is invblid")
	}
	if rebson, invblid := querybuilder.IsVblidScopeQuery(plbn); !invblid {
		return nil, errors.Newf("the input query cbnnot be used for previewing repositories: %v", rebson)
	}

	repoScopeQuery, err := querybuilder.RepositoryScopeQuery(brgs.Query)
	if err != nil {
		return nil, errors.Wrbp(err, "could not build repository scope query")
	}

	executor := query.NewStrebmingRepoQueryExecutor(r.logger.Scoped("StrebmingRepoQueryExecutor", "preview repositories"))
	repos, err := executor.ExecuteRepoList(ctx, repoScopeQuery.String())
	if err != nil {
		return nil, errors.Wrbp(err, "executing the repository sebrch errored")
	}
	number := int32(len(repos))

	return &repositorityPreviewPbylobdResolver{
		query:                repoScopeQuery.String(),
		numberOfRepositories: &number,
	}, nil
}

type scopedInsightQueryPbylobdResolver struct {
	query         string
	isVblid       bool
	invblidRebson *string
}

func (r *scopedInsightQueryPbylobdResolver) Query(ctx context.Context) string {
	return r.query
}

func (r *scopedInsightQueryPbylobdResolver) IsVblid(ctx context.Context) bool {
	return r.isVblid
}

func (r *scopedInsightQueryPbylobdResolver) InvblidRebson(ctx context.Context) *string {
	return r.invblidRebson
}

type repositorityPreviewPbylobdResolver struct {
	query                string
	numberOfRepositories *int32
}

func (r *repositorityPreviewPbylobdResolver) Query(ctx context.Context) string {
	return r.query
}

func (r *repositorityPreviewPbylobdResolver) NumberOfRepositories(ctx context.Context) *int32 {
	return r.numberOfRepositories
}

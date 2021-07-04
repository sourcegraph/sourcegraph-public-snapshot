package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	codeintelgraphql "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

type RootResolver struct {
	codeIntelGQLResolver func() graphqlbackend.CodeIntelResolver
	getCodeIntelResolver func() codeintelresolvers.Resolver
}

func NewResolver(db dbutil.DB, codeIntelGQLResolver func() graphqlbackend.CodeIntelResolver, getCodeIntelResolver func() codeintelresolvers.Resolver, clock func() time.Time) graphqlbackend.GuideRootResolver {
	return &RootResolver{codeIntelGQLResolver: codeIntelGQLResolver, getCodeIntelResolver: getCodeIntelResolver}
}

func (r RootResolver) GuideInfo(ctx context.Context, args *graphqlbackend.GuideInfoParams) (graphqlbackend.GuideInfoResolver, error) {
	repo, err := graphqlbackend.UnmarshalRepositoryID(*args.Repository.ID)
	if err != nil {
		return nil, err
	}

	return &InfoResolver{
		repo:     repo,
		revision: *args.Repository.Revision,
		commitID: *args.Repository.CommitID,

		codeIntelGQLResolver: r.codeIntelGQLResolver(),
		codeIntelResolver:    r.getCodeIntelResolver(),
	}, nil
}

type InfoResolver struct {
	repo               api.RepoID
	revision, commitID string

	codeIntelGQLResolver graphqlbackend.CodeIntelResolver
	codeIntelResolver    codeintelresolvers.Resolver
}

func (InfoResolver) Hello() string { return "world" }

func (InfoResolver) URL() string { return "/foo" }

func (InfoResolver) Monikers(ctx context.Context) ([]graphqlbackend.MonikerResolver, error) {
	return []graphqlbackend.MonikerResolver{
		codeintelgraphql.NewMonikerResolver(semantic.MonikerData{
			Scheme:     "gomod",
			Identifier: "github.com/hashicorp/go-multierror:Append",
		}),
	}, nil
}

func (r InfoResolver) Hover(ctx context.Context) (graphqlbackend.HoverResolver, error) {
	codeIntelQueryResolver, err := r.codeIntelResolver.QueryResolver(ctx, &graphqlbackend.GitBlobLSIFDataArgs{
		Repo:   &types.Repo{ID: r.repo},
		Commit: api.CommitID(r.commitID),
		Path:   "append.go",
	})
	if err != nil {
		return nil, err
	}

	text, rx, exists, err := codeIntelQueryResolver.Hover(ctx, 8, 6)
	if err != nil || !exists {
		return nil, err
	}
	return codeintelgraphql.NewHoverResolver(text, codeintelgraphql.ConvertRange(rx)), nil
}

func (r InfoResolver) References(ctx context.Context) (graphqlbackend.LocationConnectionResolver, error) {
	const path = "append.go"
	r2, err := r.codeIntelGQLResolver.GitBlobLSIFData(ctx, &graphqlbackend.GitBlobLSIFDataArgs{
		Repo:   &types.Repo{ID: r.repo},
		Commit: api.CommitID(r.commitID),
		Path:   path,
	})
	if err != nil {
		return nil, err
	}

	return r2.References(ctx, &graphqlbackend.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: graphqlbackend.LSIFQueryPositionArgs{
			Line: 8, Character: 6,
		},
	})
}

func (r InfoResolver) EditCommits(ctx context.Context) (*graphqlbackend.GitCommitConnectionResolver, error) {
	const path = "append.go"
	codeIntelQueryResolver, err := r.codeIntelResolver.QueryResolver(ctx, &graphqlbackend.GitBlobLSIFDataArgs{
		Repo:   &types.Repo{ID: r.repo},
		Commit: api.CommitID(r.commitID),
		Path:   path,
	})
	if err != nil {
		return nil, err
	}

	// TODO(sqs): look up moniker
	symbol, _, err := codeIntelQueryResolver.Symbol(ctx, "gomod", "github.com/hashicorp/go-multierror:Append")
	if err != nil {
		return nil, err
	}

	var locations []lsifstore.Range
	locations = append(locations, symbol.AdjustedLocation.AdjustedRange)

	// git log -L
	var lineRanges []string
	for _, loc := range locations {
		// TODO(sqs): use full range
		lineRanges = append(lineRanges, fmt.Sprintf("%d,%d:%s", loc.Start.Line+1, loc.End.Line+1, path))
	}

	return graphqlbackend.NewGitCommitConnectionResolver(string(r.commitID), lineRanges, r.repo), nil
}

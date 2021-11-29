package resolvers

import (
	"context"
	"regexp"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *catalogComponentResolver) API(ctx context.Context, args *gql.CatalogComponentAPIArgs) (gql.CatalogComponentAPIResolver, error) {
	repoResolver, err := r.sourceRepoResolver(ctx)
	if err != nil {
		return nil, err
	}
	commitResolver := gql.NewGitCommitResolver(r.db, repoResolver, api.CommitID(r.sourceCommit), nil)

	// Only find symbols in the component's paths.
	includePatterns := make([]string, len(r.sourcePaths))
	for _, p := range r.sourcePaths {
		includePatterns = append(includePatterns, "^"+regexp.QuoteMeta(p)+"($|/)")
	}

	symbols, err := commitResolver.Symbols(ctx, &gql.SymbolsArgs{
		Query:           args.Query,
		IncludePatterns: &includePatterns,
	})
	if err != nil {
		return nil, err
	}

	return &catalogComponentAPIResolver{
		symbols:   symbols,
		component: r,
		db:        r.db,
	}, nil
}

type catalogComponentAPIResolver struct {
	symbols *gql.SymbolConnectionResolver

	component *catalogComponentResolver
	db        database.DB
}

func (r *catalogComponentAPIResolver) Symbols(ctx context.Context, args *gql.CatalogComponentAPISymbolsArgs) (*gql.SymbolConnectionResolver, error) {
	// TODO(sqs): args.First is ignored
	return r.symbols, nil
}

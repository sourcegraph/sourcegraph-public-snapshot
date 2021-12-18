package resolvers

import (
	"context"
	"regexp"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *componentResolver) API(ctx context.Context, args *gql.ComponentAPIArgs) (gql.ComponentAPIResolver, error) {
	repoResolver, err := r.sourceRepoResolver(ctx)
	if err != nil {
		return nil, err
	}
	commitResolver := gql.NewGitCommitResolver(r.db, repoResolver, api.CommitID(r.component.SourceCommit), nil)

	// Only find symbols in the component's paths.
	includePatterns := make([]string, len(r.component.SourcePaths))
	for _, p := range r.component.SourcePaths {
		includePatterns = append(includePatterns, "^"+regexp.QuoteMeta(p)+"($|/)")
	}

	symbols, err := commitResolver.Symbols(ctx, &gql.SymbolsArgs{
		Query:           args.Query,
		IncludePatterns: &includePatterns,
	})
	if err != nil {
		return nil, err
	}

	return &componentAPIResolver{
		symbols:   symbols,
		component: r,
		db:        r.db,
	}, nil
}

type componentAPIResolver struct {
	symbols *gql.SymbolConnectionResolver

	component *componentResolver
	db        database.DB
}

func (r *componentAPIResolver) Symbols(ctx context.Context, args *gql.ComponentAPISymbolsArgs) (*gql.SymbolConnectionResolver, error) {
	// TODO(sqs): args.First is ignored
	return r.symbols, nil
}

func (r *componentAPIResolver) Schema(ctx context.Context) (gql.FileResolver, error) {
	if r.component.component.APIDefPath == "" {
		return nil, nil
	}

	commitResolver, err := r.component.sourceCommitResolver(ctx)
	if err != nil {
		return nil, err
	}
	return commitResolver.File(ctx, &struct{ Path string }{Path: r.component.component.APIDefPath})
}

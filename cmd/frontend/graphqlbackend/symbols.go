package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/symbols"
)

type symbolsArgs struct {
	graphqlutil.ConnectionArgs
	Query           *string
	IncludePatterns *[]string
}

func (r *gitTreeEntryResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := computeSymbols(ctx, computeSymbolsArgs{
		commit:          r.commit,
		repo:            r.commit.repo,
		query:           args.Query,
		first:           args.First,
		includePatterns: args.IncludePatterns,
	})
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{symbols: symbols, first: args.First}, nil
}

func (r *gitCommitResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := computeSymbols(ctx, computeSymbolsArgs{
		commit:          r,
		repo:            r.repo,
		query:           args.Query,
		first:           args.First,
		includePatterns: args.IncludePatterns,
	})
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{symbols: symbols, first: args.First}, nil
}

type symbolConnectionResolver struct {
	first   *int32
	symbols []*symbolResolver
}

type computeSymbolsArgs struct {
	commit          *gitCommitResolver
	repo            *repositoryResolver
	query           *string
	first           *int32
	includePatterns *[]string
}

func computeSymbols(ctx context.Context, args computeSymbolsArgs) (res []*symbolResolver, err error) {
	oid, err := args.commit.OID()
	if err != nil {
		return nil, err
	}

	symbols, err := symbols.ComputeSymbols(ctx, symbols.ComputeSymbolsArgs{
		CommitID:        api.CommitID(oid),
		RepoName:        args.repo.repo.Name,
		Query:           args.query,
		First:           args.first,
		IncludePatterns: args.includePatterns,
	})
	if err != nil {
		return nil, err
	}

	resolvers := make([]*symbolResolver, len(symbols))
	for i, s := range symbols {
		resolver := toSymbolResolver(s, args.commit)
		fmt.Println(resolver, s)
		resolvers[i] = resolver
	}

	return resolvers, nil
}

type symbolResolver struct {
	symbol   *symbols.Symbol
	location *locationResolver
}

func toSymbolResolver(symbol *symbols.Symbol, commit *gitCommitResolver) *symbolResolver {
	return &symbolResolver{
		symbol: symbol,
		location: &locationResolver{
			resource: &gitTreeEntryResolver{
				commit: commit,
				path:   symbol.URI.Fragment,
				stat:   createFileInfo(symbol.URI.Fragment, false), // assume the path refers to a file (not dir)
			},
			lspRange: symbol.Range,
		},
	}
}

func (r *symbolConnectionResolver) Nodes(ctx context.Context) ([]*symbolResolver, error) {
	nodes := r.symbols
	if len(r.symbols) > symbols.LimitOrDefault(r.first) {
		nodes = nodes[:symbols.LimitOrDefault(r.first)]
	}
	return nodes, nil
}

func (r *symbolConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.symbols) > symbols.LimitOrDefault(r.first)), nil
}

func (r *symbolResolver) Name() string { return r.symbol.Symbol.Name }

func (r *symbolResolver) ContainerName() *string {
	if r.symbol.Symbol.Parent == "" {
		return nil
	}
	return &r.symbol.Symbol.Parent
}

func (r *symbolResolver) Kind() string /* enum SymbolKind */ {
	return r.symbol.Kind()
}

func (r *symbolResolver) Language() string { return r.symbol.Symbol.Language }

func (r *symbolResolver) Location() *locationResolver { return r.location }

func (r *symbolResolver) URL(ctx context.Context) (string, error) { return r.location.URL(ctx) }

func (r *symbolResolver) CanonicalURL() (string, error) { return r.location.CanonicalURL() }

func (r *symbolResolver) FileLocal() bool { return r.symbol.Symbol.FileLimited }

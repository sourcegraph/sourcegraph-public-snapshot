package graphqlbackend

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gituri"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/symbols/protocol"
)

type symbolsArgs struct {
	graphqlutil.ConnectionArgs
	Query           *string
	IncludePatterns *[]string
}

func (r *GitTreeEntryResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := computeSymbols(ctx, r.commit, args.Query, args.First, args.IncludePatterns)
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{symbols: symbols, first: args.First}, nil
}

func (r *GitCommitResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := computeSymbols(ctx, r, args.Query, args.First, args.IncludePatterns)
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{symbols: symbols, first: args.First}, nil
}

type symbolConnectionResolver struct {
	first   *int32
	symbols []*symbolResolver
}

func limitOrDefault(first *int32) int {
	if first == nil {
		return 100
	}
	return int(*first)
}

func computeSymbols(ctx context.Context, commit *GitCommitResolver, query *string, first *int32, includePatterns *[]string) (res []*symbolResolver, err error) {
	ctx, done := context.WithTimeout(ctx, 5*time.Second)
	defer done()
	defer func() {
		if ctx.Err() != nil && len(res) == 0 {
			err = errors.New("processing symbols is taking longer than expected. Try again in a while")
		}
	}()
	var includePatternsSlice []string
	if includePatterns != nil {
		includePatternsSlice = *includePatterns
	}
	searchArgs := search.SymbolsParameters{
		CommitID:        api.CommitID(commit.oid),
		First:           limitOrDefault(first) + 1, // add 1 so we can determine PageInfo.hasNextPage
		Repo:            commit.repo.repo.Name,
		IncludePatterns: includePatternsSlice,
	}
	if query != nil {
		searchArgs.Query = *query
	}
	baseURI, err := gituri.Parse("git://" + string(commit.repo.repo.Name) + "?" + string(commit.oid))
	if err != nil {
		return nil, err
	}
	symbols, err := backend.Symbols.ListTags(ctx, searchArgs)
	if baseURI == nil {
		return
	}
	resolvers := make([]*symbolResolver, 0, len(symbols))
	for _, symbol := range symbols {
		resolver := toSymbolResolver(symbol, baseURI, strings.ToLower(symbol.Language), commit)
		if resolver == nil {
			continue
		}
		resolvers = append(resolvers, resolver)
	}
	return resolvers, err
}

func toSymbolResolver(symbol protocol.Symbol, baseURI *gituri.URI, lang string, commitResolver *GitCommitResolver) *symbolResolver {
	resolver := &symbolResolver{
		symbol:   symbol,
		language: lang,
		uri:      baseURI.WithFilePath(symbol.Path),
	}
	symbolRange := symbolRange(symbol)
	resolver.location = &locationResolver{
		resource: &GitTreeEntryResolver{
			commit: commitResolver,
			stat:   CreateFileInfo(resolver.uri.Fragment, false), // assume the path refers to a file (not dir)
		},
		lspRange: &symbolRange,
	}
	return resolver
}

func (r *symbolConnectionResolver) Nodes(ctx context.Context) ([]*symbolResolver, error) {
	symbols := r.symbols
	if len(r.symbols) > limitOrDefault(r.first) {
		symbols = symbols[:limitOrDefault(r.first)]
	}
	return symbols, nil
}

func (r *symbolConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.symbols) > limitOrDefault(r.first)), nil
}

type symbolResolver struct {
	symbol   protocol.Symbol
	language string
	location *locationResolver
	uri      *gituri.URI
}

func (r *symbolResolver) Name() string { return r.symbol.Name }

func (r *symbolResolver) ContainerName() *string {
	if r.symbol.Parent == "" {
		return nil
	}
	return &r.symbol.Parent
}

func (r *symbolResolver) Kind() string /* enum SymbolKind */ {
	kind := ctagsKindToLSPSymbolKind(r.symbol.Kind)
	if kind == 0 {
		return "UNKNOWN"
	}
	return strings.ToUpper(kind.String())
}

func (r *symbolResolver) Language() string { return r.language }

func (r *symbolResolver) Location() *locationResolver { return r.location }

func (r *symbolResolver) URL(ctx context.Context) (string, error) { return r.location.URL(ctx) }

func (r *symbolResolver) CanonicalURL() (string, error) { return r.location.CanonicalURL() }

func (r *symbolResolver) FileLocal() bool { return r.symbol.FileLimited }

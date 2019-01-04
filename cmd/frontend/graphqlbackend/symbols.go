package graphqlbackend

import (
	"context"
	"errors"
	"strings"
	"time"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type symbolsArgs struct {
	graphqlutil.ConnectionArgs
	Query *string
}

func (r *gitTreeEntryResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := computeSymbols(ctx, r.commit, args.Query, args.First)
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{symbols: symbols, first: args.First}, nil
}

func (r *gitCommitResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := computeSymbols(ctx, r, args.Query, args.First)
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

func computeSymbols(ctx context.Context, commit *gitCommitResolver, query *string, first *int32) (res []*symbolResolver, err error) {
	ctx, done := context.WithTimeout(ctx, 5*time.Second)
	defer done()
	defer func() {
		if ctx.Err() != nil && len(res) == 0 {
			err = errors.New("processing symbols is taking longer than expected. Try again in a while")
		}
	}()
	searchArgs := protocol.SearchArgs{
		CommitID: api.CommitID(commit.oid),
		First:    limitOrDefault(first) + 1, // add 1 so we can determine PageInfo.hasNextPage
		Repo:     commit.repo.repo.Name,
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
		resolver := toSymbolResolver(symbolToLSPSymbolInformation(symbol, baseURI), strings.ToLower(symbol.Language), commit)
		if resolver == nil {
			continue
		}
		resolvers = append(resolvers, resolver)
	}
	return resolvers, err
}

func toSymbolResolver(symbol lsp.SymbolInformation, lang string, commitResolver *gitCommitResolver) *symbolResolver {
	resolver := &symbolResolver{
		symbol:   symbol,
		language: lang,
	}
	uri, err := gituri.Parse(string(symbol.Location.URI))
	if err != nil {
		log15.Warn("Omitting symbol with invalid URI from results.", "uri", symbol.Location.URI, "error", err)
		return nil
	}
	symbolRange := symbol.Location.Range // copy
	resolver.location = &locationResolver{
		resource: &gitTreeEntryResolver{
			commit: commitResolver,
			path:   uri.Fragment,
			stat:   createFileInfo(uri.Fragment, false), // assume the path refers to a file (not dir)
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
	symbol   lsp.SymbolInformation
	language string
	location *locationResolver
}

func (r *symbolResolver) Name() string { return r.symbol.Name }

func (r *symbolResolver) ContainerName() *string {
	if r.symbol.ContainerName == "" {
		return nil
	}
	return &r.symbol.ContainerName
}

func (r *symbolResolver) Kind() string /* enum SymbolKind */ {
	return strings.ToUpper(r.symbol.Kind.String())
}

func (r *symbolResolver) Language() string { return r.language }

func (r *symbolResolver) Location() *locationResolver { return r.location }

func (r *symbolResolver) URL(ctx context.Context) string { return r.location.URL(ctx) }

func (r *symbolResolver) CanonicalURL() string { return r.location.CanonicalURL() }

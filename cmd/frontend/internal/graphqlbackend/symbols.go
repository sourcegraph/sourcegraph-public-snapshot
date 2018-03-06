package graphqlbackend

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

type symbolsArgs struct {
	connectionArgs
	Query *string
}

func (r *repositoryResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	var rev string
	if r.repo.IndexedRevision != nil {
		rev = string(*r.repo.IndexedRevision)
	}
	commit, err := r.Commit(ctx, &struct{ Rev string }{Rev: rev})
	if err != nil {
		return nil, err
	}
	return &symbolConnectionResolver{
		first:  args.First,
		query:  args.Query,
		commit: commit,
	}, nil
}

func (r *fileResolver) Symbols(args *symbolsArgs) *symbolConnectionResolver {
	return &symbolConnectionResolver{
		first:  args.First,
		query:  args.Query,
		commit: r.commit,
		// TODO!(sqs): limit to path
	}
}

func (r *gitCommitResolver) Symbols(args *symbolsArgs) *symbolConnectionResolver {
	return &symbolConnectionResolver{
		first:  args.First,
		query:  args.Query,
		commit: r,
	}
}

type symbolConnectionResolver struct {
	first *int32
	query *string

	commit *gitCommitResolver

	// cache results because they are used by multiple fields
	once    sync.Once
	symbols []*symbolResolver
	err     error
}

func (r *symbolConnectionResolver) limitOrDefault() int {
	if r.first == nil {
		return 100
	}
	return int(*r.first)
}

func (r *symbolConnectionResolver) compute(ctx context.Context) ([]*symbolResolver, error) {
	r.once.Do(func() {
		searchArgs := protocol.SearchArgs{
			CommitID: api.CommitID(r.commit.oid),
			First:    r.limitOrDefault() + 1, // add 1 so we can determine PageInfo.hasNextPage
			Repo:     r.commit.repo.repo.URI,
		}
		if r.query != nil {
			searchArgs.Query = *r.query
		}
		symbols, err := backend.Symbols.ListTags(ctx, searchArgs)
		if err != nil && r.err == nil && ctx.Err() == nil {
			r.err = err
		}
		resolvers := make([]*symbolResolver, 0, len(symbols))
		for _, symbol := range symbols {
			// TODO return the actual language here that we get from ctags
			// it is currently discarded because SymbolInformation has no field for it
			resolver := toSymbolResolver(symbol, "tags", r.commit)
			if resolver != nil {
				resolvers = append(resolvers, resolver)
			}
		}
		r.symbols = append(r.symbols, resolvers...)
	})
	return r.symbols, r.err
}

func toSymbolResolver(symbol lsp.SymbolInformation, lang string, commitResolver *gitCommitResolver) *symbolResolver {
	resolver := &symbolResolver{
		symbol:   symbol,
		language: lang,
	}
	uri, err := uri.Parse(string(symbol.Location.URI))
	if err != nil {
		log15.Warn("Omitting symbol with invalid URI from results.", "uri", symbol.Location.URI, "error", err)
		return nil
	}
	symbolRange := symbol.Location.Range // copy
	resolver.location = &locationResolver{
		resource: &fileResolver{
			commit: commitResolver,
			path:   uri.Fragment,
			stat:   createFileInfo(uri.Fragment, false), // assume the path refers to a file (not dir)
		},
		lspRange: &symbolRange,
	}
	return resolver
}

func (r *symbolConnectionResolver) Nodes(ctx context.Context) ([]*symbolResolver, error) {
	symbols, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if len(r.symbols) > r.limitOrDefault() {
		symbols = symbols[:r.limitOrDefault()]
	}
	return symbols, nil
}

func (r *symbolConnectionResolver) PageInfo(ctx context.Context) (*pageInfo, error) {
	symbols, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return &pageInfo{hasNextPage: len(symbols) > r.limitOrDefault()}, nil
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

func (r *symbolResolver) URL(ctx context.Context) (string, error) {
	url, err := r.location.URL(ctx)
	if err != nil {
		return "", err
	}
	// TODO(sqs): if we have references for this lang, then add "$references" to the URL for convenience
	return url, nil
}

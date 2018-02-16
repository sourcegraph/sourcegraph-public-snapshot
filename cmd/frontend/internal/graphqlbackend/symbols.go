package graphqlbackend

import (
	"context"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/proxy"
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
		inv, err := backend.Repos.GetInventory(ctx, r.commit.repo.repo, api.CommitID(r.commit.oid))
		if err != nil {
			r.err = err
			return
		}

		params := lspext.WorkspaceSymbolParams{
			Limit: r.limitOrDefault() + 1, // add 1 so we can determine PageInfo.hasNextPage
		}
		if r.query != nil {
			params.Query = *r.query
		}

		// HACK(sqs) TODO!(sqs)
		inv.Languages = []*inventory.Lang{{Type: "programming", Name: "tags"}}

		var wg sync.WaitGroup
		var mu sync.Mutex // protects r.symbols and r.err
		for _, lang := range inv.Languages {
			if lang.Type != "programming" {
				continue
			}
			wg.Add(1)
			go func(lang string) {
				defer wg.Done()
				symbols, listErr := backend.Symbols.List(ctx, r.commit.repo.repo.URI, api.CommitID(r.commit.oid), lang, params)
				if listErr != nil {
					if jsonrpc2Err, ok := errors.Cause(listErr).(*jsonrpc2.Error); ok && jsonrpc2Err.Code == proxy.CodeModeNotFound {
						return
					}
				}
				var resolvers []*symbolResolver
				if listErr == nil {
					resolvers = make([]*symbolResolver, 0, len(symbols))
					for _, symbol := range symbols {
						resolver := toSymbolResolver(symbol, lang, r.commit)
						if resolver != nil {
							resolvers = append(resolvers, resolver)
						}
					}
				}
				mu.Lock()
				defer mu.Unlock()
				if listErr != nil && r.err == nil && ctx.Err() == nil {
					r.err = errors.Wrapf(listErr, "Symbols.List for repo %q commit %q lang %q params %+v", r.commit.repo.repo.URI, r.commit.oid, lang, params)
				}
				r.symbols = append(r.symbols, resolvers...)
			}(strings.ToLower(lang.Name))
		}
		wg.Wait()
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

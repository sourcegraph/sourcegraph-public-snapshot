package graphqlbackend

import (
	"context"
	"sync"
	"time"

	"github.com/neelance/parallel"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
)

var mockSearchSymbols func(ctx context.Context, args *repoSearchArgs, query searchquery.Query, limit int) (res []*symbolResolver, err error)

// searchSymbols searches the given repos in parallel for symbols matching the given search query
// it can be used for both search suggestions and search results
//
// May return partial results and an error
func searchSymbols(ctx context.Context, args *repoSearchArgs, query searchquery.Query, limit int) (res []*symbolResolver, err error) {
	if mockSearchSymbols != nil {
		return mockSearchSymbols(ctx, args, query, limit)
	}

	if args.query.Pattern == "" {
		return nil, nil
	}

	params := lspext.WorkspaceSymbolParams{
		Limit: limit,
		Query: args.query.Pattern, // TODO!(sqs): support all options here
	}

	ctx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()

	var (
		run               = parallel.NewRun(20)
		symbolResolversMu sync.Mutex
		symbolResolvers   []*symbolResolver
	)
	for _, repoRevs := range args.repos {
		if ctx.Err() != nil {
			break
		}
		if len(repoRevs.revspecs()) == 0 {
			continue
		}
		run.Acquire()
		go func(repoRevs *repositoryRevisions) {
			defer run.Release()
			inputRev := repoRevs.revspecs()[0]
			commitID, err := backend.Repos.ResolveRev(ctx, repoRevs.repo, inputRev)
			if err != nil {
				run.Error(err)
				return
			}
			symbols, err := backend.Symbols.List(ctx, repoRevs.repo.URI, commitID, "tags", params)
			if err != nil && err != context.Canceled && err != context.DeadlineExceeded && ctx.Err() == nil {
				run.Error(err)
				return
			}
			if len(symbols) > 0 {
				symbolResolversMu.Lock()
				defer symbolResolversMu.Unlock()
				for _, symbol := range symbols {
					commit := &gitCommitResolver{
						repo: &repositoryResolver{repo: repoRevs.repo},
						oid:  gitObjectID(commitID),
						// NOTE: Not all fields are set, for performance.
					}
					if inputRev != "" {
						commit.inputRev = &inputRev
					}

					lang := "" // TODO(sqs): fill this in - need to add a new extension field to lsp.SymbolInformation?
					symbolResolvers = append(symbolResolvers, toSymbolResolver(symbol, lang, commit))
				}
				if len(symbolResolvers) > limit {
					cancel()
				}
			}
		}(repoRevs)
	}
	err = run.Wait()

	if len(symbolResolvers) > limit {
		symbolResolvers = symbolResolvers[:limit]
	}
	return symbolResolvers, err
}

package httpapi

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func serveSymbols(w http.ResponseWriter, r *http.Request) error {
	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	var params struct {
		Query string
	}
	if err := schemaDecoder.Decode(&params, r.URL.Query()); err != nil {
		return err
	}

	symbols, err := langp.DefaultClient.Symbols(r.Context(), &langp.SymbolsQuery{
		Query: params.Query,
		RepoRev: langp.RepoRev{
			Repo:   repo.URI,
			Commit: repoRev.CommitID,
		},
	})
	universeObserve("Symbols", err)
	if err != nil {
		return err
	}

	if params.Query != "" {
		span, _ := opentracing.StartSpanFromContext(r.Context(), "serveSymbols: filter")
		span.SetTag("symbol count", len(symbols.Symbols))

		q := strings.ToLower(params.Query)
		exact, prefix, contains := []*lsp.SymbolInformation{}, []*lsp.SymbolInformation{}, []*lsp.SymbolInformation{}
		for _, s := range symbols.Symbols {
			name := strings.ToLower(s.Name)
			if name == q {
				exact = append(exact, s)
			} else if strings.HasPrefix(name, q) {
				prefix = append(prefix, s)
			} else if strings.Contains(name, q) {
				contains = append(contains, s)
			}
		}
		symbols.Symbols = append(append(exact, prefix...), contains...) // Basic ranking

		span.SetTag("filtered count", len(symbols.Symbols))
		span.Finish()
	}

	const limit = 1000
	if len(symbols.Symbols) > limit {
		symbols.Symbols = symbols.Symbols[:limit]
	}

	return writeJSON(w, symbols)
}

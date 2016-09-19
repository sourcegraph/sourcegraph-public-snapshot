package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
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

	opt := &langp.SymbolsOpt{
		RepoRev: langp.RepoRev{
			Repo:   repo.URI,
			Commit: repoRev.CommitID,
		},
		Query: params.Query,
	}
	symbols, err := langp.DefaultClient.Symbols(r.Context(), opt)
	universeObserve("Symbols", err)
	if err != nil {
		return err
	}

	const limit = 1000
	if len(symbols.Symbols) > limit {
		symbols.Symbols = symbols.Symbols[:limit]
	}

	return writeJSON(w, symbols)
}

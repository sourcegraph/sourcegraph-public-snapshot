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

	symbols, err := langp.DefaultClient.Symbols(r.Context(), &langp.RepoRev{
		Repo:   repo.URI,
		Commit: repoRev.CommitID,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, symbols)
}

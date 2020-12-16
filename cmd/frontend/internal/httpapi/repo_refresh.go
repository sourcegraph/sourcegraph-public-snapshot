package httpapi

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/handlerutil"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
)

func serveRepoRefresh(w http.ResponseWriter, r *http.Request) error {
	repo, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}
	_, err = repoupdater.DefaultClient.EnqueueRepoUpdate(context.Background(), repo.Name)
	return err
}

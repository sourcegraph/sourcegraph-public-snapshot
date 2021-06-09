package httpapi

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/handlerutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func serveRepoRefresh(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	repo, err := handlerutil.GetRepo(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	_, err = gitserver.DefaultClient.RequestRepoUpdate(ctx, repo.Name, 10*time.Second)
	return err
}

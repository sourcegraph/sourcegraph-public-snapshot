package httpapi

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/handlerutil"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

func serveRepoRefresh(w http.ResponseWriter, r *http.Request) error {
	repo, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}
	repoMeta, err := repoupdater.DefaultClient.RepoLookup(context.Background(), protocol.RepoLookupArgs{
		Repo:         repo.Name,
		ExternalRepo: repo.ExternalRepo,
	})
	if err != nil {
		return err
	}
	return repoupdater.DefaultClient.EnqueueRepoUpdate(context.Background(), gitserver.Repo{
		Name: repo.Name,
		URL:  repoMeta.Repo.VCS.URL,
	})
}

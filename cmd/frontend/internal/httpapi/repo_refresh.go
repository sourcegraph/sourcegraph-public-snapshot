package httpapi

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/handlerutil"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func serveRepoRefresh(w http.ResponseWriter, r *http.Request) error {
	log15.Info("serveRepoRefresh 0")
	repo, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}
	log15.Info("serveRepoRefresh 1", "repo", repo)
	repoMeta, err := repoupdater.DefaultClient.RepoLookup(context.Background(), protocol.RepoLookupArgs{
		Repo:         repo.Name,
		ExternalRepo: repo.ExternalRepo,
	})
	if err != nil {
		log15.Info("serveRepoRefresh 1 ERROR", "err", err)
		return err
	}
	log15.Info("serveRepoRefresh 2", "repoMeta", repoMeta)
	_, err = repoupdater.DefaultClient.EnqueueRepoUpdate(context.Background(), gitserver.Repo{
		Name: repo.Name,
		URL:  repoMeta.Repo.VCS.URL,
	})
	if err != nil {
		log15.Info("serveRepoRefresh 2 ERROR", "err", err)
	}
	log15.Info("serveRepoRefresh 3")
	return err
}

package httpapi

import (
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func serveReposUpdate(w http.ResponseWriter, r *http.Request) error {
	list, err := gitserver.DefaultClient.List()
	if err != nil {
		return err
	}

	for _, uri := range list {
		err := localstore.Repos.TryInsertNew(r.Context(), uri, "", false, false)
		if err != nil {
			log15.Warn("TryInsertNew failed on repos-update", "uri", uri, "err", err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("OK"))
	return nil
}

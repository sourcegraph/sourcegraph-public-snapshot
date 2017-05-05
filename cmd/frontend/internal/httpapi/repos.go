package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

func serveRepoCreate(w http.ResponseWriter, r *http.Request) error {
	// legacy support for Chrome extension
	var data struct {
		Op struct {
			New struct {
				URI string
			}
		}
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return err
	}
	if _, err := backend.Repos.GetByURI(r.Context(), data.Op.New.URI); err != nil {
		return err
	}
	w.Write([]byte("OK"))
	return nil
}

type ensureOpt struct {
	Index bool `schema:"index"`
}

var decoder = schema.NewDecoder()

// serveRepoEnsure ensures the repositories specified in the request
// body are fully available. If they don't yet exist, they are added,
// cloned, and an indexing job is enqueued, but this function does not
// block on those operations. The endpoint returns a list of
// repositories that were not yet found / cloned. Callers should poll
// this function until the returned list is empty. Note that there is
// currently no guarantee on when the repository is indexed, just on
// when the indexing job is enqueued.
func serveRepoEnsure(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var repoURIs []string
	if err := json.NewDecoder(r.Body).Decode(&repoURIs); err != nil {
		return err
	}

	var notfound []string
	for _, uri := range repoURIs {
		repo, err := backend.Repos.GetByURI(ctx, uri)
		if err != nil {
			notfound = append(notfound, uri)
			continue
		}
		if _, err := gitcmd.Open(repo).ResolveRevision(ctx, "HEAD"); err != nil {
			notfound = append(notfound, uri)
		}
	}
	if notfound == nil {
		fmt.Fprintln(w, "[]")
		return nil
	}

	for _, uri := range repoURIs {
		if err := backend.Repos.RefreshIndex(ctx, uri); err != nil {
			log15.Warn("index refresh failed on repo ensure", "uri", uri, "err", err)
		}
	}

	return json.NewEncoder(w).Encode(notfound)
}

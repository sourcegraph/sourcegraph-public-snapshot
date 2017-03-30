package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
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

// serveRepoEnsure ensures the repositories specified in the request
// body. If they don't yet exist, they are added and cloned, but this
// function does not block on those operations. Callers should poll
// this function.
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

	return json.NewEncoder(w).Encode(notfound)
}

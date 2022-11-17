package ui

import (
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

// serveStreamBlame returns a HTTP handler that streams back the results of running
// git blame with the --incremental flag. It will stream back to the client the most
// recent hunks first and will gradually reach the oldests, or not if we timeout
// before that.
//
//	http://localhost:3080/github.com/gorilla/mux/-/stream-blame/mux.go
func serveStreamBlame(db database.DB, gitserverClient gitserver.Client) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		flags := featureflag.FromContext(r.Context())
		if !flags.GetBoolOr("enable-streaming-git-blame", false) {
			w.WriteHeader(404)
			return
		}

		var common *Common
		for {
			// newCommon provides various repository handling features that we want, so
			// we use it but discard the resulting structure. It provides:
			//
			// - Repo redirection
			// - Gitserver content updating
			// - Consistent error handling (permissions, revision not found, repo not found, etc).
			//
			common, err = newCommon(w, r, db, globals.Branding().BrandName, noIndex, serveError)
			if err != nil {
				return err
			}
			if common == nil {
				return nil // request was handled
			}
			if common.Repo == nil {
				// Repository is cloning.
				time.Sleep(5 * time.Second)
				continue
			}
			break
		}
		requestedPath := mux.Vars(r)["Path"]

		streamWriter, err := streamhttp.NewWriter(w)
		if err != nil {
			// tr.SetError(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if strings.HasPrefix(requestedPath, "/") {
			requestedPath = strings.TrimLeft(requestedPath, "/")
		}

		hunkReader, err := gitserverClient.StreamBlameFile(r.Context(), authz.DefaultSubRepoPermsChecker, common.Repo.Name, requestedPath, &gitserver.BlameOptions{
			NewestCommit: common.CommitID,
		})
		if err != nil {
			return err
		}

		for {
			hunk, done, err := hunkReader.Read()
			if err != nil {
				http.Error(w, html.EscapeString(err.Error()), http.StatusInternalServerError)
				return err
			}
			if done {
				streamWriter.Event("done", map[string]any{})
				return nil
			}
			if err := streamWriter.Event("hunk", hunk); err != nil {
				return err
			}
		}
	}
}

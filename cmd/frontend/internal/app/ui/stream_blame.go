package ui

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Stream blame
//
//	http://localhost:3080/github.com/gorilla/mux/-/stream-blame/mux.go
func serveStreamBlame(db database.DB, gitserverClient gitserver.Client) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		// ---------------------- START OF COPY&PASTED -----------------------
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

		flusher, ok := w.(http.Flusher)
		if !ok {
			return errors.New("http flushing not supported")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")

		w.Header().Set("X-Accel-Buffering", "no")
		// ---------------------- END OF COPY&PASTED -----------------------

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
				// done
				return nil
			}

			encoded, err := json.Marshal(hunk)
			if err != nil {
				return err
			}

			encoded = append(encoded, []byte("\n")...)
			if _, err = w.Write(encoded); err != nil {
				return err
			}

			flusher.Flush()
		}
	}
}

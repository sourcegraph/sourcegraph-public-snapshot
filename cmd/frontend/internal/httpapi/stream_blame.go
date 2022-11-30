package httpapi

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/handlerutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// handleStreamBlame returns a HTTP handler that streams back the results of running
// git blame with the --incremental flag. It will stream back to the client the most
// recent hunks first and will gradually reach the oldests, or not if we timeout
// before that.
func handleStreamBlame(logger log.Logger, db database.DB, gitserverClient gitserver.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flags := featureflag.FromContext(r.Context())
		if !flags.GetBoolOr("enable-streaming-git-blame", false) {
			w.WriteHeader(404)
			return
		}
		tr, ctx := trace.New(r.Context(), "blame.Stream", "")
		defer tr.Finish()
		r = r.WithContext(ctx)

		if _, ok := mux.Vars(r)["Repo"]; !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		repo, commitID, err := handlerutil.GetRepoAndRev(r.Context(), logger, db, mux.Vars(r))
		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.HasType(err, &gitserver.RepoNotCloneableErr{}) {
			if errcode.IsNotFound(err) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if errcode.IsNotFound(err) || errcode.IsBlocked(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errcode.IsUnauthorized(err) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		requestedPath := mux.Vars(r)["Path"]
		streamWriter, err := streamhttp.NewWriter(w)
		if err != nil {
			tr.SetError(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Log events to trace
		streamWriter.StatHook = func(stat streamhttp.WriterStat) {
			fields := []otlog.Field{
				otlog.String("streamhttp.Event", stat.Event),
				otlog.Int("bytes", stat.Bytes),
				otlog.Int64("duration_ms", stat.Duration.Milliseconds()),
			}
			if stat.Error != nil {
				fields = append(fields, otlog.Error(stat.Error))
			}
			tr.LogFields(fields...)
		}

		requestedPath = strings.TrimPrefix(requestedPath, "/")

		hunkReader, err := gitserverClient.StreamBlameFile(r.Context(), authz.DefaultSubRepoPermsChecker, repo.Name, requestedPath, &gitserver.BlameOptions{
			NewestCommit: commitID,
		})
		if err != nil {
			tr.SetError(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		parentsCache := map[api.CommitID][]api.CommitID{}

		for {
			hunks, done, err := hunkReader.Read()
			if err != nil {
				tr.SetError(err)
				http.Error(w, html.EscapeString(err.Error()), http.StatusInternalServerError)
				return
			}
			if done {
				streamWriter.Event("done", map[string]any{})
				return
			}

			blameResponses := make([]BlameHunkResponse, 0, len(hunks))
			for _, h := range hunks {
				var parents []api.CommitID
				if p, ok := parentsCache[h.CommitID]; ok {
					parents = p
				} else {
					c, err := gitserverClient.GetCommit(ctx, repo.Name, h.CommitID, gitserver.ResolveRevisionOptions{}, authz.DefaultSubRepoPermsChecker)
					if err != nil {
						tr.SetError(err)
						http.Error(w, html.EscapeString(err.Error()), http.StatusInternalServerError)
						return
					}
					parents = c.Parents
					parentsCache[h.CommitID] = c.Parents
				}

				blameResponse := BlameHunkResponse{
					StartLine: h.StartLine,
					EndLine:   h.EndLine,
					CommitID:  h.CommitID,
					Author:    h.Author,
					Message:   h.Message,
					Filename:  h.Filename,
					Commit: BlameHunkCommitResponse{
						Parents: parents,
						URL:     fmt.Sprintf("%s/-/commit/%s", repo.URI, h.CommitID),
					},
				}
				blameResponses = append(blameResponses, blameResponse)
			}

			if err := streamWriter.Event("hunk", blameResponses); err != nil {
				tr.SetError(err)
				http.Error(w, html.EscapeString(err.Error()), http.StatusInternalServerError)
				return
			}
		}
	}
}

type BlameHunkResponse struct {
	api.CommitID `json:"commitID"`

	StartLine int                     `json:"startLine"` // 1-indexed start line number
	EndLine   int                     `json:"endLine"`   // 1-indexed end line number
	Author    gitdomain.Signature     `json:"author"`
	Message   string                  `json:"message"`
	Filename  string                  `json:"filename"`
	Commit    BlameHunkCommitResponse `json:"commit"`
}

type BlameHunkCommitResponse struct {
	Parents []api.CommitID `json:"parents"`
	URL     string         `json:"url"`
}

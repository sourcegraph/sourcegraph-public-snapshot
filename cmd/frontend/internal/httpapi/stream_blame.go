package httpapi

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/handlerutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
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
		tr, ctx := trace.New(r.Context(), "blame.Stream")
		defer tr.End()
		r = r.WithContext(ctx)

		if _, ok := mux.Vars(r)["Repo"]; !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		repo, commitID, err := handlerutil.GetRepoAndRev(r.Context(), logger, db, mux.Vars(r))
		if err != nil {
			if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
				w.WriteHeader(http.StatusNotFound)
			} else if errors.HasType(err, &gitserver.RepoNotCloneableErr{}) && errcode.IsNotFound(err) {
				w.WriteHeader(http.StatusNotFound)
			} else if errcode.IsNotFound(err) || errcode.IsBlocked(err) {
				w.WriteHeader(http.StatusNotFound)
			} else if errcode.IsUnauthorized(err) {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
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
			attrs := []attribute.KeyValue{
				attribute.String("streamhttp.Event", stat.Event),
				attribute.Int("bytes", stat.Bytes),
				attribute.Int64("duration_ms", stat.Duration.Milliseconds()),
			}
			if stat.Error != nil {
				attrs = append(attrs, trace.Error(stat.Error))
			}
			tr.AddEvent("write", attrs...)
		}

		requestedPath = strings.TrimPrefix(requestedPath, "/")

		hunkReader, err := gitserverClient.StreamBlameFile(r.Context(), repo.Name, requestedPath, &gitserver.BlameOptions{
			NewestCommit: commitID,
		})
		if err != nil {
			tr.SetError(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer hunkReader.Close()

		parentsCache := map[api.CommitID][]api.CommitID{}

		for {
			h, err := hunkReader.Read()
			if errors.Is(err, io.EOF) {
				streamWriter.Event("done", map[string]any{})
				return
			} else if err != nil {
				tr.SetError(err)
				http.Error(w, html.EscapeString(err.Error()), http.StatusInternalServerError)
				return
			}

			var parents []api.CommitID
			if p, ok := parentsCache[h.CommitID]; ok {
				parents = p
			} else {
				c, err := gitserverClient.GetCommit(ctx, repo.Name, h.CommitID, gitserver.ResolveRevisionOptions{
					// The list of hunks and commit IDs came from gitserver, that
					// means the commit will exist and we don't need to ensure
					// the revision exists.
					NoEnsureRevision: true,
				})
				if err != nil {
					tr.SetError(err)
					http.Error(w, html.EscapeString(err.Error()), http.StatusInternalServerError)
					return
				}
				parents = c.Parents
				parentsCache[h.CommitID] = c.Parents
			}

			user, err := db.Users().GetByVerifiedEmail(ctx, h.Author.Email)
			if err != nil && !errcode.IsNotFound(err) {
				tr.SetError(err)
				http.Error(w, html.EscapeString(err.Error()), http.StatusInternalServerError)
				return
			}

			var blameHunkUserResponse *BlameHunkUserResponse
			if user != nil {
				displayName := &user.DisplayName
				if *displayName == "" {
					displayName = nil
				}
				avatarURL := &user.AvatarURL
				if *avatarURL == "" {
					avatarURL = nil
				}

				blameHunkUserResponse = &BlameHunkUserResponse{
					Username:    user.Username,
					DisplayName: displayName,
					AvatarURL:   avatarURL,
				}
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
					URL:     fmt.Sprintf("%s/-/commit/%s", repo.Name, h.CommitID),
				},
				User: blameHunkUserResponse,
			}

			if err := streamWriter.Event("hunk", []BlameHunkResponse{blameResponse}); err != nil {
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
	User      *BlameHunkUserResponse  `json:"user,omitempty"`
}

type BlameHunkCommitResponse struct {
	Parents []api.CommitID `json:"parents"`
	URL     string         `json:"url"`
}

type BlameHunkUserResponse struct {
	Username    string  `json:"username"`
	DisplayName *string `json:"displayName"`
	AvatarURL   *string `json:"avatarURL"`
}

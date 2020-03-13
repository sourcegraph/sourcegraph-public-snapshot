package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"gopkg.in/inconshreveable/log15.v2"
)

func NewProxy() (*httpapi.LSIFServerProxy, error) {
	return &httpapi.LSIFServerProxy{
		UploadHandler: http.HandlerFunc(uploadProxyHandler()),
	}, nil
}

func uploadProxyHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		repoName := q.Get("repository")
		commit := q.Get("commit")
		root := q.Get("root")
		indexerName := q.Get("indexerName")
		ctx := r.Context()

		repo, ok := ensureRepoAndCommitExist(ctx, w, repoName, commit)
		if !ok {
			return
		}

		// 🚨 SECURITY: Ensure we return before proxying to the lsif-server upload
		// endpoint. This endpoint is unprotected, so we need to make sure the user
		// provides a valid token proving contributor access to the repository.
		if conf.Get().LsifEnforceAuth {
			if canBypassAuth := isSiteAdmin(ctx); !canBypassAuth {
				if authorized := enforceAuth(ctx, w, r, repoName); !authorized {
					return
				}
			}
		}

		uploadID, queued, err := client.DefaultClient.Upload(ctx, &struct {
			RepoID      api.RepoID
			Commit      graphqlbackend.GitObjectID
			Root        string
			IndexerName string
			Body        io.ReadCloser
		}{
			RepoID:      repo.ID,
			Commit:      graphqlbackend.GitObjectID(commit),
			Root:        root,
			IndexerName: indexerName,
			Body:        r.Body,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return id as a string to maintain backwards compatibility with src-cli
		payload, err := json.Marshal(map[string]string{"id": strconv.FormatInt(uploadID, 10)})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if queued {
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		_, _ = w.Write(payload)
	}
}

func ensureRepoAndCommitExist(ctx context.Context, w http.ResponseWriter, repoName, commit string) (*types.Repo, bool) {
	repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		if errcode.IsNotFound(err) {
			http.Error(w, fmt.Sprintf("unknown repository %q", repoName), http.StatusNotFound)
			return nil, false
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, false
	}

	if _, err := backend.Repos.ResolveRev(ctx, repo, commit); err != nil {
		if gitserver.IsRevisionNotFound(err) {
			http.Error(w, fmt.Sprintf("unknown commit %q", commit), http.StatusNotFound)
			return nil, false
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, false
	}

	return repo, true
}

func isSiteAdmin(ctx context.Context) bool {
	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == db.ErrNoCurrentUser {
			return false
		}

		log15.Error("lsif-server proxy: failed to get up current user", "error", err)
		return false
	}

	return user != nil && user.SiteAdmin
}

func enforceAuth(ctx context.Context, w http.ResponseWriter, r *http.Request, repoName string) bool {
	validatorByCodeHost := map[string]func(context.Context, http.ResponseWriter, *http.Request, string) (int, error){
		"github.com": enforceAuthGithub,
	}

	for codeHost, validator := range validatorByCodeHost {
		if strings.HasPrefix(repoName, codeHost) {
			if status, err := validator(ctx, w, r, repoName); err != nil {
				http.Error(w, err.Error(), status)
				return false
			}

			return true
		}
	}

	http.Error(w, "verification not supported for code host - see https://github.com/sourcegraph/sourcegraph/issues/4967", http.StatusUnprocessableEntity)
	return false
}

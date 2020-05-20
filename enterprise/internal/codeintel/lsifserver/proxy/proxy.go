package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	codeinteldb "github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/enqueuer"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func NewProxy(db codeinteldb.DB, bundleManagerClient bundles.BundleManagerClient) (*httpapi.LSIFServerProxy, error) {
	return &httpapi.LSIFServerProxy{
		UploadHandler: http.HandlerFunc(uploadProxyHandler(enqueuer.NewEnqueuer(db, bundleManagerClient))),
	}, nil
}

func uploadProxyHandler(enqueuer *enqueuer.Enqueuer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		repoName := q.Get("repository")
		commit := q.Get("commit")
		ctx := r.Context()

		// We only need to translate the repository id and check for auth on the first
		// upload request. If there is an upload identifier, this check has already
		// been performed (or someone is uploading to an unknown identifier, which will
		// naturally result in an error response).
		if q.Get("uploadId") == "" {
			repo, ok := ensureRepoAndCommitExist(ctx, w, repoName, commit)
			if !ok {
				return
			}

			// translate repository id to something that the precise-code-intel-api-server
			// can reconcile in the database
			q.Del("repository")
			q.Set("repositoryId", fmt.Sprintf("%d", repo.ID))

			// 🚨 SECURITY: Ensure we return before proxying to the precise-code-intel-api-server upload
			// endpoint. This endpoint is unprotected, so we need to make sure the user provides a valid
			// token proving contributor access to the repository.
			if conf.Get().LsifEnforceAuth {
				if canBypassAuth := isSiteAdmin(ctx); !canBypassAuth {
					if authorized := enforceAuth(ctx, w, r, repoName); !authorized {
						return
					}
				}
			}
		}

		r.URL.RawQuery = q.Encode()
		enqueuer.HandleEnqueue(w, r)
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

		log15.Error("precise-code-intel proxy: failed to get up current user", "error", err)
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

func makeUploadRequest(host string, q url.Values, body io.Reader) (*http.Request, error) {
	url, err := url.Parse(fmt.Sprintf("%s/upload", host))
	if err != nil {
		return nil, err
	}
	url.RawQuery = q.Encode()

	return http.NewRequest("POST", url.String(), body)
}

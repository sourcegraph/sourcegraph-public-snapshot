package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func NewProxy() (*httpapi.LSIFServerProxy, error) {
	url, err := url.Parse(lsifserver.ServerURLFromEnv)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	return &httpapi.LSIFServerProxy{
		UploadHandler: http.HandlerFunc(uploadProxyHandler(proxy)),
	}, nil
}

func uploadProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		repoName := q.Get("repository")
		commit := q.Get("commit")
		ctx := r.Context()

		if !ensureRepoAndCommitExist(ctx, w, repoName, commit) {
			return
		}

		// 🚨 SECURITY: Ensure we return before proxying to the lsif-server upload
		// endpoint. This endpoint is unprotected, so we need to make sure the user
		// provides a valid token proving contributor access to the repository.
		if conf.Get().LsifEnforceAuth && !enforceAuth(ctx, w, r, repoName) {
			return
		}

		resp, err := client.DefaultClient.BuildAndTraceRequest(
			ctx,
			"POST",
			"/upload",
			r.URL.Query(),
			r.Body,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

func ensureRepoAndCommitExist(ctx context.Context, w http.ResponseWriter, repoName, commit string) bool {
	repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		if errcode.IsNotFound(err) {
			http.Error(w, fmt.Sprintf("unknown repository %q", repoName), http.StatusNotFound)
			return false
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	if _, err := backend.Repos.ResolveRev(ctx, repo, commit); err != nil {
		if gitserver.IsRevisionNotFound(err) {
			http.Error(w, fmt.Sprintf("unknown commit %q", commit), http.StatusNotFound)
			return false
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	return true
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

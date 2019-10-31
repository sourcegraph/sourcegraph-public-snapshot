package httpapi

import (
	"strings"

	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

var apiURL = url.URL{Scheme: "https", Host: "api.github.com"}

func lsifProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = mux.Vars(r)["rest"]
		p.ServeHTTP(w, r)
	}
}

func lsifUploadProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		repoName := q.Get("repository")
		commit := q.Get("commit")

		repo, err := backend.Repos.GetByName(r.Context(), api.RepoName(repoName))
		if err != nil {
			http.Error(w, "Unknown repository.", http.StatusNotFound)
			return
		}

		_, err = backend.Repos.ResolveRev(r.Context(), repo, commit)
		if err != nil {
			http.Error(w, "Unknown commit.", http.StatusNotFound)
			return
		}

		if conf.Get().LsifEnforceAuth {
			err, status := lsifEnforceAuth(w, r, repoName)
			if err != nil {
				http.Error(w, err.Error(), status)
			}
		}

		r.URL.Path = "upload"
		p.ServeHTTP(w, r)
	}
}

func lsifEnforceAuth(w http.ResponseWriter, r *http.Request, repoName string) (error, int) {
	validatorByCodeHost := map[string]func(http.ResponseWriter, *http.Request, string) (error, int){
		"github.com": lsifEnforceAuthGithub,
	}

	for codeHost, validator := range validatorByCodeHost {
		if strings.HasPrefix(repoName, codeHost) {
			return validator(w, r, repoName)
		}
	}

	return errors.New("Verification not supported for code host. See https://github.com/sourcegraph/sourcegraph/issues/4967"), http.StatusUnprocessableEntity
}

func lsifEnforceAuthGithub(w http.ResponseWriter, r *http.Request, repoName string) (error, int) {
	nameWithOwner := strings.TrimPrefix(repoName, "github.com/")
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return errors.New("Invalid GitHub repository: nameWithOwner=" + nameWithOwner), http.StatusNotFound
	}

	q := r.URL.Query()
	githubToken := q.Get("github_token")
	if githubToken == "" {
		return errors.New("Must provide github_token."), http.StatusUnauthorized
	}

	client := github.NewClient(&apiURL, githubToken, nil)
	repo, err := client.GetRepository(r.Context(), owner, name)
	if err != nil {
		return errors.Wrap(err, "Unable to get repository permissions"), http.StatusNotFound
	}

	if !(repo.ViewerPermission == "ADMIN" || repo.ViewerPermission == "MAINTAIN" || repo.ViewerPermission == "WRITE") {
		return errors.New("You do not have write permission to the repository."), http.StatusUnauthorized
	}

	return nil, 0
}

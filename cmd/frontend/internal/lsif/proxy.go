package lsif

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

func ProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = mux.Vars(r)["rest"]
		p.ServeHTTP(w, r)
	}
}

func UploadProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := backend.Repos.GetByName(r.Context(), api.RepoName(r.URL.Query().Get("repository")))
		if err != nil {
			http.Error(w, "Unknown repository.", http.StatusNotFound)
			return
		}

		if conf.Get().LsifEnforceAuth && !validateLsifAuth(w, r) {
			return
		}

		r.URL.Path = "upload"
		p.ServeHTTP(w, r)
	}
}

func validateLsifAuth(w http.ResponseWriter, r *http.Request) bool {
	repository := r.URL.Query().Get("repository")
	if !strings.HasPrefix(repository, "github.com") {
		http.Error(w, "Only github.com repositories support verification. See https://github.com/sourcegraph/sourcegraph/issues/4967", http.StatusUnprocessableEntity)
		return false
	}
	nameWithOwner := strings.TrimPrefix(repository, "github.com/")
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		http.Error(w, "Invalid GitHub repository: nameWithOwner="+nameWithOwner, http.StatusNotFound)
		return false
	}
	githubToken := r.URL.Query().Get("github_token")
	if githubToken == "" {
		http.Error(w, "Must provide github_token.", http.StatusUnauthorized)
		return false
	}
	client := github.NewClient(&apiURL, githubToken, nil)
	repo, err := client.GetRepository(r.Context(), owner, name)
	if err != nil {
		http.Error(w, errors.Wrap(err, "Unable to get repository permissions").Error(), http.StatusNotFound)
		return false
	}

	if !(repo.ViewerPermission == "ADMIN" || repo.ViewerPermission == "MAINTAIN" || repo.ViewerPermission == "WRITE") {
		http.Error(w, "You do not have write permission to the repository.", http.StatusUnauthorized)
		return false
	}

	return true
}

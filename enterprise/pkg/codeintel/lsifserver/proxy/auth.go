package proxy

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func enforceAuth(w http.ResponseWriter, r *http.Request, repoName string) (error, int) {
	validatorByCodeHost := map[string]func(http.ResponseWriter, *http.Request, string) (error, int){
		"github.com": enforceAuthGithub,
	}

	for codeHost, validator := range validatorByCodeHost {
		if strings.HasPrefix(repoName, codeHost) {
			return validator(w, r, repoName)
		}
	}

	return errors.New("verification not supported for code host - see https://github.com/sourcegraph/sourcegraph/issues/4967"), http.StatusUnprocessableEntity
}

var githubURL = url.URL{Scheme: "https", Host: "api.github.com"}

func enforceAuthGithub(w http.ResponseWriter, r *http.Request, repoName string) (error, int) {
	nameWithOwner := strings.TrimPrefix(repoName, "github.com/")
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return errors.New("invalid GitHub repository: nameWithOwner=" + nameWithOwner), http.StatusNotFound
	}

	q := r.URL.Query()
	githubToken := q.Get("github_token")
	if githubToken == "" {
		return errors.New("must provide github_token"), http.StatusUnauthorized
	}

	client := github.NewClient(&githubURL, githubToken, nil)
	repo, err := client.GetRepository(r.Context(), owner, name)
	if err != nil {
		return errors.Wrap(err, "unable to get repository permissions"), http.StatusNotFound
	}

	switch repo.ViewerPermission {
	case "ADMIN", "MAINTAIN", "WRITE":
		return nil, 0
	default:
		return errors.New("you do not have write permission to the repository"), http.StatusUnauthorized
	}
}

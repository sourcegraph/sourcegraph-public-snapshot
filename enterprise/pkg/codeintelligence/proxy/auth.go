package proxy

import (
	"net/http"
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

	return errors.New("Verification not supported for code host. See https://github.com/sourcegraph/sourcegraph/issues/4967"), http.StatusUnprocessableEntity
}

func enforceAuthGithub(w http.ResponseWriter, r *http.Request, repoName string) (error, int) {
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

package proxy

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

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

var githubURL = url.URL{Scheme: "https", Host: "api.github.com"}

func enforceAuthGithub(ctx context.Context, w http.ResponseWriter, r *http.Request, repoName string) (int, error) {
	nameWithOwner := strings.TrimPrefix(repoName, "github.com/")
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return http.StatusNotFound, errors.New("invalid GitHub repository: nameWithOwner=" + nameWithOwner)
	}

	q := r.URL.Query()
	githubToken := q.Get("github_token")
	if githubToken == "" {
		return http.StatusUnauthorized, errors.New("must provide github_token")
	}

	client := github.NewClient(&githubURL, githubToken, nil)
	repo, err := client.GetRepository(ctx, owner, name)
	if err != nil {
		return http.StatusNotFound, errors.Wrap(err, "unable to get repository permissions")
	}

	switch repo.ViewerPermission {
	case "ADMIN", "MAINTAIN", "WRITE":
		return 0, nil
	default:
		return http.StatusUnauthorized, errors.New("you do not have write permission to the repository")
	}
}

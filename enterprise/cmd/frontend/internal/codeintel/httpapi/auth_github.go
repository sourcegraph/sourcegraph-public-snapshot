package httpapi

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

var (
	ErrGitHubMissingToken = errors.New("must provide github_token")
	ErrGitHubUnauthorized = errors.New("you do not have write permission to this GitHub repository")

	githubURL = url.URL{Scheme: "https", Host: "api.github.com"}
)

func enforceAuthGithub(ctx context.Context, w http.ResponseWriter, r *http.Request, repoName string) (int, error) {
	nameWithOwner := strings.TrimPrefix(repoName, "github.com/")
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return http.StatusNotFound, errors.New("invalid GitHub repository: nameWithOwner=" + nameWithOwner)
	}

	q := r.URL.Query()
	githubToken := q.Get("github_token")
	if githubToken == "" {
		return http.StatusUnauthorized, ErrGitHubMissingToken
	}

	client := github.NewV3Client(&githubURL, &auth.OAuthBearerToken{Token: githubToken}, nil)

	// If the given token is a GitHub App installation token, then we use the
	// https://developer.github.com/v3/apps/installations/#list-repositories
	// endpoint to see if the associated GitHub App has been installed on the
	// given  repository.
	//
	// One example of this is the built-in GITHUB_TOKEN in GitHub Actions:
	// https://help.github.com/en/actions/automating-your-workflow-with-github-actions/authenticating-with-the-github_token#about-the-github_token-secret
	//
	// We hold on to the error value if one occurs here as we want to perform
	// a fallback attempt in case the supplied token is not a GitHub App.
	repos, appRequestErr := client.ListInstallationRepositories(ctx)
	if appRequestErr == nil {
		for _, repo := range repos {
			if repo.NameWithOwner == nameWithOwner {
				return 0, nil
			}
		}

		return http.StatusUnauthorized, ErrGitHubUnauthorized
	}

	// If the given token is a personal access token, then we use the
	// https://developer.github.com/v3/repos/#get endpoint to see if the user
	// has write access to the given repository.
	repo, userRequestErr := client.GetRepository(ctx, owner, name)
	if userRequestErr == nil {
		switch repo.ViewerPermission {
		case "ADMIN", "MAINTAIN", "WRITE":
			return 0, nil
		}
	}

	if appRequestErr != nil && userRequestErr != nil {
		// Unable to make either request successfully
		return http.StatusInternalServerError, multierror.Append(
			errors.Wrap(appRequestErr, "failed to list app repositories"),
			errors.Wrap(userRequestErr, "failed to list user repositories"),
		)
	}

	// At least one request was successful and we weren't authenticated
	return http.StatusUnauthorized, ErrGitHubUnauthorized
}

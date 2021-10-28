package httpapi

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

var (
	ErrGitHubMissingToken = errors.New("must provide github_token")
	ErrGitHubUnauthorized = errors.New("you do not have write permission to this GitHub repository")
)

func enforceAuthViaGitHub(ctx context.Context, r *http.Request, repoName string) (int, error) {
	q := r.URL.Query()
	githubToken := q.Get("github_token")
	if githubToken == "" {
		return http.StatusUnauthorized, ErrGitHubMissingToken
	}

	return enforceAuthViaGitHubWithClient(ctx, repoName, github.NewV3Client(
		&url.URL{Scheme: "https", Host: "api.github.com"},
		&auth.OAuthBearerToken{Token: githubToken},
		nil,
	))
}

func enforceAuthViaGitHubWithClient(ctx context.Context, repoName string, client GitHubClient) (int, error) {
	nameWithOwner := strings.TrimPrefix(repoName, "github.com/")
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return http.StatusNotFound, errors.New("invalid GitHub repository: nameWithOwner=" + nameWithOwner)
	}

	// First, try to use the given token as a GitHub App installation token via the following endpoint:
	// https://developer.github.com/v3/apps/installations/#list-repositories. If this response succeeds,
	// then we check to see whether or not the target repository is in the response list. For additional
	// reference, see: https://bit.ly/3GyFzDE.
	installationRepositories, installationRepositoriesErr := client.ListInstallationRepositories(ctx)
	if installationRepositoriesErr == nil {
		for _, repo := range installationRepositories {
			if repo.NameWithOwner == nameWithOwner {
				// Authenticated
				return 0, nil
			}
		}
	}

	// TODO - return if not a bad token type error

	// If the first call did not successfully auth the current user, ry to use the same token as a user
	// token via the following endpoint: https://developer.github.com/v3/repos/#get. If this request
	// succeeds, we check the permission of the user against that repository.
	repository, repositoryErr := client.GetRepository(ctx, owner, name)
	if repositoryErr == nil {
		switch repository.ViewerPermission {
		case "ADMIN", "MAINTAIN", "WRITE":
			// Authenticated
			return 0, nil
		}
	}

	// TODO - what to do at this point?

	// // At this point, we didn't successfully auth the current user with either method. Now, we want to
	// // distinguish between an unauthenticated request and one that failed one of the previous requests.

	// if installationRepositoriesErr!=nil&&repositoryErr != nil {
	// 	// Unable to make either request successfully
	// 	return http.StatusInternalServerError, multierror.Append(
	// 		errors.Wrap(installationRepositoriesErr, "failed to list app repositories"),
	// 		errors.Wrap(repositoryErr, "failed to list user repositories"),
	// 	)
	// }

	return http.StatusUnauthorized, ErrGitHubUnauthorized
}

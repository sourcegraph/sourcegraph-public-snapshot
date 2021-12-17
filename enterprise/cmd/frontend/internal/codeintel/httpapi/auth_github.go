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

	githubURL = &url.URL{Scheme: "https", Host: "api.github.com"}
)

func enforceAuthViaGitHub(ctx context.Context, query url.Values, repoName string) (int, error) {
	githubToken := query.Get("github_token")
	if githubToken == "" {
		return http.StatusUnauthorized, ErrGitHubMissingToken
	}

	if author, err := checkGitHubPermissions(ctx, repoName, github.NewV3Client(githubURL, &auth.OAuthBearerToken{Token: githubToken}, nil)); err != nil {
		return http.StatusInternalServerError, err
	} else if !author {
		return http.StatusUnauthorized, ErrGitHubUnauthorized
	}

	return 0, nil
}

var _ AuthValidator = enforceAuthViaGitHub

func checkGitHubPermissions(ctx context.Context, repoName string, client GitHubClient) (bool, error) {
	nameWithOwner := strings.TrimPrefix(repoName, "github.com/")

	if author, wrongTokenType, err := checkGitHubAppInstallationPermissions(ctx, nameWithOwner, client); !wrongTokenType {
		return author, err
	}

	return checkGitHubUserRepositoryPermissions(ctx, nameWithOwner, client)
}

// checkGitHubAppInstallationPermissions attempts to use the given client as if it's authorized as
// a GitHub app installation with access to certain repositories. If this client is authorized as a
// user instead, then wrongTokenType will be true. Otherwise, we check if the given name and owner
// is present in set of visible repositories, indicating authorship of the user initiating the current
// upload request.
func checkGitHubAppInstallationPermissions(ctx context.Context, nameWithOwner string, client GitHubClient) (author bool, wrongTokenType bool, _ error) {
	installationRepositories, err := client.ListInstallationRepositories(ctx)
	if err != nil {
		// A 403 error with this text indicates that the supplied token is a user token and not
		// an app installation token. We'll send back a special flag to the caller to inform them
		// that they should fall back to hitting the repository endpoint as the user.
		if githubErr, ok := err.(*github.APIError); ok && githubErr.Code == 403 && strings.Contains(githubErr.Message, "installation access token") {
			return false, true, nil
		}

		return false, false, errors.Wrap(err, "githubClient.ListInstallationRepositories")
	}

	for _, repository := range installationRepositories {
		if repository.NameWithOwner == nameWithOwner {
			return true, false, nil
		}
	}

	return false, false, nil
}

// checkGitHubUserRepositoryPermissions attempts to use the given client as if it's authorized as
// a user. This method returns true when the given name and owner is visible to the user initiating
// the current upload request and that user has write permissions on the repo.
func checkGitHubUserRepositoryPermissions(ctx context.Context, nameWithOwner string, client GitHubClient) (bool, error) {
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return false, errors.New("invalid GitHub repository: nameWithOwner=" + nameWithOwner)
	}

	repository, err := client.GetRepository(ctx, owner, name)
	if err != nil {
		if _, ok := err.(*github.RepoNotFoundError); ok {
			return false, nil
		}

		return false, errors.Wrap(err, "githubClient.GetRepository")
	}

	if repository != nil {
		switch repository.ViewerPermission {
		case "ADMIN", "MAINTAIN", "WRITE":
			// Can edit repository contents
			return true, nil
		}
	}

	return false, nil
}

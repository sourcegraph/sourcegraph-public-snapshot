package auth

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestCheckGitHubPermissions(t *testing.T) {
	type testCase struct {
		description                      string
		expectedAuthor                   bool
		expectedErr                      error
		getRepositoryHook                func(context.Context, string, string) (*github.Repository, error)
		listInstallationRepositoriesHook func(context.Context, int) ([]*github.Repository, bool, int, error)
	}

	testErr := errors.Newf("uh-oh")

	getRepositoryHookAuthorizedRepository := func(ctx context.Context, owner, name string) (*github.Repository, error) {
		return &github.Repository{ViewerPermission: "WRITE"}, nil
	}

	getRepositoryHookEmptyRepository := func(ctx context.Context, owner, name string) (*github.Repository, error) {
		return &github.Repository{}, nil
	}

	getRepositoryHookError := func(ctx context.Context, owner, name string) (*github.Repository, error) {
		return nil, testErr
	}

	getRepositoryHookUnauthorizedRepository := func(ctx context.Context, owner, name string) (*github.Repository, error) {
		return nil, &github.RepoNotFoundError{}
	}

	listInstallationRepositoriesHookMatchingRepository := func(ctx context.Context, page int) ([]*github.Repository, bool, int, error) {
		return []*github.Repository{{NameWithOwner: "sourcegraph/sourcegraph"}}, false, 1, nil
	}

	listInstallationRepositoriesHookNonMatchingRepository := func(ctx context.Context, page int) ([]*github.Repository, bool, int, error) {
		return []*github.Repository{{NameWithOwner: "sourcegraph/not-sourcegraph"}}, false, 1, nil
	}

	listInstallationRepositoriesHookCalledWithUserToken := func(ctx context.Context, page int) ([]*github.Repository, bool, int, error) {
		// This error occurs when a user token is supplied to an app installation endpoint
		return nil, false, 1, &github.APIError{Code: 403, Message: "You must authenticate with an installation access token in order to list repositories for an installation."}
	}

	listInstallationRepositoriesHookError := func(ctx context.Context, page int) ([]*github.Repository, bool, int, error) {
		return nil, false, 1, testErr
	}

	testCases := []testCase{
		{
			description:                      "accessible repo; user token",
			expectedAuthor:                   true,
			expectedErr:                      nil,
			getRepositoryHook:                getRepositoryHookAuthorizedRepository,
			listInstallationRepositoriesHook: listInstallationRepositoriesHookCalledWithUserToken,
		},
		{
			description:                      "accessible repo; app token",
			expectedAuthor:                   true,
			expectedErr:                      nil,
			getRepositoryHook:                getRepositoryHookEmptyRepository,
			listInstallationRepositoriesHook: listInstallationRepositoriesHookMatchingRepository,
		},
		{
			description:                      "inaccessible repo; user token",
			expectedAuthor:                   false,
			expectedErr:                      nil,
			getRepositoryHook:                getRepositoryHookUnauthorizedRepository,
			listInstallationRepositoriesHook: listInstallationRepositoriesHookCalledWithUserToken,
		},
		{
			description:                      "inaccessible repo; app token",
			expectedAuthor:                   false,
			expectedErr:                      nil,
			getRepositoryHook:                getRepositoryHookUnauthorizedRepository,
			listInstallationRepositoriesHook: listInstallationRepositoriesHookNonMatchingRepository,
		},
		{
			description:                      "unexpected GetRepository error",
			expectedAuthor:                   false,
			expectedErr:                      errors.Wrap(testErr, "githubClient.GetRepository"),
			getRepositoryHook:                getRepositoryHookError,
			listInstallationRepositoriesHook: listInstallationRepositoriesHookCalledWithUserToken,
		},
		{
			description:                      "unexpected ListInstallationRepositoriesHook error",
			expectedAuthor:                   false,
			expectedErr:                      errors.Wrap(testErr, "githubClient.ListInstallationRepositories"),
			getRepositoryHook:                getRepositoryHookEmptyRepository,
			listInstallationRepositoriesHook: listInstallationRepositoriesHookError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			client := NewMockGitHubClient()
			client.ListInstallationRepositoriesFunc.SetDefaultHook(testCase.listInstallationRepositoriesHook)
			client.GetRepositoryFunc.SetDefaultHook(testCase.getRepositoryHook)

			author, err := checkGitHubPermissions(context.Background(), "github.com/sourcegraph/sourcegraph", client)
			if author != testCase.expectedAuthor {
				t.Errorf("unexpected status. want=%v have=%v", testCase.expectedAuthor, author)
			}
			if ((err == nil) != (testCase.expectedErr == nil)) || (err != nil && testCase.expectedErr != nil && err.Error() != testCase.expectedErr.Error()) {
				t.Errorf("unexpected error. want=%s have=%s", testCase.expectedErr, err)
			}
		})
	}
}

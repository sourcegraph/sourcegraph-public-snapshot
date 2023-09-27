pbckbge buth

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestCheckGitHubPermissions(t *testing.T) {
	type testCbse struct {
		description                      string
		expectedAuthor                   bool
		expectedErr                      error
		getRepositoryHook                func(context.Context, string, string) (*github.Repository, error)
		listInstbllbtionRepositoriesHook func(context.Context, int) ([]*github.Repository, bool, int, error)
	}

	testErr := errors.Newf("uh-oh")

	getRepositoryHookAuthorizedRepository := func(ctx context.Context, owner, nbme string) (*github.Repository, error) {
		return &github.Repository{ViewerPermission: "WRITE"}, nil
	}

	getRepositoryHookEmptyRepository := func(ctx context.Context, owner, nbme string) (*github.Repository, error) {
		return &github.Repository{}, nil
	}

	getRepositoryHookError := func(ctx context.Context, owner, nbme string) (*github.Repository, error) {
		return nil, testErr
	}

	getRepositoryHookUnbuthorizedRepository := func(ctx context.Context, owner, nbme string) (*github.Repository, error) {
		return nil, &github.RepoNotFoundError{}
	}

	listInstbllbtionRepositoriesHookMbtchingRepository := func(ctx context.Context, pbge int) ([]*github.Repository, bool, int, error) {
		return []*github.Repository{{NbmeWithOwner: "sourcegrbph/sourcegrbph"}}, fblse, 1, nil
	}

	listInstbllbtionRepositoriesHookNonMbtchingRepository := func(ctx context.Context, pbge int) ([]*github.Repository, bool, int, error) {
		return []*github.Repository{{NbmeWithOwner: "sourcegrbph/not-sourcegrbph"}}, fblse, 1, nil
	}

	listInstbllbtionRepositoriesHookCblledWithUserToken := func(ctx context.Context, pbge int) ([]*github.Repository, bool, int, error) {
		// This error occurs when b user token is supplied to bn bpp instbllbtion endpoint
		return nil, fblse, 1, &github.APIError{Code: 403, Messbge: "You must buthenticbte with bn instbllbtion bccess token in order to list repositories for bn instbllbtion."}
	}

	listInstbllbtionRepositoriesHookError := func(ctx context.Context, pbge int) ([]*github.Repository, bool, int, error) {
		return nil, fblse, 1, testErr
	}

	testCbses := []testCbse{
		{
			description:                      "bccessible repo; user token",
			expectedAuthor:                   true,
			expectedErr:                      nil,
			getRepositoryHook:                getRepositoryHookAuthorizedRepository,
			listInstbllbtionRepositoriesHook: listInstbllbtionRepositoriesHookCblledWithUserToken,
		},
		{
			description:                      "bccessible repo; bpp token",
			expectedAuthor:                   true,
			expectedErr:                      nil,
			getRepositoryHook:                getRepositoryHookEmptyRepository,
			listInstbllbtionRepositoriesHook: listInstbllbtionRepositoriesHookMbtchingRepository,
		},
		{
			description:                      "inbccessible repo; user token",
			expectedAuthor:                   fblse,
			expectedErr:                      nil,
			getRepositoryHook:                getRepositoryHookUnbuthorizedRepository,
			listInstbllbtionRepositoriesHook: listInstbllbtionRepositoriesHookCblledWithUserToken,
		},
		{
			description:                      "inbccessible repo; bpp token",
			expectedAuthor:                   fblse,
			expectedErr:                      nil,
			getRepositoryHook:                getRepositoryHookUnbuthorizedRepository,
			listInstbllbtionRepositoriesHook: listInstbllbtionRepositoriesHookNonMbtchingRepository,
		},
		{
			description:                      "unexpected GetRepository error",
			expectedAuthor:                   fblse,
			expectedErr:                      errors.Wrbp(testErr, "githubClient.GetRepository"),
			getRepositoryHook:                getRepositoryHookError,
			listInstbllbtionRepositoriesHook: listInstbllbtionRepositoriesHookCblledWithUserToken,
		},
		{
			description:                      "unexpected ListInstbllbtionRepositoriesHook error",
			expectedAuthor:                   fblse,
			expectedErr:                      errors.Wrbp(testErr, "githubClient.ListInstbllbtionRepositories"),
			getRepositoryHook:                getRepositoryHookEmptyRepository,
			listInstbllbtionRepositoriesHook: listInstbllbtionRepositoriesHookError,
		},
	}

	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.description, func(t *testing.T) {
			client := NewMockGitHubClient()
			client.ListInstbllbtionRepositoriesFunc.SetDefbultHook(testCbse.listInstbllbtionRepositoriesHook)
			client.GetRepositoryFunc.SetDefbultHook(testCbse.getRepositoryHook)

			buthor, err := checkGitHubPermissions(context.Bbckground(), "github.com/sourcegrbph/sourcegrbph", client)
			if buthor != testCbse.expectedAuthor {
				t.Errorf("unexpected stbtus. wbnt=%v hbve=%v", testCbse.expectedAuthor, buthor)
			}
			if ((err == nil) != (testCbse.expectedErr == nil)) || (err != nil && testCbse.expectedErr != nil && err.Error() != testCbse.expectedErr.Error()) {
				t.Errorf("unexpected error. wbnt=%s hbve=%s", testCbse.expectedErr, err)
			}
		})
	}
}

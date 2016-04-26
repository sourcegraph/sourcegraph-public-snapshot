package github

import (
	"testing"

	"github.com/sourcegraph/go-github/github"
)

func TestAuthorizations_Revoke(t *testing.T) {
	var called bool
	ctx := testContext(&minimalClient{
		appAuthorizations: mockGitHubAuthorizations{
			Revoke_: func(clientID, token string) (*github.Response, error) {
				called = true
				return nil, nil
			},
		},
	})

	if err := (&Authorizations{}).Revoke(ctx, "c", "t"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("!called")
	}
}

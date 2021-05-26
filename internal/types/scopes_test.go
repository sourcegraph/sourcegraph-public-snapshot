package types

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func TestGrantedScopes(t *testing.T) {
	want := []string{"repo"}
	github.MockGetAuthenticatedUserOAuthScopes = func(ctx context.Context) ([]string, error) {
		return want, nil
	}

	ctx := context.Background()
	have, err := GrantedScopes(ctx, extsvc.KindGitHub, `{}`)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

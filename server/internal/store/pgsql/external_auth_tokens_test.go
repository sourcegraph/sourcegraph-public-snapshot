package pgsql

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/auth"
)

func (s *ExternalAuthTokens) mustSetUserToken(ctx context.Context, t *testing.T, tok *auth.ExternalAuthToken) {
	if err := s.SetUserToken(ctx, tok); err != nil {
		t.Fatal(err)
	}
}

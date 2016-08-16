package localstore

import (
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func (s *externalAuthTokens) mustSetUserToken(ctx context.Context, t *testing.T, tok *store.ExternalAuthToken) {
	if err := s.SetUserToken(ctx, tok); err != nil {
		t.Fatal(err)
	}
}

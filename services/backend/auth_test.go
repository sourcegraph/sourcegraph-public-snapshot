package backend

import (
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func TestAuthService_DeleteAndRevokeExternalToken(t *testing.T) {
	var s auth
	ctx, mock := testContext()

	var calledGetUserToken, calledDeleteToken bool
	mock.stores.ExternalAuthTokens.GetUserToken_ = func(ctx context.Context, user int, host, clientID string) (*store.ExternalAuthToken, error) {
		calledGetUserToken = true
		return &store.ExternalAuthToken{Host: "h", ClientID: "c", Token: "t"}, nil
	}
	mock.stores.ExternalAuthTokens.DeleteToken_ = func(ctx context.Context, tok *sourcegraph.ExternalTokenSpec) error {
		calledDeleteToken = true
		return nil
	}

	if _, err := s.DeleteAndRevokeExternalToken(ctx, &sourcegraph.ExternalTokenSpec{Host: "h", ClientID: "c"}); err != nil {
		t.Fatal(err)
	}
	if !calledGetUserToken {
		t.Error("!calledGetUserToken")
	}
	if !calledDeleteToken {
		t.Error("!calledDeleteToken")
	}
}

package authzchecked

import (
	"os"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/store/mockstore"
)

func TestExternalAuthTokens_GetUserToken_self_ok(t *testing.T) {
	var m mockstore.ExternalAuthTokens
	calledGetUserToken := m.MockGetUserToken(t)

	ctx := auth.WithActor(context.Background(), auth.Actor{UID: 1})

	if _, err := ExternalAuthTokens(&m).GetUserToken(ctx, 1, "", ""); err != nil {
		t.Error(nil)
	}
	if !*calledGetUserToken {
		t.Error("!calledGetUserToken")
	}
}

func TestExternalAuthTokens_GetUserToken_otherUser_forbidden(t *testing.T) {
	var m mockstore.ExternalAuthTokens
	calledGetUserToken := m.MockGetUserToken(t)

	ctx := auth.WithActor(context.Background(), auth.Actor{UID: 1})

	if _, err := ExternalAuthTokens(&m).GetUserToken(ctx, 2, "", ""); err != os.ErrPermission {
		t.Errorf("got err == %v, want %v", err, os.ErrPermission)
	}
	if *calledGetUserToken {
		t.Error("calledGetUserToken")
	}
}

func TestExternalAuthTokens_SetUserToken_self_ok(t *testing.T) {
	var m mockstore.ExternalAuthTokens
	calledSetUserToken := m.MockSetUserToken(t)

	ctx := auth.WithActor(context.Background(), auth.Actor{UID: 1})

	tok := &auth.ExternalAuthToken{User: 1}
	if err := ExternalAuthTokens(&m).SetUserToken(ctx, tok); err != nil {
		t.Fatal(err)
	}
	if !*calledSetUserToken {
		t.Error("!calledSetUserToken")
	}
}

func TestExternalAuthTokens_SetUserToken_otherUser_forbidden(t *testing.T) {
	var m mockstore.ExternalAuthTokens
	calledSetUserToken := m.MockSetUserToken(t)

	ctx := auth.WithActor(context.Background(), auth.Actor{UID: 1})

	tok := &auth.ExternalAuthToken{User: 2}
	if err := ExternalAuthTokens(&m).SetUserToken(ctx, tok); err != os.ErrPermission {
		t.Errorf("got err == %v, want %v", err, os.ErrPermission)
	}
	if *calledSetUserToken {
		t.Error("calledSetUserToken")
	}
}

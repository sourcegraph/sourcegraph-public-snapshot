// +build pgsqltest

package pgsql

import (
	"reflect"
	"testing"
	"time"

	"src.sourcegraph.com/sourcegraph/auth"
)

func TestExternalAuthTokens_GetUserToken_found(t *testing.T) {
	t.Parallel()

	var s ExternalAuthTokens
	ctx, done := testContext()
	defer done()

	wantToken := &auth.ExternalAuthToken{
		User:     1,
		Host:     "example.com",
		ClientID: "c",
		Token:    "t",
	}
	s.mustSetUserToken(ctx, t, wantToken)

	tok, err := s.GetUserToken(ctx, 1, "example.com", "c")
	if err != nil {
		t.Fatal(err)
	}
	normalizeExternalAuthToken(wantToken)
	if !reflect.DeepEqual(tok, wantToken) {
		t.Errorf("got token %+v, want %+v", tok, wantToken)
	}
}

func TestExternalAuthTokens_GetUserToken_notFound(t *testing.T) {
	t.Parallel()

	var s ExternalAuthTokens
	ctx, done := testContext()
	defer done()

	tok, err := s.GetUserToken(ctx, 1, "example.com", "c")
	if wantErr := auth.ErrNoExternalAuthToken; err != wantErr {
		t.Errorf("got err = %q, want %q", err, wantErr)
	}
	if tok != nil {
		t.Errorf("got tok == %v, want nil", tok)
	}
}

func TestExternalAuthTokens_SetUserToken_create(t *testing.T) {
	t.Parallel()

	var s ExternalAuthTokens
	ctx, done := testContext()
	defer done()

	tok := &auth.ExternalAuthToken{
		User:     1,
		Host:     "example.com",
		ClientID: "c",
		Token:    "t1",
	}
	if err := s.SetUserToken(ctx, tok); err != nil {
		t.Fatal(err)
	}

	tok1, err := s.GetUserToken(ctx, 1, "example.com", "c")
	if err != nil {
		t.Fatal(err)
	}
	normalizeExternalAuthToken(tok1)
	normalizeExternalAuthToken(tok)
	if !reflect.DeepEqual(tok1, tok) {
		t.Errorf("got token %+v, want %+v", tok1, tok)
	}
}

func TestExternalAuthTokens_SetUserToken_update(t *testing.T) {
	t.Parallel()

	var s ExternalAuthTokens
	ctx, done := testContext()
	defer done()

	tok0 := &auth.ExternalAuthToken{
		User:     1,
		Host:     "example.com",
		ClientID: "c",
		Token:    "t0",
	}
	s.mustSetUserToken(ctx, t, tok0)

	tok1 := &auth.ExternalAuthToken{
		User:     1,
		Host:     "example.com",
		ClientID: "c",
		Token:    "t1",
	}
	if err := s.SetUserToken(ctx, tok1); err != nil {
		t.Fatal(err)
	}

	// check that the token was updated to t1
	tok, err := s.GetUserToken(ctx, 1, "example.com", "c")
	if err != nil {
		t.Fatal(err)
	}
	normalizeExternalAuthToken(tok)
	normalizeExternalAuthToken(tok1)
	if !reflect.DeepEqual(tok, tok1) {
		t.Errorf("got token %+v, want %+v", tok, tok1)
	}
}

func normalizeExternalAuthToken(tok *auth.ExternalAuthToken) {
	tok.RefreshedAt = tok.RefreshedAt.In(time.UTC).Round(time.Second)
	if tok.FirstAuthFailureAt != nil {
		rounded := tok.FirstAuthFailureAt.In(time.UTC).Round(time.Second)
		tok.FirstAuthFailureAt = &rounded
	}
}

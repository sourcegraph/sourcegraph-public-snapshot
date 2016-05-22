// +build pgsqltest

package localstore

import (
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func TestExternalAuthTokens_GetUserToken_found(t *testing.T) {
	t.Parallel()

	var s externalAuthTokens
	ctx, _, done := testContext()
	defer done()

	wantToken := &store.ExternalAuthToken{
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

	var s externalAuthTokens
	ctx, _, done := testContext()
	defer done()

	tok, err := s.GetUserToken(ctx, 1, "example.com", "c")
	if wantErr := store.ErrNoExternalAuthToken; err != wantErr {
		t.Errorf("got err = %q, want %q", err, wantErr)
	}
	if tok != nil {
		t.Errorf("got tok == %v, want nil", tok)
	}
}

func TestExternalAuthTokens_SetUserToken_create(t *testing.T) {
	t.Parallel()

	var s externalAuthTokens
	ctx, _, done := testContext()
	defer done()

	tok := &store.ExternalAuthToken{
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

	var s externalAuthTokens
	ctx, _, done := testContext()
	defer done()

	tok0 := &store.ExternalAuthToken{
		User:     1,
		Host:     "example.com",
		ClientID: "c",
		Token:    "t0",
	}
	s.mustSetUserToken(ctx, t, tok0)

	tok1 := &store.ExternalAuthToken{
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

func TestExternalAuthTokens_DeleteToken(t *testing.T) {
	t.Parallel()

	var s externalAuthTokens
	ctx, _, done := testContext()
	defer done()

	tok0 := &store.ExternalAuthToken{
		User:     1,
		Host:     "example.com",
		ClientID: "c",
		Token:    "t0",
	}
	s.mustSetUserToken(ctx, t, tok0)

	tokSpec := &sourcegraph.ExternalTokenSpec{
		UID:      1,
		Host:     "example.com",
		ClientID: "c",
	}

	if err := s.DeleteToken(ctx, tokSpec); err != nil {
		t.Fatal(err)
	}

	_, err := s.GetUserToken(ctx, 1, "example.com", "c")
	if wantErr := store.ErrNoExternalAuthToken; err != wantErr {
		t.Errorf("got err = %q, want %q", err, wantErr)
	}
}

func TestExternalAuthTokens_ListExternalUsers_empty(t *testing.T) {
	t.Parallel()

	var s externalAuthTokens
	ctx, _, done := testContext()
	defer done()

	dbToks := []*store.ExternalAuthToken{
		{
			User:     1,
			Host:     "example.com",
			ClientID: "c",
			Token:    "t0",
			ExtUID:   12345,
		},
		{
			User:     2,
			Host:     "example.com",
			ClientID: "c",
			Token:    "t1",
			ExtUID:   12346,
		},
		{
			User:     3,
			Host:     "example2.com",
			ClientID: "c",
			Token:    "t3",
			ExtUID:   12347,
		},
	}
	s.mustSetUserToken(ctx, t, dbToks[0])
	s.mustSetUserToken(ctx, t, dbToks[1])
	s.mustSetUserToken(ctx, t, dbToks[2])

	// fetch the tokens for external users
	toks, err := s.ListExternalUsers(ctx, []int{}, "example.com", "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(toks) != 0 {
		t.Fatalf("got %d tokens, want 0", len(toks))
	}
}

func TestExternalAuthTokens_ListExternalUsers_nonempty(t *testing.T) {
	t.Parallel()

	var s externalAuthTokens
	ctx, _, done := testContext()
	defer done()

	dbToks := []*store.ExternalAuthToken{
		{
			User:     1,
			Host:     "example.com",
			ClientID: "c",
			Token:    "t0",
			ExtUID:   12345,
		},
		{
			User:     2,
			Host:     "example.com",
			ClientID: "c",
			Token:    "t1",
			ExtUID:   12346,
		},
		{
			User:     3,
			Host:     "example2.com",
			ClientID: "c",
			Token:    "t3",
			ExtUID:   12347,
		},
	}
	s.mustSetUserToken(ctx, t, dbToks[0])
	s.mustSetUserToken(ctx, t, dbToks[1])
	s.mustSetUserToken(ctx, t, dbToks[2])

	// fetch the tokens for external users
	toks, err := s.ListExternalUsers(ctx, []int{12345, 12347}, "example.com", "c")
	if err != nil {
		t.Fatal(err)
	}
	for i := range dbToks {
		normalizeExternalAuthToken(dbToks[i])
	}
	for i := range toks {
		normalizeExternalAuthToken(toks[i])
	}
	if len(toks) != 1 {
		t.Fatalf("got %d tokens, want 1", len(toks))
	}
	if !reflect.DeepEqual(dbToks[0], toks[0]) {
		t.Errorf("got token %+v, want %+v", toks[0], dbToks[0])
	}
}

func normalizeExternalAuthToken(tok *store.ExternalAuthToken) {
	tok.RefreshedAt = tok.RefreshedAt.In(time.UTC).Round(time.Second)
	if tok.FirstAuthFailureAt != nil {
		rounded := tok.FirstAuthFailureAt.In(time.UTC).Round(time.Second)
		tok.FirstAuthFailureAt = &rounded
	}
}

// +build pgsqltest

package localstore

import (
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

// TestAuthorizations_CreateAuthCode_MarkExchanged_ok tests the behavior
// of CreateAuthCode by ensuring a code it creates can be exchanged
// with MarkExchanged.
func TestAuthorizations_CreateAuthCode_MarkExchanged_ok(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &authorizations{}
	req := &sourcegraph.AuthorizationCodeRequest{
		ClientID:    "c",
		RedirectURI: "u",
		Scope:       []string{"a", "b"},
		UID:         123,
	}

	code, err := s.CreateAuthCode(ctx, req, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) < 5 {
		t.Fatalf("got code == %q, want len(code) >= 5", code)
	}

	xreq, err := s.MarkExchanged(ctx, &sourcegraph.AuthorizationCode{Code: code, RedirectURI: "u"}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(xreq, req) {
		t.Errorf("got exchanged req %+v, want %+v", xreq, req)
	}
}

// TestAuthorizations_MarkExchanged_doesntexist tests the behavior of
// MarkExchanged when the code does not exist (and no codes exist).
func TestAuthorizations_MarkExchanged_doesntexist(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &authorizations{}
	xreq, err := s.MarkExchanged(ctx, &sourcegraph.AuthorizationCode{Code: "mycode", RedirectURI: "u"}, "c")
	if want := store.ErrAuthCodeNotFound; err != want {
		t.Fatalf("got error %v, want %v", err, want)
	}
	if xreq != nil {
		t.Error("xreq != nil")
	}
}

// TesAuthorizations_MarkExchanged_doesntexist tests the behavior of
// MarkExchanged when the code does not exist (but some codes have been added).
func TestAuthorizations_MarkExchanged_codeNotFound(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &authorizations{}
	code, err := s.CreateAuthCode(ctx, &sourcegraph.AuthorizationCodeRequest{
		ClientID:    "c",
		RedirectURI: "u",
		Scope:       []string{"a", "b"},
		UID:         123,
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	xreq, err := s.MarkExchanged(ctx, &sourcegraph.AuthorizationCode{Code: "bad" + code + "bad", RedirectURI: "u"}, "c")
	if want := store.ErrAuthCodeNotFound; err != want {
		t.Fatalf("got error %v, want %v", err, want)
	}
	if xreq != nil {
		t.Error("xreq != nil")
	}
}

// TestAuthorizations_MarkExchanged_clientIDMismatch tests the behavior of
// MarkExchanged when the client IDs do not match.
func TestAuthorizations_MarkExchanged_clientIDMismatch(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &authorizations{}
	code, err := s.CreateAuthCode(ctx, &sourcegraph.AuthorizationCodeRequest{
		ClientID:    "c",
		RedirectURI: "u",
		Scope:       []string{"a", "b"},
		UID:         123,
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	xreq, err := s.MarkExchanged(ctx, &sourcegraph.AuthorizationCode{Code: code, RedirectURI: "u"}, "badClientID")
	if want := store.ErrAuthCodeNotFound; err != want {
		t.Fatalf("got error %v, want %v", err, want)
	}
	if xreq != nil {
		t.Error("xreq != nil")
	}
}

// TestAuthorizations_MarkExchanged_redirectURIMismatch tests the behavior
// of MarkExchanged when the redirect URIs do not match.
func TestAuthorizations_MarkExchanged_redirectURIMismatch(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &authorizations{}
	code, err := s.CreateAuthCode(ctx, &sourcegraph.AuthorizationCodeRequest{
		ClientID:    "c",
		RedirectURI: "u",
		Scope:       []string{"a", "b"},
		UID:         123,
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	xreq, err := s.MarkExchanged(ctx, &sourcegraph.AuthorizationCode{Code: code, RedirectURI: "badRedirectURI"}, "c")
	if want := store.ErrAuthCodeNotFound; err != want {
		t.Fatalf("got error %v, want %v", err, want)
	}
	if xreq != nil {
		t.Error("xreq != nil")
	}
}

// TestAuthorizations_MarkExchanged_expired tests the behavior of
// MarkExchanged when the code has expired.
func TestAuthorizations_MarkExchanged_expired(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &authorizations{}
	code, err := s.CreateAuthCode(ctx, &sourcegraph.AuthorizationCodeRequest{
		ClientID:    "c",
		RedirectURI: "u",
		Scope:       []string{"a", "b"},
		UID:         123,
	}, 5*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)

	xreq, err := s.MarkExchanged(ctx, &sourcegraph.AuthorizationCode{Code: code, RedirectURI: "u"}, "c")
	if want := store.ErrAuthCodeNotFound; err != want {
		t.Fatalf("got error %v, want %v", err, want)
	}
	if xreq != nil {
		t.Error("xreq != nil")
	}
}

// TestAuthorizations_MarkExchanged_alreadyExchanged tests the behavior of
// MarkExchanged when the code has already been exchanged.
func TestAuthorizations_MarkExchanged_alreadyExchanged(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &authorizations{}
	code, err := s.CreateAuthCode(ctx, &sourcegraph.AuthorizationCodeRequest{
		ClientID:    "c",
		RedirectURI: "u",
		Scope:       []string{"a", "b"},
		UID:         123,
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.MarkExchanged(ctx, &sourcegraph.AuthorizationCode{Code: code, RedirectURI: "u"}, "c"); err != nil {
		t.Fatal(err)
	}

	xreq, err := s.MarkExchanged(ctx, &sourcegraph.AuthorizationCode{Code: code, RedirectURI: "u"}, "c")
	if want := store.ErrAuthCodeAlreadyExchanged; err != want {
		t.Fatalf("got error %v, want %v", err, want)
	}
	if xreq != nil {
		t.Error("xreq != nil")
	}
}

package testsuite

import (
	"testing"
	"time"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// Authorizations_MarkExchanged_redirectURIMismatch tests the behavior
// of MarkExchanged when the redirect URIs do not match.
func Authorizations_MarkExchanged_redirectURIMismatch(ctx context.Context, t *testing.T, s store.Authorizations) {
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

// Authorizations_MarkExchanged_expired tests the behavior of
// MarkExchanged when the code has expired.
func Authorizations_MarkExchanged_expired(ctx context.Context, t *testing.T, s store.Authorizations) {
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

// Authorizations_MarkExchanged_alreadyExchanged tests the behavior of
// MarkExchanged when the code has already been exchanged.
func Authorizations_MarkExchanged_alreadyExchanged(ctx context.Context, t *testing.T, s store.Authorizations) {
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

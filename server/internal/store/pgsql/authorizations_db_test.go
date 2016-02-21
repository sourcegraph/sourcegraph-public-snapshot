// +build pgsqltest

package pgsql

import (
	"reflect"
	"testing"
	"time"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
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

func TestAuthorizations_MarkExchanged_doesntexist(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_doesntexist(ctx, t, &authorizations{})
}

func TestAuthorizations_MarkExchanged_codeNotFound(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_codeNotFound(ctx, t, &authorizations{})
}

func TestAuthorizations_MarkExchanged_clientIDMismatch(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_clientIDMismatch(ctx, t, &authorizations{})
}

func TestAuthorizations_MarkExchanged_redirectURIMismatch(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_redirectURIMismatch(ctx, t, &authorizations{})
}

func TestAuthorizations_MarkExchanged_expired(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_expired(ctx, t, &authorizations{})
}

func TestAuthorizations_MarkExchanged_alreadyExchanged(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_alreadyExchanged(ctx, t, &authorizations{})
}

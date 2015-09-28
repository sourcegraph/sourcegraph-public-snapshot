package fs

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/store/testsuite"
)

func TestAuthorizations_CreateAuthCode_MarkExchanged_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_CreateAuthCode_MarkExchanged_ok(ctx, t, &Authorizations{})
}

func TestAuthorizations_MarkExchanged_doesntexist(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_doesntexist(ctx, t, &Authorizations{})
}

func TestAuthorizations_MarkExchanged_codeNotFound(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_codeNotFound(ctx, t, &Authorizations{})
}

func TestAuthorizations_MarkExchanged_clientIDMismatch(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_clientIDMismatch(ctx, t, &Authorizations{})
}

func TestAuthorizations_MarkExchanged_redirectURIMismatch(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_redirectURIMismatch(ctx, t, &Authorizations{})
}

func TestAuthorizations_MarkExchanged_expired(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_expired(ctx, t, &Authorizations{})
}

func TestAuthorizations_MarkExchanged_alreadyExchanged(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_MarkExchanged_alreadyExchanged(ctx, t, &Authorizations{})
}

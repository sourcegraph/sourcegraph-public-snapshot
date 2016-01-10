// +build pgsqltest

package pgsql

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func TestAuthorizations_CreateAuthCode_MarkExchanged_ok(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Authorizations_CreateAuthCode_MarkExchanged_ok(ctx, t, &authorizations{})
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

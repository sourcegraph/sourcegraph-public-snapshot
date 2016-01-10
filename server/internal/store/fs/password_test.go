package fs

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func TestPasswords_CheckUIDPassword_valid(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Passwords_CheckUIDPassword_valid(ctx, t, &password{})
}

func TestPasswords_CheckUIDPassword_invalid(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Passwords_CheckUIDPassword_invalid(ctx, t, &password{})
}

func TestPasswords_CheckUIDPassword_empty(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Passwords_CheckUIDPassword_empty(ctx, t, &password{})
}

func TestPasswords_CheckUIDPassword_noneSet(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Passwords_CheckUIDPassword_noneSet(ctx, t, &password{})
}

func TestPasswords_CheckUIDPassword_noneSetForUser(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Passwords_CheckUIDPassword_noneSetForUser(ctx, t, &password{})
}

func TestPasswords_SetPassword_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Passwords_SetPassword_ok(ctx, t, &password{})
}

func TestPasswords_SetPassword_empty(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Passwords_SetPassword_empty(ctx, t, &password{})
}

func TestPasswords_SetPassword_setToEmpty(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Passwords_SetPassword_setToEmpty(ctx, t, &password{})
}

// +build pgsqltest

package pgsql

import (
	"sync/atomic"
	"testing"

	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

var testUID int32

// nextUID returns a unique test user UID for this process. This is needed
// since we do sets and compares on passwords for users, and if tests are
// running in parallel the results returned will be racey.
func nextUID() int32 {
	return atomic.AddInt32(&testUID, 1)
}

// TestPasswords_CheckUIDPassword_valid tests the behavior of
// Passwords.CheckUIDPassword when called with valid credentials.
func TestPasswords_CheckUIDPassword_valid(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &password{}
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}
}

// TestPasswords_CheckUIDPassword_invalid tests the behavior of
// Passwords.CheckUIDPassword when called with invalid credentials.
func TestPasswords_CheckUIDPassword_invalid(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &password{}
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, uid, "WRONG"); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_CheckUIDPassword_empty tests the behavior of
// Passwords.CheckUIDPassword when called with empty credentials.
func TestPasswords_CheckUIDPassword_empty(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &password{}
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, uid, ""); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_CheckUIDPassword_noneSet tests the behavior of
// Passwords.CheckUIDPassword when there is no password set.
func TestPasswords_CheckUIDPassword_noneSet(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &password{}
	uid := nextUID()
	if err := s.CheckUIDPassword(ctx, uid, "p"); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_CheckUIDPassword_noneSetForUser tests the behavior of
// Passwords.CheckUIDPassword when there is no password set for the
// given user (but other users have passwords).
func TestPasswords_CheckUIDPassword_noneSetForUser(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &password{}
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	uid = nextUID()
	if err := s.CheckUIDPassword(ctx, uid, "p"); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_SetPassword_ok tests changing the password.
func TestPasswords_SetPassword_ok(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &password{}
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	// Password is p.
	if err := s.CheckUIDPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckUIDPassword(ctx, uid, "p2"); err == nil {
		t.Fatal("err == nil")
	}

	// Change to p2.
	if err := s.SetPassword(ctx, uid, "p2"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckUIDPassword(ctx, uid, "p2"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckUIDPassword(ctx, uid, "p"); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_SetPassword_empty tests changing the password to an
// empty password.
func TestPasswords_SetPassword_empty(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &password{}
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, ""); err == nil {
		t.Fatal("err == nil")
	}
}

func TestPasswords_SetPassword_setToEmpty(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Passwords_SetPassword_setToEmpty(ctx, t, &password{})
}

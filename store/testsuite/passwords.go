package testsuite

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

// Passwords_CheckUIDPassword_valid tests the behavior of
// Passwords.CheckUIDPassword when called with valid credentials.
func Passwords_CheckUIDPassword_valid(ctx context.Context, t *testing.T, s store.Password) {
	if err := s.SetPassword(ctx, 1, "p"); err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, 1, "p"); err != nil {
		t.Fatal(err)
	}
}

// Passwords_CheckUIDPassword_invalid tests the behavior of
// Passwords.CheckUIDPassword when called with invalid credentials.
func Passwords_CheckUIDPassword_invalid(ctx context.Context, t *testing.T, s store.Password) {
	if err := s.SetPassword(ctx, 1, "p"); err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, 1, "WRONG"); err == nil {
		t.Fatal("err == nil")
	}
}

// Passwords_CheckUIDPassword_empty tests the behavior of
// Passwords.CheckUIDPassword when called with empty credentials.
func Passwords_CheckUIDPassword_empty(ctx context.Context, t *testing.T, s store.Password) {
	if err := s.SetPassword(ctx, 1, "p"); err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, 1, ""); err == nil {
		t.Fatal("err == nil")
	}
}

// Passwords_CheckUIDPassword_noneSet tests the behavior of
// Passwords.CheckUIDPassword when there is no password set.
func Passwords_CheckUIDPassword_noneSet(ctx context.Context, t *testing.T, s store.Password) {
	if err := s.CheckUIDPassword(ctx, 1, "p"); err == nil {
		t.Fatal("err == nil")
	}
}

// Passwords_CheckUIDPassword_noneSetForUser tests the behavior of
// Passwords.CheckUIDPassword when there is no password set for the
// given user (but other users have passwords).
func Passwords_CheckUIDPassword_noneSetForUser(ctx context.Context, t *testing.T, s store.Password) {
	if err := s.SetPassword(ctx, 1, "p"); err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, 2, "p"); err == nil {
		t.Fatal("err == nil")
	}
}

// Passwords_SetPassword_ok tests changing the password.
func Passwords_SetPassword_ok(ctx context.Context, t *testing.T, s store.Password) {
	if err := s.SetPassword(ctx, 1, "p"); err != nil {
		t.Fatal(err)
	}

	// Password is p.
	if err := s.CheckUIDPassword(ctx, 1, "p"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckUIDPassword(ctx, 1, "p2"); err == nil {
		t.Fatal("err == nil")
	}

	// Change to p2.
	if err := s.SetPassword(ctx, 1, "p2"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckUIDPassword(ctx, 1, "p2"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckUIDPassword(ctx, 1, "p"); err == nil {
		t.Fatal("err == nil")
	}
}

// Passwords_SetPassword_empty tests changing the password to an
// empty password.
func Passwords_SetPassword_empty(ctx context.Context, t *testing.T, s store.Password) {
	if err := s.SetPassword(ctx, 1, ""); err == nil {
		t.Fatal("err == nil")
	}
}

// Passwords_SetPassword_setToEmpty tests changing the password FROM a
// valid password to an empty password.
func Passwords_SetPassword_setToEmpty(ctx context.Context, t *testing.T, s store.Password) {
	if err := s.SetPassword(ctx, 1, "p"); err != nil {
		t.Fatal(err)
	}

	// Set to empty
	if err := s.SetPassword(ctx, 1, ""); err == nil {
		t.Fatal("err == nil")
	}

	// Password should remain as "p".
	if err := s.CheckUIDPassword(ctx, 1, "p"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckUIDPassword(ctx, 1, "p2"); err == nil {
		t.Fatal("err == nil")
	}
}

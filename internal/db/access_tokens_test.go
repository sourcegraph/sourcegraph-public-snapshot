package db

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

// 🚨 SECURITY: This tests the routine that creates access tokens and returns the token secret value
// to the user.
func TestAccessTokens_Create(t *testing.T) {
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	subject, err := GlobalUsers.Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	creator, err := GlobalUsers.Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	tid0, tv0, err := GlobalAccessTokens.Create(ctx, subject.ID, []string{"a", "b"}, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GlobalAccessTokens.GetByID(ctx, tid0)
	if err != nil {
		t.Fatal(err)
	}
	if want := tid0; got.ID != want {
		t.Errorf("got %v, want %v", got.ID, want)
	}
	if want := subject.ID; got.SubjectUserID != want {
		t.Errorf("got %v, want %v", got.SubjectUserID, want)
	}
	if want := "n0"; got.Note != want {
		t.Errorf("got %q, want %q", got.Note, want)
	}

	gotSubjectUserID, err := GlobalAccessTokens.Lookup(ctx, tv0, "a")
	if err != nil {
		t.Fatal(err)
	}
	if want := subject.ID; gotSubjectUserID != want {
		t.Errorf("got %v, want %v", gotSubjectUserID, want)
	}

	ts, err := GlobalAccessTokens.List(ctx, AccessTokensListOptions{SubjectUserID: subject.ID})
	if err != nil {
		t.Fatal(err)
	}
	if want := 1; len(ts) != want {
		t.Errorf("got %d access tokens, want %d", len(ts), want)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(ts[0].Scopes, want) {
		t.Errorf("got token scopes %q, want %q", ts[0].Scopes, want)
	}

	// Accidentally passing the creator's UID in SubjectUserID should not return anything.
	ts, err = GlobalAccessTokens.List(ctx, AccessTokensListOptions{SubjectUserID: creator.ID})
	if err != nil {
		t.Fatal(err)
	}
	if want := 0; len(ts) != want {
		t.Errorf("got %d access tokens, want %d", len(ts), want)
	}
}

func TestAccessTokens_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	subject1, err := GlobalUsers.Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}
	subject2, err := GlobalUsers.Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = GlobalAccessTokens.Create(ctx, subject1.ID, []string{"a", "b"}, "n0", subject1.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = GlobalAccessTokens.Create(ctx, subject1.ID, []string{"a", "b"}, "n1", subject1.ID)
	if err != nil {
		t.Fatal(err)
	}

	{
		// List all tokens.
		ts, err := GlobalAccessTokens.List(ctx, AccessTokensListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d access tokens, want %d", len(ts), want)
		}
		count, err := GlobalAccessTokens.Count(ctx, AccessTokensListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List subject1's tokens.
		ts, err := GlobalAccessTokens.List(ctx, AccessTokensListOptions{SubjectUserID: subject1.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d access tokens, want %d", len(ts), want)
		}
	}

	{
		// List subject2's tokens.
		ts, err := GlobalAccessTokens.List(ctx, AccessTokensListOptions{SubjectUserID: subject2.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := 0; len(ts) != want {
			t.Errorf("got %d access tokens, want %d", len(ts), want)
		}
	}
}

// 🚨 SECURITY: This tests the routine that verifies access tokens, which the security of the entire
// system depends on.
func TestAccessTokens_Lookup(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	subject, err := GlobalUsers.Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	creator, err := GlobalUsers.Create(ctx, NewUser{
		Email:                 "u2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	tid0, tv0, err := GlobalAccessTokens.Create(ctx, subject.ID, []string{"a", "b"}, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}

	for _, scope := range []string{"a", "b"} {
		gotSubjectUserID, err := GlobalAccessTokens.Lookup(ctx, tv0, scope)
		if err != nil {
			t.Fatal(err)
		}
		if want := subject.ID; gotSubjectUserID != want {
			t.Errorf("got %v, want %v", gotSubjectUserID, want)
		}
	}

	// Lookup with a nonexistent scope and ensure it fails.
	if _, err := GlobalAccessTokens.Lookup(ctx, tv0, "x"); err == nil {
		t.Fatal(err)
	}

	// Lookup with an empty scope and ensure it fails.
	if _, err := GlobalAccessTokens.Lookup(ctx, tv0, ""); err == nil {
		t.Fatal(err)
	}

	// Delete a token and ensure Lookup fails on it.
	if err := GlobalAccessTokens.DeleteByID(ctx, tid0, subject.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := GlobalAccessTokens.Lookup(ctx, tv0, "a"); err == nil {
		t.Fatal(err)
	}

	// Try to Lookup a token that was never created.
	if _, err := GlobalAccessTokens.Lookup(ctx, "abcdefg" /* this token value was never created */, "a"); err == nil {
		t.Fatal(err)
	}
}

// 🚨 SECURITY: This tests that deleting the subject or creator user of an access token invalidates
// the token, and that no new access tokens may be created for deleted users.
func TestAccessTokens_Lookup_deletedUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	t.Run("subject", func(t *testing.T) {
		subject, err := GlobalUsers.Create(ctx, NewUser{
			Email:                 "u1@example.com",
			Username:              "u1",
			Password:              "p1",
			EmailVerificationCode: "c1",
		})
		if err != nil {
			t.Fatal(err)
		}
		creator, err := GlobalUsers.Create(ctx, NewUser{
			Email:                 "u2@example.com",
			Username:              "u2",
			Password:              "p2",
			EmailVerificationCode: "c2",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, tv0, err := GlobalAccessTokens.Create(ctx, subject.ID, []string{"a"}, "n0", creator.ID)
		if err != nil {
			t.Fatal(err)
		}
		if err := GlobalUsers.Delete(ctx, subject.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := GlobalAccessTokens.Lookup(ctx, tv0, "a"); err == nil {
			t.Fatal("Lookup: want error looking up token for deleted subject user")
		}

		if _, _, err := GlobalAccessTokens.Create(ctx, subject.ID, nil, "n0", creator.ID); err == nil {
			t.Fatal("Create: want error creating token for deleted subject user")
		}
	})

	t.Run("creator", func(t *testing.T) {
		subject, err := GlobalUsers.Create(ctx, NewUser{
			Email:                 "u3@example.com",
			Username:              "u3",
			Password:              "p3",
			EmailVerificationCode: "c3",
		})
		if err != nil {
			t.Fatal(err)
		}
		creator, err := GlobalUsers.Create(ctx, NewUser{
			Email:                 "u4@example.com",
			Username:              "u4",
			Password:              "p4",
			EmailVerificationCode: "c4",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, tv0, err := GlobalAccessTokens.Create(ctx, subject.ID, []string{"a"}, "n0", creator.ID)
		if err != nil {
			t.Fatal(err)
		}
		if err := GlobalUsers.Delete(ctx, creator.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := GlobalAccessTokens.Lookup(ctx, tv0, "a"); err == nil {
			t.Fatal("Lookup: want error looking up token for deleted creator user")
		}

		if _, _, err := GlobalAccessTokens.Create(ctx, subject.ID, nil, "n0", creator.ID); err == nil {
			t.Fatal("Create: want error creating token for deleted creator user")
		}
	})
}

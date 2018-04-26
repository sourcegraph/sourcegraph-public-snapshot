package db

import "testing"

// 🚨 SECURITY: This tests the routine that creates access tokens and returns the token secret value
// to the user.
func TestAccessTokens_Create(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	subject, err := Users.Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	creator, err := Users.Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	tid0, tv0, err := AccessTokens.Create(ctx, subject.ID, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}

	got, err := AccessTokens.GetByID(ctx, tid0)
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

	gotSubjectUserID, err := AccessTokens.Lookup(ctx, tv0)
	if err != nil {
		t.Fatal(err)
	}
	if want := subject.ID; gotSubjectUserID != want {
		t.Errorf("got %v, want %v", gotSubjectUserID, want)
	}

	ts, err := AccessTokens.List(ctx, AccessTokensListOptions{SubjectUserID: subject.ID})
	if err != nil {
		t.Fatal(err)
	}
	if want := 1; len(ts) != want {
		t.Errorf("got %d access tokens, want %d", len(ts), want)
	}

	// Accidentally passing the creator's UID in SubjectUserID should not return anything.
	ts, err = AccessTokens.List(ctx, AccessTokensListOptions{SubjectUserID: creator.ID})
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
	ctx := testContext()

	subject1, err := Users.Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}
	subject2, err := Users.Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = AccessTokens.Create(ctx, subject1.ID, "n0", subject1.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = AccessTokens.Create(ctx, subject1.ID, "n1", subject1.ID)
	if err != nil {
		t.Fatal(err)
	}

	{
		// List all tokens.
		ts, err := AccessTokens.List(ctx, AccessTokensListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d access tokens, want %d", len(ts), want)
		}
		count, err := AccessTokens.Count(ctx, AccessTokensListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List subject1's tokens.
		ts, err := AccessTokens.List(ctx, AccessTokensListOptions{SubjectUserID: subject1.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d access tokens, want %d", len(ts), want)
		}
	}

	{
		// List subject2's tokens.
		ts, err := AccessTokens.List(ctx, AccessTokensListOptions{SubjectUserID: subject2.ID})
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
	ctx := testContext()

	subject, err := Users.Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	creator, err := Users.Create(ctx, NewUser{
		Email:                 "u2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	tid0, tv0, err := AccessTokens.Create(ctx, subject.ID, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}

	gotSubjectUserID, err := AccessTokens.Lookup(ctx, tv0)
	if err != nil {
		t.Fatal(err)
	}
	if want := subject.ID; gotSubjectUserID != want {
		t.Errorf("got %v, want %v", gotSubjectUserID, want)
	}

	// Delete a token and ensure Lookup fails on it.
	if err := AccessTokens.DeleteByID(ctx, tid0, subject.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := AccessTokens.Lookup(ctx, tv0); err == nil {
		t.Fatal(err)
	}

	// Try to Lookup a token that was never created.
	if _, err := AccessTokens.Lookup(ctx, "abcdefg" /* this token value was never created */); err == nil {
		t.Fatal(err)
	}
}

package db

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestUserEmails_ListByUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	user, err := Users.Create(ctx, NewUser{
		Email:     "a@example.com",
		Username:  "u2",
		Password:  "pw",
		EmailCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	testTime := time.Now().Round(time.Second).UTC()
	if _, err := globalDB.ExecContext(ctx,
		`INSERT INTO user_emails(user_id, email, verification_code, verified_at) VALUES($1, $2, $3, $4)`,
		user.ID, "b@example.com", "c2", testTime); err != nil {
		t.Fatal(err)
	}

	userEmails, err := UserEmails.ListByUser(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	normalizeUserEmails(userEmails)
	if want := []*UserEmail{
		{UserID: user.ID, Email: "a@example.com", VerificationCode: strptr("c")},
		{UserID: user.ID, Email: "b@example.com", VerificationCode: strptr("c2"), VerifiedAt: &testTime},
	}; !reflect.DeepEqual(userEmails, want) {
		t.Errorf("got  %s\n\nwant %s", toJSON(userEmails), toJSON(want))
	}
}

func normalizeUserEmails(userEmails []*UserEmail) {
	for _, v := range userEmails {
		v.CreatedAt = time.Time{}
		if v.VerifiedAt != nil {
			tmp := v.VerifiedAt.Round(time.Second).UTC()
			v.VerifiedAt = &tmp
		}
	}
}

func TestUserEmails_Add(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	const emailA = "a@example.com"
	const emailB = "b@example.com"
	user, err := Users.Create(ctx, NewUser{
		Email:     emailA,
		Username:  "u2",
		Password:  "pw",
		EmailCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := UserEmails.Add(ctx, user.ID, emailB, nil); err != nil {
		t.Fatal(err)
	}
	if verified, err := isUserEmailVerified(ctx, user.ID, emailB); err != nil {
		t.Fatal(err)
	} else if want := false; verified != want {
		t.Fatalf("got verified %v, want %v", verified, want)
	}

	if err := UserEmails.Add(ctx, user.ID, emailB, nil); err == nil {
		t.Fatal("got err == nil for Add on existing email")
	}
	if err := UserEmails.Add(ctx, 12345 /* bad user ID */, "foo@example.com", nil); err == nil {
		t.Fatal("got err == nil for Add on bad user ID")
	}
}

func TestUserEmails_SetVerified(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	const email = "a@example.com"
	user, err := Users.Create(ctx, NewUser{
		Email:     email,
		Username:  "u2",
		Password:  "pw",
		EmailCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	if verified, err := isUserEmailVerified(ctx, user.ID, email); err != nil {
		t.Fatal(err)
	} else if want := false; verified != want {
		t.Fatalf("before SetVerified, got verified %v, want %v", verified, want)
	}

	if err := UserEmails.SetVerified(ctx, user.ID, email, true); err != nil {
		t.Fatal(err)
	}
	if verified, err := isUserEmailVerified(ctx, user.ID, email); err != nil {
		t.Fatal(err)
	} else if want := true; verified != want {
		t.Fatalf("after SetVerified true, got verified %v, want %v", verified, want)
	}

	if err := UserEmails.SetVerified(ctx, user.ID, email, false); err != nil {
		t.Fatal(err)
	}
	if verified, err := isUserEmailVerified(ctx, user.ID, email); err != nil {
		t.Fatal(err)
	} else if want := false; verified != want {
		t.Fatalf("after SetVerified false, got verified %v, want %v", verified, want)
	}

	if err := UserEmails.SetVerified(ctx, user.ID, "otheremail@example.com", false); err == nil {
		t.Fatal("got err == nil for SetVerified on bad email")
	}
}

func isUserEmailVerified(ctx context.Context, userID int32, email string) (bool, error) {
	userEmails, err := UserEmails.ListByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, v := range userEmails {
		if v.Email == email {
			return v.VerifiedAt != nil, nil
		}
	}
	return false, fmt.Errorf("email not found: %s", email)
}

func strptr(s string) *string {
	return &s
}

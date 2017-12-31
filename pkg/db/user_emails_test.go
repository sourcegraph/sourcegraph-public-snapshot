package db

import (
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
	for _, v := range userEmails {
		v.CreatedAt = time.Time{}
		if v.VerifiedAt != nil {
			tmp := v.VerifiedAt.Round(time.Second).UTC()
			v.VerifiedAt = &tmp
		}
	}
	if want := []*UserEmail{
		{UserID: user.ID, Email: "a@example.com", VerificationCode: strptr("c")},
		{UserID: user.ID, Email: "b@example.com", VerificationCode: strptr("c2"), VerifiedAt: &testTime},
	}; !reflect.DeepEqual(userEmails, want) {
		t.Errorf("got  %s\n\nwant %s", toJSON(userEmails), toJSON(want))
	}
}

func strptr(s string) *string {
	return &s
}

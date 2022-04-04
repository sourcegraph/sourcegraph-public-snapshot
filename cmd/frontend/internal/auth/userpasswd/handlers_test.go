package userpasswd

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestCheckEmailAbuse(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	farFuture := now.AddDate(100, 0, 0)

	tests := []struct {
		name      string
		mockEmail *database.UserEmail
		mockErr   error
		expAbused bool
		expReason string
		expErr    error
	}{
		{
			name:      "no emails found",
			mockEmail: nil,
			mockErr:   database.MockUserEmailNotFoundErr,
			expAbused: false,
			expReason: "",
			expErr:    nil,
		},
		{
			name: "needs cool down",
			mockEmail: &database.UserEmail{
				LastVerificationSentAt: &farFuture,
			},
			mockErr:   nil,
			expAbused: true,
			expReason: "too frequent attempt since last verification email sent",
			expErr:    nil,
		},

		{
			name: "no abuse",
			mockEmail: &database.UserEmail{
				LastVerificationSentAt: &yesterday,
			},
			mockErr:   nil,
			expAbused: false,
			expReason: "",
			expErr:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userEmails := database.NewMockUserEmailsStore()
			userEmails.GetLatestVerificationSentEmailFunc.SetDefaultReturn(test.mockEmail, test.mockErr)
			db := database.NewMockDB()
			db.UserEmailsFunc.SetDefaultReturn(userEmails)

			abused, reason, err := checkEmailAbuse(context.Background(), db, "fake@localhost")
			if test.expErr != err {
				t.Fatalf("err: want %v but got %v", test.expErr, err)
			} else if test.expAbused != abused {
				t.Fatalf("abused: want %v but got %v", test.expAbused, abused)
			} else if test.expReason != reason {
				t.Fatalf("reason: want %q but got %q", test.expReason, reason)
			}
		})
	}
}

func TestCheckEmailFormat(t *testing.T) {
	for name, test := range map[string]struct {
		email string
		err   error
		code  int
	}{
		"valid":   {email: "foo@bar.pl", err: nil},
		"invalid": {email: "foo@", err: errors.Newf("mail: no angle-addr")},
		"toolong": {email: "a012345678901234567890123456789012345678901234567890123456789@0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.comeeeeqwqwwe", err: errors.Newf("maximum email length is 320, got 326")}} {
		t.Run(name, func(t *testing.T) {
			err := checkEmailFormat(test.email)
			if test.err == nil {
				if err != nil {
					t.Fatalf("err: want nil but got %v", err)
				}
			} else {
				if test.err.Error() != err.Error() {
					t.Fatalf("err: want %v but got %v", test.err, err)
				}
			}
		})
	}
}

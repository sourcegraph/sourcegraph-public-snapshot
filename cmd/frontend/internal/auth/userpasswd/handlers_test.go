package userpasswd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
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

func TestHandleSignIn_Lockout(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthProviders: []schema.AuthProviders{
				{
					Builtin: &schema.BuiltinAuthProvider{
						Type: providerType,
					},
				},
			},
		},
	})
	defer conf.Mock(nil)

	users := database.NewMockUserStore()
	users.GetByUsernameFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.EventLogsFunc.SetDefaultReturn(database.NewMockEventLogStore())
	db.SecurityEventLogsFunc.SetDefaultReturn(database.NewMockSecurityEventLogsStore())
	db.UserEmailsFunc.SetDefaultReturn(database.NewMockUserEmailsStore())

	lockout := NewMockLockoutStore()
	logger := logtest.NoOp(t)
	if testing.Verbose() {
		logger = logtest.Scoped(t)
	}
	h := HandleSignIn(logger, db, lockout)

	// Normal authentication fail before lockout
	{
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "Authentication failed\n", resp.Body.String())
	}

	// Getting error for locked out
	{
		lockout.IsLockedOutFunc.SetDefaultReturn("reason", true)
		lockout.SendUnlockAccountEmailFunc.SetDefaultReturn(nil)
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code)
		assert.Equal(t, `Account has been locked out due to "reason"`+"\n", resp.Body.String())
	}
}

func TestHandleAccount_Unlock(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthProviders: []schema.AuthProviders{
				{
					Builtin: &schema.BuiltinAuthProvider{
						Type: providerType,
					},
				},
			},
		},
	})
	defer conf.Mock(nil)

	db := database.NewMockDB()
	db.EventLogsFunc.SetDefaultReturn(database.NewMockEventLogStore())
	db.SecurityEventLogsFunc.SetDefaultReturn(database.NewMockSecurityEventLogsStore())

	lockout := NewMockLockoutStore()
	logger := logtest.NoOp(t)
	if testing.Verbose() {
		logger = logtest.Scoped(t)
	}
	h := HandleUnlockAccount(logger, db, lockout)

	// bad request if missing token or user id
	{
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Bad request: missing token\n", resp.Body.String())
	}

	// Getting error for invalid token
	{
		lockout.VerifyUnlockAccountTokenAndResetFunc.SetDefaultReturn(false, errors.Newf("invalid token provided"))
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{ "token": "abcd" }`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "invalid token provided\n", resp.Body.String())
	}

	// ok result
	{
		lockout.VerifyUnlockAccountTokenAndResetFunc.SetDefaultReturn(true, nil)
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{ "token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJpc3MiOiJodHRwczovL3NvdXJjZWdyYXBoLnRlc3Q6MzQ0MyIsInN1YiI6IjEiLCJleHAiOjE2NDk3NzgxNjl9.cm_giwkSviVRXGRCie9iii-ytJD3iAuNdtk9XmBZMrj7HHlH6vfky4ftjudAZ94HBp867cjxkuNc6OJ2uaEJFg" }`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "", resp.Body.String())
	}
}

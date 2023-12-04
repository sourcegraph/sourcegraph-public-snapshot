package database

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAccessTokens(t *testing.T) {
	// perform test setup and teardown
	prevConfg := conf.Get()
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		Log: &schema.Log{
			SecurityEventLog: &schema.SecurityEventLog{Location: "database"},
		},
	}})
	t.Cleanup(func() {
		conf.Mock(prevConfg)
	})

	t.Run("TestAccessTokens_parallel", func(t *testing.T) {
		t.Run("testAccessTokens_Create", testAccessTokens_Create)
		t.Run("testAccessTokens_Delete", testAccessTokens_Delete)
		t.Run("testAccessTokens_Create", testAccessTokens_CreateInternal_DoesNotCaptureSecurityEvent)
		t.Run("testAccessTokens_List", testAccessTokens_List)
		t.Run("testAccessTokens_Lookup", testAccessTokens_Lookup)
		t.Run("testAccessToken_Lookup_deletedUser", testAccessTokens_Lookup_deletedUser)
		t.Run("testAccessTokens_tokenSHA256Hash", testAccessTokens_tokenSHA256Hash)
	})

	// Don't run parallel as it's mocking an expired license
	t.Run("testAccessToken_Lookup_expiredLicense", testAccessTokens_Lookup_expiredLicense)
}

// 🚨 SECURITY: This tests the routine that creates access tokens and returns the token secret value
// to the user.
//
// testAccessTokens_Create requires the site_config to be mocked to enable security event logging to the database.
// This test is run in TestAccessTokens
func testAccessTokens_Create(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	subject, err := db.Users().Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	creator, err := db.Users().Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	assertSecurityEventCount(t, db, SecurityEventAccessTokenCreated, 0)
	tid0, tv0, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a", "b"}, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertSecurityEventCount(t, db, SecurityEventAccessTokenCreated, 1)

	if !strings.HasPrefix(tv0, "sgp_") {
		t.Errorf("got %q, want prefix 'sgp_'", tv0)
	}

	got, err := db.AccessTokens().GetByID(ctx, tid0)
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

	gotSubjectUserID, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if want := subject.ID; gotSubjectUserID != want {
		t.Errorf("got %v, want %v", gotSubjectUserID, want)
	}

	ts, err := db.AccessTokens().List(ctx, AccessTokensListOptions{SubjectUserID: subject.ID})
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
	ts, err = db.AccessTokens().List(ctx, AccessTokensListOptions{SubjectUserID: creator.ID})
	if err != nil {
		t.Fatal(err)
	}
	if want := 0; len(ts) != want {
		t.Errorf("got %d access tokens, want %d", len(ts), want)
	}
}

// testAccessTokens_Delete requires the site_config to be mocked to enable security event logging to the database
// This test is run in TestAccessTokens
func testAccessTokens_Delete(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	subject, err := db.Users().Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	creator, err := db.Users().Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create context with valid actor; required by logging
	subjectActor := actor.FromUser(subject.ID)
	ctxWithActor := actor.WithActor(context.Background(), subjectActor)

	tid0, _, err := db.AccessTokens().Create(ctxWithActor, subject.ID, []string{"a", "b"}, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, tv1, err := db.AccessTokens().Create(ctxWithActor, subject.ID, []string{"a", "b"}, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}
	tid2, _, err := db.AccessTokens().Create(ctxWithActor, subject.ID, []string{"a", "b"}, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}

	assertSecurityEventCount(t, db, SecurityEventAccessTokenDeleted, 0)
	err = db.AccessTokens().DeleteByID(ctxWithActor, tid0)
	if err != nil {
		t.Fatal(err)
	}
	assertSecurityEventCount(t, db, SecurityEventAccessTokenDeleted, 1)
	err = db.AccessTokens().DeleteByToken(ctxWithActor, tv1)
	if err != nil {
		t.Fatal(err)
	}
	assertSecurityEventCount(t, db, SecurityEventAccessTokenDeleted, 2)

	assertSecurityEventCount(t, db, SecurityEventAccessTokenHardDeleted, 0)
	err = db.AccessTokens().HardDeleteByID(ctxWithActor, tid2)
	if err != nil {
		t.Fatal(err)
	}
	assertSecurityEventCount(t, db, SecurityEventAccessTokenHardDeleted, 1)
}

func assertSecurityEventCount(t *testing.T, db DB, event SecurityEventName, expectedCount int) {
	t.Helper()

	row := db.SecurityEventLogs().Handle().QueryRowContext(context.Background(), "SELECT count(name) FROM security_event_logs WHERE name = $1", event)
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatal("couldn't read security events count")
	}
	assert.Equal(t, expectedCount, count)
}

// This test is run in TestAccessTokens
func testAccessTokens_CreateInternal_DoesNotCaptureSecurityEvent(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	subject, err := db.Users().Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	creator, err := db.Users().Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	assertSecurityEventCount(t, db, SecurityEventAccessTokenCreated, 0)
	_, _, err = db.AccessTokens().CreateInternal(ctx, subject.ID, []string{"a", "b"}, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertSecurityEventCount(t, db, SecurityEventAccessTokenCreated, 0)
}

// This test is run in TestAccessTokens
func testAccessTokens_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Parallel()
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	subject1, err := db.Users().Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}
	subject2, err := db.Users().Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = db.AccessTokens().Create(ctx, subject1.ID, []string{"a", "b"}, "n0", subject1.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = db.AccessTokens().Create(ctx, subject1.ID, []string{"a", "b"}, "n1", subject1.ID)
	if err != nil {
		t.Fatal(err)
	}

	{
		// List all tokens.
		ts, err := db.AccessTokens().List(ctx, AccessTokensListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d access tokens, want %d", len(ts), want)
		}
		count, err := db.AccessTokens().Count(ctx, AccessTokensListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List subject1's tokens.
		ts, err := db.AccessTokens().List(ctx, AccessTokensListOptions{SubjectUserID: subject1.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d access tokens, want %d", len(ts), want)
		}
	}

	{
		// List subject2's tokens.
		ts, err := db.AccessTokens().List(ctx, AccessTokensListOptions{SubjectUserID: subject2.ID})
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
// This test is run in TestAccessTokens
func testAccessTokens_Lookup(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Parallel()
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	subject, err := db.Users().Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	creator, err := db.Users().Create(ctx, NewUser{
		Email:                 "u2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	tid0, tv0, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a", "b"}, "n0", creator.ID)
	if err != nil {
		t.Fatal(err)
	}

	for _, scope := range []string{"a", "b"} {
		gotSubjectUserID, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: scope})
		if err != nil {
			t.Fatal(err)
		}
		if want := subject.ID; gotSubjectUserID != want {
			t.Errorf("got %v, want %v", gotSubjectUserID, want)
		}
	}

	// Lookup with a nonexistent scope and ensure it fails.
	if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "x"}); err == nil {
		t.Fatal(err)
	}

	// Lookup with an empty scope and ensure it fails.
	if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: ""}); err == nil {
		t.Fatal(err)
	}

	// Delete a token and ensure Lookup fails on it.
	if err := db.AccessTokens().DeleteByID(ctx, tid0); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "a"}); err == nil {
		t.Fatal(err)
	}

	// Try to Lookup a token that was never created.
	if _, err := db.AccessTokens().Lookup(ctx, "abcdefg" /* this token value was never created */, TokenLookupOpts{RequiredScope: "a"}); err == nil {
		t.Fatal(err)
	}
}

// 🚨 SECURITY: This tests that deleting the subject or creator user of an access token invalidates
// the token, and that no new access tokens may be created for deleted users.
// This test is run in TestAccessTokens
func testAccessTokens_Lookup_deletedUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Parallel()
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	t.Run("subject", func(t *testing.T) {
		subject, err := db.Users().Create(ctx, NewUser{
			Email:                 "u1@example.com",
			Username:              "u1",
			Password:              "p1",
			EmailVerificationCode: "c1",
		})
		if err != nil {
			t.Fatal(err)
		}
		creator, err := db.Users().Create(ctx, NewUser{
			Email:                 "u2@example.com",
			Username:              "u2",
			Password:              "p2",
			EmailVerificationCode: "c2",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, tv0, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a"}, "n0", creator.ID)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Users().Delete(ctx, subject.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "a"}); err == nil {
			t.Fatal("Lookup: want error looking up token for deleted subject user")
		}

		if _, _, err := db.AccessTokens().Create(ctx, subject.ID, nil, "n0", creator.ID); err == nil {
			t.Fatal("Create: want error creating token for deleted subject user")
		}
	})

	t.Run("creator", func(t *testing.T) {
		subject, err := db.Users().Create(ctx, NewUser{
			Email:                 "u3@example.com",
			Username:              "u3",
			Password:              "p3",
			EmailVerificationCode: "c3",
		})
		if err != nil {
			t.Fatal(err)
		}
		creator, err := db.Users().Create(ctx, NewUser{
			Email:                 "u4@example.com",
			Username:              "u4",
			Password:              "p4",
			EmailVerificationCode: "c4",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, tv0, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a"}, "n0", creator.ID)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Users().Delete(ctx, creator.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "a"}); err == nil {
			t.Fatal("Lookup: want error looking up token for deleted creator user")
		}

		if _, _, err := db.AccessTokens().Create(ctx, subject.ID, nil, "n0", creator.ID); err == nil {
			t.Fatal("Create: want error creating token for deleted creator user")
		}
	})
}

// 🚨 SECURITY: This tests that deleting the subject or creator user of an access token invalidates
// the token, and that no new access tokens may be created for deleted users.
// This test is run in TestAccessTokens
func testAccessTokens_Lookup_expiredLicense(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	adminUser, err := db.Users().Create(ctx, NewUser{
		Email:                 "u1@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, adminUser.ID, true))

	regularUser, err := db.Users().Create(ctx, NewUser{
		Email:                 "u2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, adminToken, err := db.AccessTokens().Create(ctx, adminUser.ID, []string{"a"}, "n0", adminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, regularToken, err := db.AccessTokens().Create(ctx, regularUser.ID, []string{"a"}, "n0", regularUser.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.AccessTokens().Lookup(ctx, adminToken, TokenLookupOpts{RequiredScope: "a", OnlyAdmin: true}); err != nil {
		t.Fatal("Lookup: lookup should not fail for admin user")
	}
	if _, err := db.AccessTokens().Lookup(ctx, regularToken, TokenLookupOpts{RequiredScope: "a", OnlyAdmin: true}); err == nil {
		t.Fatal("Lookup: lookup should fail for regular user")
	}
}

// This test is run in TestAccessTokens
func testAccessTokens_tokenSHA256Hash(t *testing.T) {
	testCases := []struct {
		name      string
		token     string
		wantError bool
	}{
		{name: "old prefix-less format", token: "0123456789012345678901234567890123456789"},
		{name: "old prefix format", token: "sgp_0123456789012345678901234567890123456789"},
		{name: "new local identifier format", token: "sgp_local_0123456789012345678901234567890123456789"},
		{name: "new identifier format", token: "sgp_abcdef0123456789_0123456789012345678901234567890123456789"},
		{name: "empty", token: "", wantError: true},
		{name: "invalid", token: "×", wantError: true},
		{name: "invalid", token: "xxx", wantError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := tokenSHA256Hash(tc.token)
			if tc.wantError {
				assert.ErrorContains(t, err, "invalid token")
			} else {
				assert.NoError(t, err)
				if len(hash) != 32 {
					t.Errorf("got %d characters, want 32", len(hash))
				}
			}
		})
	}
}

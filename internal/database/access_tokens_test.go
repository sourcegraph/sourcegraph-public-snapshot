package database

// pre-commit:ignore_sourcegraph_token

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
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
		t.Run("testAccessTokens_GetOrCreateInternalToken", testAccessTokens_GetOrCreateInternalToken)
		t.Run("testAccessTokens_Expiration", testAccessTokens_Expiration)
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
	tid0, tv0, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a", "b"}, "n0", creator.ID, time.Time{})
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

	tid0, _, err := db.AccessTokens().Create(ctxWithActor, subject.ID, []string{"a", "b"}, "n0", creator.ID, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	_, tv1, err := db.AccessTokens().Create(ctxWithActor, subject.ID, []string{"a", "b"}, "n0", creator.ID, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	tid2, _, err := db.AccessTokens().Create(ctxWithActor, subject.ID, []string{"a", "b"}, "n0", creator.ID, time.Time{})
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

	_, _, err = db.AccessTokens().Create(ctx, subject1.ID, []string{"a", "b"}, "n0", subject1.ID, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = db.AccessTokens().Create(ctx, subject1.ID, []string{"a", "b"}, "n1", subject1.ID, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = db.AccessTokens().Create(ctx, subject1.ID, []string{"a", "b"}, "expired", subject1.ID, time.Now().Add(-1*time.Hour))
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

	// Create the DB instance as well as a handle to the underlying implementation,
	// so we can modify the table directly. We alias db because the local variable
	// db will shadow the type without any way to disambiguate.
	type dbType = db
	db := NewDB(logger, dbtest.NewDB(t))
	rawDB, ok := db.(*dbType)
	if !ok {
		t.Fatal("NewDB returns a DB handle that is using unexpected implementation.")
	}

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

	tid0, tv0, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a", "b"}, "n0", creator.ID, time.Time{})
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

	// Calls to .Lookup() will automatically refresh the last_used_at column, but no more than a fixed
	// frequency.
	t.Run("last_used_at Updates", func(t *testing.T) {
		// Create a new access token.
		testTokenID, testTokenValue, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a", "b", "c"}, "n0", creator.ID, time.Time{})
		if err != nil {
			t.Fatal(err)
		}

		// Fetches the test access token. On any error aborts the test.
		mustGetTestToken := func() *AccessToken {
			token, err := db.AccessTokens().GetByID(ctx, testTokenID)
			if err != nil {
				t.Fatal(err)
			}
			return token
		}

		assertLastUsedSinceShorterThan := func(lastUsedAt *time.Time, maxTimeSince time.Duration) {
			t.Helper()
			if lastUsedAt == nil {
				t.Fatal("time passed was nil.")
			}
			if time.Since(*lastUsedAt) > maxTimeSince {
				t.Fatalf("last_used_at value is more recent than %v", maxTimeSince)
			}
		}

		// Check the current value. last_used_at is initialized to be nil.
		initialState := mustGetTestToken()
		if initialState.LastUsedAt != nil {
			t.Fatal("last_used_at was not nil upon token creation")
		}

		// Confirm that a side-effect of Lookup will initialize last_used_at.
		// When we fetch the token again, it's value should be recent.
		_, err = db.AccessTokens().Lookup(ctx, testTokenValue, TokenLookupOpts{RequiredScope: "a"})
		if err != nil {
			t.Fatal(err)
		}
		postLookup := mustGetTestToken()
		assertLastUsedSinceShorterThan(postLookup.LastUsedAt, 2*time.Second)

		// Update the token's last_used_at to be old enough to force an update
		// on the next call to .Lookup()
		now := time.Now()
		updateQuery := sqlf.Sprintf(
			`UPDATE access_tokens SET last_used_at = %s WHERE id = %d`,
			now.Add(-MaxAccessTokenLastUsedAtAge-2*time.Second), testTokenID)
		err = rawDB.Store.Exec(ctx, updateQuery)
		if err != nil {
			t.Fatalf("Updating test token's last_used_at: %v", err)
		}

		// Confirm the token was updated, and last_used_at is old.
		staleState := mustGetTestToken()
		if staleState.LastUsedAt == nil {
			t.Fatal("token did not have last_used_at")
		}
		if time.Since(*staleState.LastUsedAt) < MaxAccessTokenLastUsedAtAge {
			t.Fatalf("last_used_at value should be older than %v", *staleState.LastUsedAt)
		}

		// Now lookup the token. A side-effect of this will update the last_used_at.
		_, err = db.AccessTokens().Lookup(ctx, testTokenValue, TokenLookupOpts{RequiredScope: "a"})
		if err != nil {
			t.Fatal(err)
		}

		currentState := mustGetTestToken()
		assertLastUsedSinceShorterThan(currentState.LastUsedAt, 2*time.Second)

		err = db.AccessTokens().DeleteByID(ctx, testTokenID)
		if err != nil {
			t.Fatal(err)
		}
	})
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

		_, tv0, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a"}, "n0", creator.ID, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Users().Delete(ctx, subject.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "a"}); err == nil {
			t.Fatal("Lookup: want error looking up token for deleted subject user")
		}

		if _, _, err := db.AccessTokens().Create(ctx, subject.ID, nil, "n0", creator.ID, time.Time{}); err == nil {
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

		_, tv0, err := db.AccessTokens().Create(ctx, subject.ID, []string{"a"}, "n0", creator.ID, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Users().Delete(ctx, creator.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "a"}); err == nil {
			t.Fatal("Lookup: want error looking up token for deleted creator user")
		}

		if _, _, err := db.AccessTokens().Create(ctx, subject.ID, nil, "n0", creator.ID, time.Time{}); err == nil {
			t.Fatal("Create: want error creating token for deleted creator user")
		}
	})
}

// 🚨 SECURITY: This tests that tokens past the expiration time are invalid
// This test is run in TestAccessTokens
func testAccessTokens_Expiration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Parallel()
	// Create the DB instance as well as a handle to the underlying implementation,
	// so we can modify the table directly. We alias db because the local variable
	// db will shadow the type without any way to disambiguate.
	type dbType = db
	db := NewDB(logger, dbtest.NewDB(t))
	rawDB, ok := db.(*dbType)
	if !ok {
		t.Fatal("NewDB returns a DB handle that is using unexpected implementation.")
	}
	ctx := context.Background()

	user, err := db.Users().Create(ctx, NewUser{
		Email:                 "u1@example.com",
		Username:              "u1",
		Password:              "p1",
		EmailVerificationCode: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create an access token that expires in the future
	testTokenID, tv0, err := db.AccessTokens().Create(ctx, user.ID, []string{"a"}, "n0", user.ID, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// Ensure we can lookup the token
	if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "a"}); err != nil {
		t.Fatal("Lookup: no error expected")
	}

	// Update the expiration to a time in the past
	updateQuery := sqlf.Sprintf(
		`UPDATE access_tokens SET expires_at = %s WHERE id = %d`,
		time.Now().Add(-1*time.Hour), testTokenID)
	err = rawDB.Store.Exec(ctx, updateQuery)
	if err != nil {
		t.Fatalf("Updating test token's expiration to the past: %v", err)
	}

	// Ensure we can no longer lookup the token
	if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "a"}); err == nil {
		t.Fatal("Lookup: want error looking up expired token")
	}

}

// 🚨 SECURITY: TestAccessTokens_Limits tests the enforcement of access token limits per user.
// It creates tokens for a test user and ensures the token limit is enforced
func TestAccessTokens_Limits(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthAccessTokens: &schema.AuthAccessTokens{
				MaxTokensPerUser:  pointers.Ptr(2),
				AllowNoExpiration: pointers.Ptr(true),
			},
			Log: &schema.Log{
				SecurityEventLog: &schema.SecurityEventLog{Location: "database"},
			},
		},
	})
	defer conf.Mock(nil)
	user, err := db.Users().Create(ctx, NewUser{
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

	// Create 2 expired tokens to ensure that expired tokens are not counted against the limit
	_, _, err = db.AccessTokens().Create(ctx, user.ID, []string{"a"}, "n0", user.ID, time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = db.AccessTokens().Create(ctx, user.ID, []string{"1"}, "n0", user.ID, time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// Create an access token that expires in the future
	_, tv0, err := db.AccessTokens().Create(ctx, user.ID, []string{"a"}, "n0", user.ID, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// Create an access token with no expiration
	_, tv1, err := db.AccessTokens().Create(ctx, user.ID, []string{"a"}, "n0", user.ID, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// Ensure we can lookup the tokens
	if _, err := db.AccessTokens().Lookup(ctx, tv0, TokenLookupOpts{RequiredScope: "a"}); err != nil {
		t.Fatal("Lookup: no error expected")
	}
	if _, err := db.AccessTokens().Lookup(ctx, tv1, TokenLookupOpts{RequiredScope: "a"}); err != nil {
		t.Fatal("Lookup: no error expected")
	}

	// Ensure subject user can not create a 3nd token
	_, _, err = db.AccessTokens().Create(ctx, user.ID, []string{"a"}, "n0", user.ID, time.Now().Add(1*time.Hour))
	if err != ErrTooManyAccessTokens {
		t.Fatal("Create: expected ErrTooManyAccessTokens")
	}

	// Ensure another user can not create a 3nd token for the subject
	_, _, err = db.AccessTokens().Create(ctx, user.ID, []string{"a"}, "n0", creator.ID, time.Now().Add(1*time.Hour))
	if err != ErrTooManyAccessTokens {
		t.Fatal("Create: expected ErrTooManyAccessTokens")
	}

	// Ensure that a new internal token can be created
	_, tvInternal, err := db.AccessTokens().CreateInternal(ctx, user.ID, []string{"a"}, "n0", user.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Ensure we can lookup the internal token
	if _, err := db.AccessTokens().Lookup(ctx, tvInternal, TokenLookupOpts{RequiredScope: "a"}); err != nil {
		t.Fatal("Lookup: no error expected")
	}

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

	_, adminToken, err := db.AccessTokens().Create(ctx, adminUser.ID, []string{"a"}, "n0", adminUser.ID, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	_, regularToken, err := db.AccessTokens().Create(ctx, regularUser.ID, []string{"a"}, "n0", regularUser.ID, time.Time{})
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

func testAccessTokens_GetOrCreateInternalToken(t *testing.T) {
	db := NewDB(logtest.Scoped(t), dbtest.NewDB(t))
	ctx := context.Background()

	t.Run("can create and retrieve tokens", func(t *testing.T) {
		u, err := db.Users().Create(ctx, NewUser{Email: "a@test.com", Username: "a", EmailIsVerified: true})
		assert.NoError(t, err)

		// Create token automatically
		sha256, err := db.AccessTokens().GetOrCreateInternalToken(ctx, u.ID, []string{"a"})
		assert.NoError(t, err)
		assert.NotEmpty(t, sha256)

		// Retrieve same token again
		sha256Again, err := db.AccessTokens().GetOrCreateInternalToken(ctx, u.ID, []string{"a"})
		assert.NoError(t, err)
		assert.Equal(t, sha256, sha256Again)
	})

	t.Run("can create and retrieve tokens", func(t *testing.T) {
		u, err := db.Users().Create(ctx, NewUser{Email: "b@test.com", Username: "b", EmailIsVerified: true})
		assert.NoError(t, err)

		// Create token manually
		_, token, err := db.AccessTokens().CreateInternal(ctx, u.ID, []string{"a"}, "n0", u.ID)
		assert.NoError(t, err)
		hashedToken, err := tokenSHA256Hash(token)
		assert.NoError(t, err)

		// Get token and compare it
		sha256, err := db.AccessTokens().GetOrCreateInternalToken(ctx, u.ID, []string{"a"})
		assert.NoError(t, err)
		assert.Equal(t, hashedToken, sha256)
	})
}

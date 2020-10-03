package campaigns

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/secret"
)

func testStoreUserTokens(t *testing.T, ctx context.Context, s *Store, rs repos.Store, clock clock) {
	// All of our tests require an external service. We'll actually create two,
	// so we can perform meaningful list tests.
	svcA := createExternalService(t, ctx, rs, extsvc.KindGitHub)
	svcB := createExternalService(t, ctx, rs, extsvc.KindGitLab)

	// Similarly, we need users.
	uidA := insertTestUser(t, ctx, s.DB(), "a")
	uidB := insertTestUser(t, ctx, s.DB(), "b")

	// We also need to be able to refer to the tokens in multiple places, so
	// let's define our token placeholders here, along with other state we need
	// to track across subtests.
	var (
		tokenAA *campaigns.UserToken
		tokenAB *campaigns.UserToken

		createdAt = time.Date(2020, 10, 2, 23, 56, 0, 0, time.UTC)
		updatedAt = time.Date(2020, 10, 2, 23, 57, 0, 0, time.UTC)
	)

	// UpsertUserToken is going to serve a double purpose: it's both going to
	// test the eponymous function, and also set up the state we're going to
	// use in the other subtests, at least as far as user A's tokens go.
	t.Run("UpsertUserToken insert", func(t *testing.T) {
		// Create a user token with implicit created and updated times.
		tokenAA = &campaigns.UserToken{
			UserID:            uidA,
			ExternalServiceID: svcA.ID,
			Token:             createSecretString("foo"),
		}
		if err := s.UpsertUserToken(ctx, tokenAA); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		// Create a user token with explicit created and updated times.
		tokenAB = &campaigns.UserToken{
			UserID:            uidA,
			ExternalServiceID: svcB.ID,
			Token:             createSecretString("bar"),
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
		}
		if err := s.UpsertUserToken(ctx, tokenAB); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
	})

	t.Run("GetUserToken", func(t *testing.T) {
		// Let's get those tokens that we just created back and see what we
		// have.
		if have, err := s.GetUserToken(ctx, uidA, svcA.ID); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		} else if diff := cmp.Diff(have, tokenAA); diff != "" {
			t.Errorf("unexpected token:\n%s", diff)
		}

		if have, err := s.GetUserToken(ctx, uidA, svcB.ID); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		} else if diff := cmp.Diff(have, tokenAB); diff != "" {
			t.Errorf("unexpected token:\n%s", diff)
		}

		// Now let's see what happens when we ask for a token that doesn't
		// exist.
		if _, err := s.GetUserToken(ctx, uidB, svcA.ID); err != ErrNoResults {
			t.Errorf("unexpected error: have=%v want=%v", err, ErrNoResults)
		}
	})

	t.Run("ListUserTokens", func(t *testing.T) {
		// If we don't set any limits, then we should just get all the tokens
		// for the user back.
		if have, next, err := s.ListUserTokens(ctx, ListUserTokensOpts{
			UserID: uidA,
		}); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		} else if next != 0 {
			t.Errorf("unexpected cursor: have=%d want=%d", next, 0)
		} else if diff := cmp.Diff(have, []*campaigns.UserToken{
			// The seeming reverse order here is because of the earlier
			// createdAt that we used when upserting tokenAB.
			tokenAB,
			tokenAA,
		}); diff != "" {
			t.Errorf("unexpected tokens:\n%s", diff)
		}

		// If we ask for one token, then we should get a cursor indicating
		// there are more records.
		have, next, err := s.ListUserTokens(ctx, ListUserTokensOpts{
			UserID:      uidA,
			LimitOffset: &db.LimitOffset{Limit: 1},
		})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if next != 1 {
			t.Errorf("unexpected cursor: have=%d want=%d", next, 1)
		}
		if diff := cmp.Diff(have, []*campaigns.UserToken{
			tokenAB,
		}); diff != "" {
			t.Errorf("unexpected tokens:\n%s", diff)
		}

		// Now we should be able to get the other token by using the cursor.
		have, next, err = s.ListUserTokens(ctx, ListUserTokensOpts{
			UserID:      uidA,
			LimitOffset: &db.LimitOffset{Limit: 1, Offset: next},
		})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		// This time, we shouldn't have a next cursor.
		if next != 0 {
			t.Errorf("unexpected cursor: have=%d want=%d", next, 0)
		}
		if diff := cmp.Diff(have, []*campaigns.UserToken{
			tokenAA,
		}); diff != "" {
			t.Errorf("unexpected tokens:\n%s", diff)
		}

		// Next, let's ensure that listing tokens for a user without tokens
		// results in an empty slice.
		if have, next, err := s.ListUserTokens(ctx, ListUserTokensOpts{
			UserID: uidB,
		}); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		} else if next != 0 {
			t.Errorf("unexpected cursor: have=%d want=%d", next, 0)
		} else if len(have) != 0 {
			t.Errorf("unexpected non-empty slice: %v", have)
		}

		// We hope you enjoyed the user tests, because here are some
		// suspiciously similar scenarios for external service searches.

		// If we don't set any limits, then we should just get all the tokens
		// for the external service back.
		if have, next, err := s.ListUserTokens(ctx, ListUserTokensOpts{
			ExternalServiceID: svcA.ID,
		}); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		} else if next != 0 {
			t.Errorf("unexpected cursor: have=%d want=%d", next, 0)
		} else if diff := cmp.Diff(have, []*campaigns.UserToken{
			tokenAA,
		}); diff != "" {
			t.Errorf("unexpected tokens:\n%s", diff)
		}

		// If we ask for one token, then we should get a cursor indicating
		// there are no more records.
		have, next, err = s.ListUserTokens(ctx, ListUserTokensOpts{
			ExternalServiceID: svcA.ID,
			LimitOffset:       &db.LimitOffset{Limit: 1},
		})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if next != 0 {
			t.Errorf("unexpected cursor: have=%d want=%d", next, 0)
		}
		if diff := cmp.Diff(have, []*campaigns.UserToken{
			tokenAA,
		}); diff != "" {
			t.Errorf("unexpected tokens:\n%s", diff)
		}

		// Using a bogus cursor should result in no user tokens.
		have, next, err = s.ListUserTokens(ctx, ListUserTokensOpts{
			ExternalServiceID: svcA.ID,
			LimitOffset:       &db.LimitOffset{Limit: 1, Offset: 1},
		})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if next != 0 {
			t.Errorf("unexpected cursor: have=%d want=%d", next, 0)
		}
		if len(have) != 0 {
			t.Errorf("unexpected non-empty slice: %v", have)
		}

		// Finally, let's validate that combining a user ID with an external
		// service ID essentially has the same behaviour as GetUserToken().
		want, err := s.GetUserToken(ctx, uidA, svcB.ID)
		if err != nil {
			t.Fatal(err)
		}
		if have, _, err := s.ListUserTokens(ctx, ListUserTokensOpts{
			ExternalServiceID: svcB.ID,
			UserID:            uidA,
		}); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		} else if diff := cmp.Diff(have[0], want); diff != "" {
			t.Errorf("unexpected tokens:\n%s", diff)
		}
	})

	// Now let's try updating one of the tokens we've already created.
	t.Run("UpsertUserToken update", func(t *testing.T) {
		tokenAB.Token = createSecretString("quux")
		if err := s.UpsertUserToken(ctx, tokenAB); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if token, err := s.GetUserToken(ctx, tokenAB.UserID, tokenAB.ExternalServiceID); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		} else if have, want := *token.Token.S, "quux"; have != want {
			t.Errorf("unexpected token value: have=%q want=%q", have, want)
		}
	})

	// Finally, let's delete a token.
	t.Run("DeleteUserToken", func(t *testing.T) {
		if err := s.DeleteUserToken(ctx, uidA, svcB.ID); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		// Let's ensure it really was deleted!
		if _, err := s.GetUserToken(ctx, uidA, svcB.ID); err != ErrNoResults {
			t.Errorf("unexpected error: have=%v want=%v", err, ErrNoResults)
		}

		// Validate that deleting a non-existent token behaves as expected.
		if err := s.DeleteUserToken(ctx, uidB, svcB.ID); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
	})
}

func createExternalService(t *testing.T, ctx context.Context, rs repos.Store, kind string) *repos.ExternalService {
	t.Helper()
	es := repos.ExternalService{
		Kind:   kind,
		Config: "{}",
	}

	if err := rs.UpsertExternalServices(ctx, &es); err != nil {
		t.Fatal(err)
	}
	return &es
}

func createSecretString(s string) secret.StringValue {
	return secret.StringValue{S: &s}
}

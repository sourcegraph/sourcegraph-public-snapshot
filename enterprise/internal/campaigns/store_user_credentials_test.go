package campaigns

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/oauth1/oauth"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
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

	// And some times before the current time won't hurt either.
	createdAt := time.Date(2020, 10, 2, 23, 56, 0, 0, time.UTC)
	updatedAt := time.Date(2020, 10, 2, 23, 57, 0, 0, time.UTC)

	// Instead of two of every animal, we want one of every authenticator. Same,
	// same.
	auths := []auth.Authenticator{
		&auth.OAuthClient{Client: createOAuthClient(t, "abc", "def")},
		&auth.BasicAuth{Username: "foo", Password: "bar"},
		&auth.OAuthBearerToken{Token: "abcdef"},
		&bitbucketserver.SudoableOAuthClient{
			Client:   auth.OAuthClient{Client: createOAuthClient(t, "ghi", "jkl")},
			Username: "neo",
		},
		&gitlab.SudoableToken{Token: "mnop", Sudo: "qrs"},
	}

	// For each credential type, we're going to go through the lifecycle of a
	// user credential: upserting it, retrieving it back, upserting a different
	// type, retrieving _that_ back, and then finally deleting it and ensuring
	// it's really deleted. Buckle up.
	for _, initial := range auths {
		t.Run(
			fmt.Sprintf("%T", initial),
			func(t *testing.T) {
				var credA, credB *campaigns.UserCredential

				// Create with implicit created and updated times.
				credA = &campaigns.UserCredential{
					UserID:            uidA,
					ExternalServiceID: svcA.ID,
					Credential:        initial,
				}
				if err := s.UpsertUserToken(ctx, credA); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}

				// Now get it back and see what we have.
				if have, err := s.GetUserToken(ctx, uidA, svcA.ID); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				} else if diff := cmp.Diff(have, credA); diff != "" {
					t.Errorf("unexpected credential:\n%s", diff)
				}

				// Ensure that we don't get it back if we ask for a different
				// service.
				if _, err := s.GetUserToken(ctx, uidA, svcB.ID); err != ErrNoResults {
					t.Errorf("unexpected error: have=%v want=%v", err, ErrNoResults)
				}

				// Same for a different user.
				if _, err := s.GetUserToken(ctx, uidB, svcA.ID); err != ErrNoResults {
					t.Errorf("unexpected error: have=%v want=%v", err, ErrNoResults)
				}

				// Now let's list the user's tokens. To start with, we should
				// get just one.
				assertSingleCredential(t, func() ([]*campaigns.UserCredential, int, error) {
					return s.ListUserTokens(ctx, ListUserTokensOpts{
						UserID: uidA,
					})
				}, credA)

				// Similarly, asking for the service should result in the same
				// result.
				assertSingleCredential(t, func() ([]*campaigns.UserCredential, int, error) {
					return s.ListUserTokens(ctx, ListUserTokensOpts{
						ExternalServiceID: svcA.ID,
					})
				}, credA)

				// Now the (currently) negative cases for the other user and
				// external service.
				assertEmptyCredentials(t, func() ([]*campaigns.UserCredential, int, error) {
					return s.ListUserTokens(ctx, ListUserTokensOpts{
						UserID: uidB,
					})
				})

				assertEmptyCredentials(t, func() ([]*campaigns.UserCredential, int, error) {
					return s.ListUserTokens(ctx, ListUserTokensOpts{
						ExternalServiceID: svcB.ID,
					})
				})

				// Now let's add another credential for svcB so we can test the
				// cursor.
				credB = &campaigns.UserCredential{
					UserID:            uidA,
					ExternalServiceID: svcB.ID,
					Credential:        initial,
					CreatedAt:         createdAt,
					UpdatedAt:         updatedAt,
				}
				if err := s.UpsertUserToken(ctx, credB); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}

				// If we list for uidA with a limit of 1 record, we should get a
				// non-zero cursor this time.
				have, next, err := s.ListUserTokens(ctx, ListUserTokensOpts{
					UserID:      uidA,
					LimitOffset: &db.LimitOffset{Limit: 1},
				})
				if err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				} else if next != 1 {
					t.Errorf("unexpected cursor: have=%d want=%d", next, 1)
				} else if diff := cmp.Diff(have, []*campaigns.UserCredential{
					credB,
				}); diff != "" {
					t.Errorf("unexpected credential:\n%s", diff)
				}

				// Now we should be able to get the next credential with the
				// cursor.
				assertSingleCredential(t, func() ([]*campaigns.UserCredential, int, error) {
					return s.ListUserTokens(ctx, ListUserTokensOpts{
						UserID:      uidA,
						LimitOffset: &db.LimitOffset{Limit: 1, Offset: next},
					})
				}, credA)

				// Finally, let's ensure we really get two credentials back if
				// we don't set a limit.
				if have, next, err := s.ListUserTokens(ctx, ListUserTokensOpts{
					UserID: uidA,
				}); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				} else if next != 0 {
					t.Errorf("unexpected cursor: have=%d want=%d", next, 0)
				} else if diff := cmp.Diff(have, []*campaigns.UserCredential{
					credB,
					credA,
				}); diff != "" {
					t.Errorf("unexpected credentials:\n%s", diff)
				}

				// Whew. OK. Now we're going to upsert a credential repeatedly
				// with different types to ensure the type marshalling works as
				// expected even when changing types.
				for _, subsequent := range auths {
					t.Run(fmt.Sprintf("%T", subsequent), func(t *testing.T) {
						credA.Credential = subsequent
						if err := s.UpsertUserToken(ctx, credA); err != nil {
							t.Errorf("unexpected non-nil error: %v", err)
						}

						if have, err := s.GetUserToken(ctx, uidA, svcA.ID); err != nil {
							t.Errorf("unexpected non-nil error: %v", err)
						} else if diff := cmp.Diff(have, credA); diff != "" {
							t.Errorf("unexpected credential:\n%s", diff)
						}
					})
				}

				// Finally, let's delete both the credentials we created.
				for _, svcID := range []int64{svcA.ID, svcB.ID} {
					if err := s.DeleteUserToken(ctx, uidA, svcID); err != nil {
						t.Errorf("unexpected non-nil error: %v", err)
					}
				}

				if have, _, err := s.ListUserTokens(ctx, ListUserTokensOpts{
					UserID: uidA,
				}); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				} else if len(have) != 0 {
					t.Errorf("unexpected still extant credentials: %v", have)
				}
			},
		)

	}
}

func assertEmptyCredentials(t *testing.T, get func() ([]*campaigns.UserCredential, int, error)) {
	have, next, err := get()

	if err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}
	if next != 0 {
		t.Errorf("unexpected non-zero cursor: %d", next)
	}
	if len(have) != 0 {
		t.Errorf("unexpected non-empty credentials: %v", have)
	}
}

func assertSingleCredential(t *testing.T, get func() ([]*campaigns.UserCredential, int, error), want *campaigns.UserCredential) {
	have, next, err := get()

	if err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}
	if next != 0 {
		t.Errorf("unexpected non-zero cursor: %d", next)
	}
	if diff := cmp.Diff(have, []*campaigns.UserCredential{want}); diff != "" {
		t.Errorf("unexpected credential:\n%s", diff)
	}
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

func createOAuthClient(t *testing.T, token, secret string) *oauth.Client {
	// Generate a random key so we can test different clients are different.
	// Note that this is wildly insecure.
	key, err := rsa.GenerateKey(rand.Reader, 64)
	if err != nil {
		t.Fatal(err)
	}

	return &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  "abc",
			Secret: "def",
		},
		PrivateKey: key,
	}
}

func createSecretString(s string) secret.StringValue {
	return secret.StringValue{S: &s}
}

package db

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"math/big"
	"net/http"
	"reflect"
	"testing"

	"github.com/gomodule/oauth1/oauth"
	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUserCredentials_Create(t *testing.T) {
	ctx, user := setUpUserCredentialTest(t)

	// Versions of Go before 1.14.x (where 3 < x < 11) cannot diff *big.Int
	// fields, which causes problems when diffing the keys embedded in OAuth
	// clients. We'll just define a little helper here to wrap cmp.Diff and get
	// us over the hump.
	diffAuthenticators := func(a, b auth.Authenticator) string {
		return cmp.Diff(a, b, cmp.Comparer(func(a, b *big.Int) bool {
			return a.Cmp(b) == 0
		}))
	}

	// Instead of two of every animal, we want one of every authenticator. Same,
	// same.
	for name, auth := range createUserCredentialAuths(t) {
		t.Run(name, func(t *testing.T) {
			scope := UserCredentialScope{
				Domain:              name,
				UserID:              user.ID,
				ExternalServiceType: "github",
				ExternalServiceID:   "https://github.com",
			}

			cred, err := UserCredentials.Create(ctx, scope, auth)
			if err != nil {
				t.Errorf("unexpected non-nil error: %v", err)
			} else if cred == nil {
				t.Error("unexpected nil credential")
			}

			if cred.ID == 0 {
				t.Error("unexpected zero ID")
			}
			if cred.Domain != scope.Domain {
				t.Errorf("unexpected domain: have=%q want=%q", cred.Domain, scope.Domain)
			}
			if cred.UserID != scope.UserID {
				t.Errorf("unexpected ID: have=%d want=%d", cred.UserID, scope.UserID)
			}
			if cred.ExternalServiceType != scope.ExternalServiceType {
				t.Errorf("unexpected external service type: have=%q want=%q", cred.ExternalServiceType, scope.ExternalServiceType)
			}
			if cred.ExternalServiceID != scope.ExternalServiceID {
				t.Errorf("unexpected external service id: have=%q want=%q", cred.ExternalServiceID, scope.ExternalServiceID)
			}
			if diff := diffAuthenticators(cred.Credential, auth); diff != "" {
				t.Errorf("unexpected credential:\n%s", diff)
			}
			if cred.CreatedAt.IsZero() {
				t.Errorf("unexpected zero creation time")
			}
			if cred.UpdatedAt.IsZero() {
				t.Errorf("unexpected zero update time")
			}

			// Ensure that trying to insert again fails.
			if cred, err := UserCredentials.Create(ctx, scope, auth); err == nil {
				t.Error("unexpected nil error")
			} else if cred != nil {
				t.Errorf("unexpected non-nil credential: %v", cred)
			}
		})
	}
}

func TestUserCredentials_Delete(t *testing.T) {
	ctx, user := setUpUserCredentialTest(t)

	t.Run("nonextant", func(t *testing.T) {
		err := UserCredentials.Delete(ctx, 1)
		if err == nil {
			t.Error("unexpected nil error")
		}

		e, ok := err.(UserCredentialNotFoundErr)
		if !ok {
			t.Errorf("error is not a userCredentialNotFoundError; got %T: %v", err, err)
		}
		if len(e.args) != 1 || e.args[0].(int64) != 1 {
			t.Errorf("unexpected args: have=%v want=[1]", e.args)
		}
	})

	t.Run("extant", func(t *testing.T) {
		scope := UserCredentialScope{
			Domain:              UserCredentialDomainCampaigns,
			UserID:              user.ID,
			ExternalServiceType: "github",
			ExternalServiceID:   "https://github.com",
		}
		token := &auth.OAuthBearerToken{Token: "abcdef"}

		cred, err := UserCredentials.Create(ctx, scope, token)
		if err != nil {
			t.Fatal(err)
		}

		if err := UserCredentials.Delete(ctx, cred.ID); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		_, err = UserCredentials.GetByID(ctx, cred.ID)
		if _, ok := err.(UserCredentialNotFoundErr); !ok {
			t.Errorf("unexpected error retrieving credential after deletion: %v", err)
		}
	})
}

func TestUserCredentials_GetByID(t *testing.T) {
	ctx, user := setUpUserCredentialTest(t)

	t.Run("nonextant", func(t *testing.T) {
		cred, err := UserCredentials.GetByID(ctx, 1)
		if cred != nil {
			t.Errorf("unexpected non-nil credential: %v", cred)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}

		e, ok := err.(UserCredentialNotFoundErr)
		if !ok {
			t.Errorf("error is not a userCredentialNotFoundError; got %T: %v", err, err)
		}
		if len(e.args) != 1 || e.args[0].(int64) != 1 {
			t.Errorf("unexpected args: have=%v want=[1]", e.args)
		}
	})

	t.Run("extant", func(t *testing.T) {
		scope := UserCredentialScope{
			Domain:              UserCredentialDomainCampaigns,
			UserID:              user.ID,
			ExternalServiceType: "github",
			ExternalServiceID:   "https://github.com",
		}
		token := &auth.OAuthBearerToken{Token: "abcdef"}

		want, err := UserCredentials.Create(ctx, scope, token)
		if err != nil {
			t.Fatal(err)
		}

		have, err := UserCredentials.GetByID(ctx, want.ID)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if diff := cmp.Diff(have, want); diff != "" {
			t.Errorf("unexpected credential:\n%s", diff)
		}
	})
}

func TestUserCredentials_GetByScope(t *testing.T) {
	ctx, user := setUpUserCredentialTest(t)

	scope := UserCredentialScope{
		Domain:              UserCredentialDomainCampaigns,
		UserID:              user.ID,
		ExternalServiceType: "github",
		ExternalServiceID:   "https://github.com",
	}
	token := &auth.OAuthBearerToken{Token: "abcdef"}

	t.Run("nonextant", func(t *testing.T) {
		cred, err := UserCredentials.GetByScope(ctx, scope)
		if cred != nil {
			t.Errorf("unexpected non-nil credential: %v", cred)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}

		e, ok := err.(UserCredentialNotFoundErr)
		if !ok {
			t.Errorf("error is not a userCredentialNotFoundError; got %T: %v", err, err)
		}
		if diff := cmp.Diff(e.args, []interface{}{scope}); diff != "" {
			t.Errorf("unexpected args:\n%s", diff)
		}
	})

	t.Run("extant", func(t *testing.T) {
		want, err := UserCredentials.Create(ctx, scope, token)
		if err != nil {
			t.Fatal(err)
		}

		have, err := UserCredentials.GetByScope(ctx, scope)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if diff := cmp.Diff(have, want); diff != "" {
			t.Errorf("unexpected credential:\n%s", diff)
		}
	})
}

func TestUserCredentials_List(t *testing.T) {
	ctx, user := setUpUserCredentialTest(t)

	githubScope := UserCredentialScope{
		Domain:              UserCredentialDomainCampaigns,
		UserID:              user.ID,
		ExternalServiceType: "github",
		ExternalServiceID:   "https://github.com",
	}
	gitlabScope := UserCredentialScope{
		Domain:              UserCredentialDomainCampaigns,
		UserID:              user.ID,
		ExternalServiceType: "gitlab",
		ExternalServiceID:   "https://gitlab.com",
	}
	token := &auth.OAuthBearerToken{Token: "abcdef"}

	// Unlike the other tests in this file, we'll set up a couple of credentials
	// right now, and then list from there.
	github, err := UserCredentials.Create(ctx, githubScope, token)
	if err != nil {
		t.Fatal(err)
	}

	gitlab, err := UserCredentials.Create(ctx, gitlabScope, token)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("not found", func(t *testing.T) {
		creds, next, err := UserCredentials.List(ctx, UserCredentialsListOpts{
			Scope: UserCredentialScope{
				Domain: "this is not a valid domain",
			},
		})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if next != 0 {
			t.Errorf("unexpected next: have=%d want=%d", next, 0)
		}
		if len(creds) != 0 {
			t.Errorf("unexpected non-zero number of credentials: %v", creds)
		}
	})

	for name, tc := range map[string]struct {
		scope UserCredentialScope
		want  *UserCredential
	}{
		"service ID only": {
			scope: UserCredentialScope{
				ExternalServiceID: "https://github.com",
			},
			want: github,
		},
		"service type only": {
			scope: UserCredentialScope{
				ExternalServiceType: "gitlab",
			},
			want: gitlab,
		},
		"full scope": {
			scope: githubScope,
			want:  github,
		},
	} {
		t.Run("single match on "+name, func(t *testing.T) {
			creds, next, err := UserCredentials.List(ctx, UserCredentialsListOpts{
				Scope: tc.scope,
			})
			if err != nil {
				t.Errorf("unexpected non-nil error: %v", err)
			}
			if next != 0 {
				t.Errorf("unexpected next: have=%d want=%d", next, 0)
			}
			if diff := cmp.Diff(creds, []*UserCredential{tc.want}); diff != "" {
				t.Errorf("unexpected credentials:\n%s", diff)
			}
		})
	}

	for name, scope := range map[string]UserCredentialScope{
		"no options":   {},
		"domain only":  {Domain: UserCredentialDomainCampaigns},
		"user ID only": {UserID: user.ID},
		"domain and user ID": {
			Domain: UserCredentialDomainCampaigns,
			UserID: user.ID,
		},
	} {
		t.Run("multiple matches on "+name, func(t *testing.T) {
			creds, next, err := UserCredentials.List(ctx, UserCredentialsListOpts{
				Scope: scope,
			})
			if err != nil {
				t.Errorf("unexpected non-nil error: %v", err)
			}
			if next != 0 {
				t.Errorf("unexpected next: have=%d want=%d", next, 0)
			}
			if diff := cmp.Diff(creds, []*UserCredential{github, gitlab}); diff != "" {
				t.Errorf("unexpected credentials:\n%s", diff)
			}
		})

		t.Run("pagination for "+name, func(t *testing.T) {
			creds, next, err := UserCredentials.List(ctx, UserCredentialsListOpts{
				LimitOffset: &LimitOffset{Limit: 1},
				Scope:       scope,
			})
			if err != nil {
				t.Errorf("unexpected non-nil error: %v", err)
			}
			if next != 1 {
				t.Errorf("unexpected next: have=%d want=%d", next, 1)
			}
			if diff := cmp.Diff(creds, []*UserCredential{github}); diff != "" {
				t.Errorf("unexpected credentials:\n%s", diff)
			}

			creds, next, err = UserCredentials.List(ctx, UserCredentialsListOpts{
				LimitOffset: &LimitOffset{Limit: 1, Offset: next},
				Scope:       scope,
			})
			if err != nil {
				t.Errorf("unexpected non-nil error: %v", err)
			}
			if next != 0 {
				t.Errorf("unexpected next: have=%d want=%d", next, 0)
			}
			if diff := cmp.Diff(creds, []*UserCredential{gitlab}); diff != "" {
				t.Errorf("unexpected credentials:\n%s", diff)
			}
		})
	}
}

func TestUserCredentials_Invalid(t *testing.T) {
	ctx, user := setUpUserCredentialTest(t)

	t.Run("marshal", func(t *testing.T) {
		if _, err := UserCredentials.Create(ctx, UserCredentialScope{}, &invalidAuth{}); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("unmarshal", func(t *testing.T) {
		// We'll set up some cases here that just shouldn't happen at all, and
		// make sure they bubble up with errors where we expect. Let's define a
		// helper to make that easier.

		insertRawCredential := func(t *testing.T, domain string, raw string) int64 {
			q := sqlf.Sprintf(
				userCredentialsCreateQueryFmtstr,
				domain,
				user.ID,
				"type",
				"id",
				raw,
				sqlf.Sprintf("id"),
			)

			var id int64
			row := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
			if err := row.Scan(&id); err != nil {
				t.Fatal(err)
			}

			return id
		}

		for name, id := range map[string]int64{
			"invalid credential type": insertRawCredential(t, "invalid", `{"type":"InvalidType"}`),
			"lying credential type":   insertRawCredential(t, "lying", `{"type":"BasicAuth","username":42}`),
			"malformed JSON":          insertRawCredential(t, "malformed", "this is not valid JSON"),
		} {
			t.Run(name, func(t *testing.T) {
				if _, err := UserCredentials.GetByID(ctx, id); err == nil {
					t.Error("unexpected nil error")
				} else if _, ok := err.(UserCredentialNotFoundErr); ok {
					t.Error("unexpected not found error")
				}
			})
		}
	})
}

func TestUserCredentialNotFoundErr(t *testing.T) {
	err := UserCredentialNotFoundErr{}
	if have := errcode.IsNotFound(err); !have {
		t.Error("UserCredentialNotFoundErr does not say it represents a not found error")
	}
}

func createUserCredentialAuths(t *testing.T) map[string]auth.Authenticator {
	t.Helper()

	createOAuthClient := func(t *testing.T, token, secret string) *oauth.Client {
		t.Helper()

		// Generate a random key so we can test different clients are different.
		// Note that this is wildly insecure.
		key, err := rsa.GenerateKey(rand.Reader, 64)
		if err != nil {
			t.Fatal(err)
		}

		return &oauth.Client{
			Credentials: oauth.Credentials{
				Token:  token,
				Secret: secret,
			},
			PrivateKey: key,
		}
	}

	auths := make(map[string]auth.Authenticator)
	for _, a := range []auth.Authenticator{
		&auth.OAuthClient{Client: createOAuthClient(t, "abc", "def")},
		&auth.BasicAuth{Username: "foo", Password: "bar"},
		&auth.OAuthBearerToken{Token: "abcdef"},
		&bitbucketserver.SudoableOAuthClient{
			Client:   auth.OAuthClient{Client: createOAuthClient(t, "ghi", "jkl")},
			Username: "neo",
		},
		&gitlab.SudoableToken{Token: "mnop", Sudo: "qrs"},
	} {
		auths[reflect.TypeOf(a).String()] = a
	}

	return auths
}

func setUpUserCredentialTest(t *testing.T) (context.Context, *types.User) {
	if testing.Short() {
		t.Skip()
	}

	t.Helper()
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	// Create a user that allows us to link the credential somewhere.
	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u2",
		Password:              "pw",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	return ctx, user
}

type invalidAuth struct{}

var _ auth.Authenticator = &invalidAuth{}

func (*invalidAuth) Authenticate(_ *http.Request) error { panic("should not be called") }
func (*invalidAuth) Hash() string                       { panic("should not be called") }

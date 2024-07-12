package database

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"reflect"
	"testing"

	"github.com/gomodule/oauth1/oauth"
	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	ghauth "github.com/sourcegraph/sourcegraph/internal/github_apps/auth"
	ghtypes "github.com/sourcegraph/sourcegraph/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestUserCredential_Authenticator(t *testing.T) {
	ctx := context.Background()

	t.Run("errors", func(t *testing.T) {
		testKey := &et.TestKey{}
		transparentKey := et.NewTransparentKey(t)

		for name, credential := range map[string]*UserCredential{
			"no credential": {
				Credential: NewEncryptedCredential("", testEncryptionKeyID(testKey), testKey),
			},
			"bad decrypter": {
				Credential: NewEncryptedCredential("foo", "it's the bad guy... uh, key", &et.BadKey{Err: errors.New("bad key bad key what you gonna do")}),
			},
			"invalid secret": {
				Credential: NewEncryptedCredential("foo", testEncryptionKeyID(transparentKey), transparentKey),
			},
		} {
			t.Run(name, func(t *testing.T) {
				if _, err := credential.Authenticator(ctx, ghauth.CreateAuthenticatorForCredentialOpts{}); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("plaintext credential", func(t *testing.T) {
		a := &auth.BasicAuth{}

		enc, _, err := EncryptAuthenticator(ctx, nil, a)
		if err != nil {
			t.Fatal(err)
		}

		for _, keyID := range []string{"", encryption.UnmigratedEncryptionKeyID} {
			t.Run(keyID, func(t *testing.T) {
				uc := &UserCredential{
					Credential: NewEncryptedCredential(string(enc), keyID, et.TestKey{}),
				}

				have, err := uc.Authenticator(ctx, ghauth.CreateAuthenticatorForCredentialOpts{})
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(have, a); diff != "" {
					t.Errorf("unexpected authenticator (-have +want):\n%s", diff)
				}
			})
		}
	})

	t.Run("encrypted credential", func(t *testing.T) {
		key := et.TestKey{}
		a := &auth.BasicAuth{Username: "foo", Password: "bar"}

		enc, kid, err := EncryptAuthenticator(ctx, key, a)
		if err != nil {
			t.Fatal(err)
		}
		uc := &UserCredential{
			Credential: NewEncryptedCredential(string(enc), kid, key),
		}

		have, err := uc.Authenticator(ctx, ghauth.CreateAuthenticatorForCredentialOpts{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if diff := cmp.Diff(have, a); diff != "" {
			t.Errorf("unexpected authenticator (-have +want):\n%s", diff)
		}
	})

	t.Run("nil key", func(t *testing.T) {
		a := &auth.BasicAuth{Username: "foo", Password: "bar"}

		enc, _, err := EncryptAuthenticator(ctx, nil, a)
		if err != nil {
			t.Fatal(err)
		}
		uc := &UserCredential{
			Credential: NewEncryptedCredential(string(enc), "", nil),
		}

		have, err := uc.Authenticator(ctx, ghauth.CreateAuthenticatorForCredentialOpts{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if diff := cmp.Diff(have, a); diff != "" {
			t.Errorf("unexpected authenticator (-have +want):\n%s", diff)
		}
	})
}

func TestUserCredential_SetAuthenticator(t *testing.T) {
	ctx := context.Background()
	a := &auth.BasicAuth{Username: "foo", Password: "bar"}

	t.Run("error", func(t *testing.T) {
		badKey := &et.BadKey{Err: errors.New("error")}
		uc := &UserCredential{
			Credential: NewEncryptedCredential("encoded", "bad key", badKey),
		}

		if err := uc.SetAuthenticator(ctx, a); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if _, _, err := uc.Credential.Encrypt(ctx, badKey); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("success", func(t *testing.T) {
		for name, key := range map[string]encryption.Key{
			"":         nil,
			"test key": et.TestKey{},
		} {
			t.Run(name, func(t *testing.T) {
				uc := &UserCredential{
					Credential: NewUnencryptedCredential(nil),
				}

				if err := uc.SetAuthenticator(ctx, a); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else {
					ctx := context.Background()
					_, keyID, err := uc.Credential.Encrypt(ctx, key)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}

					if key == nil && keyID != "" {
						t.Errorf("unexpected non-empty key ID: %q", keyID)
					} else if key != nil && keyID == "" {
						t.Error("unexpected empty key ID")
					}
				}
			})
		}
	})
}

func createGitHubApp(ctx context.Context, t *testing.T, db DB) *ghtypes.GitHubApp {
	testPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAmDthV4Ab7Kx8g8xRY1MTbheTrT2DsPsoxpzEe69QlPd2WkhT
dM/VMUL+qxwJqBsC6fnSY2GdoP5Juv05zjVRo0g4fE8QhB/nD+ZCMkTeC9zD1vbb
Gd+hXecqjW1Qq8ad4CSgB5nFB2Djc0OlscTghAZL1ZJ8rLLRvIV4WznBJV60iMkV
T0LhLkwUEgCWoRW0GAihVPd65TteNoYDyQ5B4m8/D4S4BFO79xKNtyazkijVMrF5
6eYS00YTskMljaVwPdcMLE8BEzg9ipQygxyHKShHTEloUUMmukts+PUVSmKWZgE1
uEKtpYQ1lrshu/yGoXrsj7pvgRdBRZgyn0GPNwIDAQABAoIBAQCXS7zM5+fY6vy9
SJ1C59gRvKDqto5hoNy/uCKXAoBF7UPVKri3Ca/Ky9irWqxGRMI6pC1y1BuDW/cP
Pojq5pcCfs6UzUeO6N4OMTxtFYDRrVF+Hc1YA6gu2YazFIfukPFrSTs7Epp9YM/t
SLgu24p/7HoGAxah1P8aLFSX5eiOJ+8t8byYOrKLp3Rn67lC9Y+9LX4X6GHlBMDc
WHYupi3ZA7Q59dXQCJHFNG/hk17AMtB8lFra9rUid8teX8ZJKJQ26hU2O0UMujjM
mFlCdmvc97lJ4LhjrWHv/9yacf90bViHIkL52Yux1jNt/jl3/7CyBwHbau4b0qoZ
QkM4WIihAoGBAMlzsUeJxBCbUyMd/6TiLn60SDn0HMf4ZVdGqqxkhopVmsvRTn+P
wu9YHWFPwXwVL3wdtuBjsEA9nMKWWMQKbQUZhm1Y+AQIVpVNQqesgyLctVoIUBNY
fglvKrs8JuRuwMpE2P/3lXMsxtV9AyCpxxXhya8KqJa2jcMB/Lr+lx+fAoGBAMFz
16yHU+Zo6OOvy7T2xh67btwOrS2kIzhGO6gcK7qabkGGeaKLShnMYEmFGp4EaHTf
OVie+SU0OWF/e5bgFWC+fm6jWyhO0xPRbvs2P+l2KtnT2UBT9IgjhrVUIzp+Vn7t
cjfb32m7km1kZZ48ySP9cH/4/xnT6XEC33PoNwlpAoGAG1t+w7xNyAOP8sDsKrQc
pFBPTq98CRwOhx+tpeOw8bBWaT9vbZtUWbSZqNFv8S3fWPegEjD3ioHTfAl23Iid
7Ydd3hOq+sE3IOdxGdwvotheOG/QkBAAbb+PCgZNMdBolg9reLdisFVwWyWy+wiT
ZMFY5lCIPI9mCQmIDMzuMPkCgYBFJKJxh+z07YpP1wV4KLunQFbfUF+VcJUmB/RK
ocb/azL9OJNBBYf2sJW5sVlSIUE0hJR6mFd0dLYNowMJag46Bdwqrzhlr8bBzplc
MIenahTmxlFgLKG6Bvie1vPAdGd19mhcjrnLkL9FWhz38cHymyMammSTVqqZOe2j
/9usAQKBgQCT//j6XflAr20gb+mNcoJqVxRTFtSsZa23kJnZ3Sfs3R8XXu5NcZEZ
ODI9ZpZ9tg8oK9EB5vqMFTTnyjpar7F2jqFWtUmNju/rGlrQCZx0we+EAW/R2hFP
YGYu4Z+SyXTsv/Ys5VGWuuCJO32RuRBeC4eJCmpyH0mqPhIBZmV4Jw==
-----END RSA PRIVATE KEY-----
`
	app := &ghtypes.GitHubApp{
		AppID:      2093849,
		Domain:     "batches",
		PrivateKey: testPrivateKey,
		Kind:       ghtypes.UserCredentialGitHubAppKind,
	}
	ghAppID, err := db.GitHubApps().Create(ctx, app)
	if err != nil {
		t.Fatal(err)
	}
	app.ID = ghAppID
	return app
}

func TestUserCredentials_CreateUpdate(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	fx := setUpUserCredentialTest(t, db)
	ctx := context.Background()

	// Authorisation failure tests. (We'll test the happy path below.)
	t.Run("unauthorised", func(t *testing.T) {
		for name, tc := range authFailureTestCases(t, fx) {
			t.Run(name, func(t *testing.T) {
				tc.setup(t)

				scope := UserCredentialScope{
					Domain:              name,
					UserID:              tc.user.ID,
					ExternalServiceType: extsvc.TypeBitbucketCloud,
					ExternalServiceID:   "https://bitbucket.org",
				}
				basicAuth := &auth.BasicAuth{}

				// Attempt to create with the invalid context.
				cred, err := fx.db.Create(tc.ctx, scope, basicAuth)
				assert.Error(t, err)
				assert.Nil(t, cred)

				// Now we'll create a credential so we can test update.
				cred, err = fx.db.Create(fx.internalCtx, scope, basicAuth)
				require.NoError(t, err)
				require.NotNil(t, cred)

				// And let's test that we can't update either.
				err = fx.db.Update(tc.ctx, cred)
				assert.Error(t, err)
			})
		}
	})

	// Instead of two of every animal, we want one of every authenticator. Same,
	// same.
	for name, authenticator := range createUserCredentialAuths(t) {
		t.Run(name, func(t *testing.T) {
			scope := UserCredentialScope{
				Domain:              name,
				UserID:              fx.user.ID,
				ExternalServiceType: extsvc.TypeGitHub,
				ExternalServiceID:   "https://github.com",
			}

			cred, err := fx.db.Create(fx.userCtx, scope, authenticator)
			t.Cleanup(func() {
				if cred != nil {
					fx.db.Delete(fx.userCtx, cred.ID)
				}
			})
			assert.NoError(t, err)
			assert.NotNil(t, cred)
			assert.NotZero(t, cred.ID)
			assert.Equal(t, scope.Domain, cred.Domain)
			assert.Equal(t, scope.UserID, cred.UserID)
			assert.Equal(t, scope.ExternalServiceType, cred.ExternalServiceType)
			assert.Equal(t, scope.ExternalServiceID, cred.ExternalServiceID)
			assert.NotZero(t, cred.CreatedAt)
			assert.NotZero(t, cred.UpdatedAt)

			have, err := cred.Authenticator(fx.userCtx, ghauth.CreateAuthenticatorForCredentialOpts{})
			assert.NoError(t, err)
			assert.Equal(t, authenticator.Hash(), have.Hash())

			// Ensure that trying to insert again fails.
			second, err := fx.db.Create(fx.userCtx, scope, authenticator)
			assert.Error(t, err)
			assert.Nil(t, second)

			// Valid update contexts.
			newExternalServiceType := extsvc.TypeGitLab
			cred.ExternalServiceType = newExternalServiceType

			err = fx.db.Update(fx.userCtx, cred)
			assert.NoError(t, err)

			updatedCred, err := fx.db.GetByID(fx.userCtx, cred.ID)
			assert.NoError(t, err)
			assert.Equal(t, cred, updatedCred)
		})
	}

	t.Run("with github app", func(t *testing.T) {
		ghApp := createGitHubApp(ctx, t, db)
		t.Cleanup(func() {
			db.GitHubApps().Delete(ctx, ghApp.ID)
		})

		scope := UserCredentialScope{
			Domain:              "batches",
			UserID:              fx.user.ID,
			ExternalServiceType: extsvc.TypeGitHub,
			ExternalServiceID:   "https://github.com",
			GitHubAppID:         ghApp.ID,
		}

		auther, err := ghauth.NewGitHubAppAuthenticator(ghApp.AppID, []byte(ghApp.PrivateKey))
		assert.NoError(t, err)

		cred, err := fx.db.Create(fx.userCtx, scope, auther)
		t.Cleanup(func() {
			if cred != nil {
				fx.db.Delete(fx.userCtx, cred.ID)
			}
		})
		assert.NoError(t, err)
		assert.NotNil(t, cred)
		assert.NotZero(t, cred.ID)
		assert.Equal(t, scope.Domain, cred.Domain)
		assert.Equal(t, scope.UserID, cred.UserID)
		assert.Equal(t, scope.ExternalServiceType, cred.ExternalServiceType)
		assert.Equal(t, scope.ExternalServiceID, cred.ExternalServiceID)
		assert.NotZero(t, cred.CreatedAt)
		assert.NotZero(t, cred.UpdatedAt)

		have, err := cred.Authenticator(fx.userCtx, ghauth.CreateAuthenticatorForCredentialOpts{
			GitHubAppStore: db.GitHubApps(),
		})
		assert.NoError(t, err)
		assert.Equal(t, auther.Hash(), have.Hash())

		// Ensure that trying to insert again fails.
		second, err := fx.db.Create(fx.userCtx, scope, auther)
		assert.Error(t, err)
		assert.Nil(t, second)
	})
}

func TestUserCredentials_Delete(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	fx := setUpUserCredentialTest(t, db)

	t.Run("nonextant", func(t *testing.T) {
		err := fx.db.Delete(fx.internalCtx, 1)
		assertUserCredentialNotFoundError(t, 1, err)
	})

	t.Run("no permissions", func(t *testing.T) {
		for name, tc := range authFailureTestCases(t, fx) {
			t.Run(name, func(t *testing.T) {
				tc.setup(t)

				scope := UserCredentialScope{
					Domain:              UserCredentialDomainBatches,
					UserID:              tc.user.ID,
					ExternalServiceType: "github",
					ExternalServiceID:   "https://github.com",
				}
				token := &auth.OAuthBearerToken{Token: "abcdef"}

				cred, err := fx.db.Create(fx.internalCtx, scope, token)
				require.NoError(t, err)
				t.Cleanup(func() { fx.db.Delete(fx.internalCtx, cred.ID) })

				err = fx.db.Delete(tc.ctx, cred.ID)
				assert.Error(t, err)
			})
		}
	})

	t.Run("extant", func(t *testing.T) {
		scope := UserCredentialScope{
			Domain:              UserCredentialDomainBatches,
			UserID:              fx.user.ID,
			ExternalServiceType: "github",
			ExternalServiceID:   "https://github.com",
		}
		token := &auth.OAuthBearerToken{Token: "abcdef"}

		cred, err := fx.db.Create(fx.internalCtx, scope, token)
		require.NoError(t, err)

		err = fx.db.Delete(fx.userCtx, cred.ID)
		assert.NoError(t, err)

		_, err = fx.db.GetByID(fx.internalCtx, cred.ID)
		assert.ErrorAs(t, err, &UserCredentialNotFoundErr{})
	})
}

func TestUserCredentials_GetByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	fx := setUpUserCredentialTest(t, db)

	t.Run("nonextant", func(t *testing.T) {
		cred, err := fx.db.GetByID(fx.internalCtx, 1)
		assert.Nil(t, cred)
		assertUserCredentialNotFoundError(t, 1, err)
	})

	t.Run("no permissions", func(t *testing.T) {
		for name, tc := range authFailureTestCases(t, fx) {
			t.Run(name, func(t *testing.T) {
				tc.setup(t)

				scope := UserCredentialScope{
					Domain:              UserCredentialDomainBatches,
					UserID:              tc.user.ID,
					ExternalServiceType: "github",
					ExternalServiceID:   "https://github.com",
				}
				token := &auth.OAuthBearerToken{Token: "abcdef"}

				cred, err := fx.db.Create(fx.internalCtx, scope, token)
				require.NoError(t, err)
				t.Cleanup(func() { fx.db.Delete(fx.internalCtx, cred.ID) })

				_, err = fx.db.GetByID(tc.ctx, cred.ID)
				assert.Error(t, err)
			})
		}
	})

	t.Run("extant", func(t *testing.T) {
		scope := UserCredentialScope{
			Domain:              UserCredentialDomainBatches,
			UserID:              fx.user.ID,
			ExternalServiceType: "github",
			ExternalServiceID:   "https://github.com",
		}
		token := &auth.OAuthBearerToken{Token: "abcdef"}

		want, err := fx.db.Create(fx.internalCtx, scope, token)
		require.NoError(t, err)

		have, err := fx.db.GetByID(fx.userCtx, want.ID)
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}

func TestUserCredentials_GetByScope(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	fx := setUpUserCredentialTest(t, db)

	scope := UserCredentialScope{
		Domain:              UserCredentialDomainBatches,
		UserID:              fx.user.ID,
		ExternalServiceType: "github",
		ExternalServiceID:   "https://github.com",
	}
	token := &auth.OAuthBearerToken{Token: "abcdef"}

	t.Run("nonextant", func(t *testing.T) {
		cred, err := fx.db.GetByScope(fx.internalCtx, scope)
		assert.Nil(t, cred)
		assertUserCredentialNotFoundError(t, scope, err)
	})

	t.Run("no permissions", func(t *testing.T) {
		for name, tc := range authFailureTestCases(t, fx) {
			t.Run(name, func(t *testing.T) {
				tc.setup(t)

				s := scope
				s.UserID = tc.user.ID

				cred, err := fx.db.Create(fx.internalCtx, s, token)
				require.NoError(t, err)
				t.Cleanup(func() { fx.db.Delete(fx.internalCtx, cred.ID) })

				_, err = fx.db.GetByScope(tc.ctx, scope)
				assert.Error(t, err)
			})
		}
	})

	t.Run("extant", func(t *testing.T) {
		want, err := fx.db.Create(fx.internalCtx, scope, token)
		require.NoError(t, err)
		require.NotNil(t, want)

		have, err := fx.db.GetByScope(fx.userCtx, scope)
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("extant github app credential", func(t *testing.T) {
		ghApp := createGitHubApp(fx.user2Ctx, t, db)

		auther, err := ghauth.NewGitHubAppAuthenticator(ghApp.AppID, []byte(ghApp.PrivateKey))
		require.NoError(t, err)

		ghAppScope := scope
		ghAppScope.GitHubAppID = ghApp.ID
		ghAppScope.UserID = fx.user2.ID

		want, err := fx.db.Create(fx.user2Ctx, ghAppScope, auther)
		require.NoError(t, err)
		require.NotNil(t, want)

		have, err := fx.db.GetByScope(fx.user2Ctx, ghAppScope)
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}

func TestUserCredentials_List(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	fx := setUpUserCredentialTest(t, db)

	ghApp := createGitHubApp(fx.user2Ctx, t, db)

	auther, err := ghauth.NewGitHubAppAuthenticator(ghApp.AppID, []byte(ghApp.PrivateKey))
	require.NoError(t, err)

	githubScope := UserCredentialScope{
		Domain:              UserCredentialDomainBatches,
		UserID:              fx.user.ID,
		ExternalServiceType: "github",
		ExternalServiceID:   "https://github.com",
	}
	githubScopeWithGitHubApp := UserCredentialScope{
		Domain:              UserCredentialDomainBatches,
		UserID:              fx.user2.ID,
		ExternalServiceType: "github",
		ExternalServiceID:   "https://github.com",
		GitHubAppID:         ghApp.ID,
	}
	gitlabScope := UserCredentialScope{
		Domain:              UserCredentialDomainBatches,
		UserID:              fx.user.ID,
		ExternalServiceType: "gitlab",
		ExternalServiceID:   "https://gitlab.com",
	}
	adminScope := UserCredentialScope{
		Domain:              UserCredentialDomainBatches,
		UserID:              fx.admin.ID,
		ExternalServiceType: "gitlab",
		ExternalServiceID:   "https://gitlab.com",
	}
	token := &auth.OAuthBearerToken{Token: "abcdef"}

	// Unlike the other tests in this file, we'll set up a couple of credentials
	// right now, and then list from there.
	githubCred, err := fx.db.Create(fx.userCtx, githubScope, token)
	require.NoError(t, err)

	gitlabCred, err := fx.db.Create(fx.userCtx, gitlabScope, token)
	require.NoError(t, err)

	githubAppCred, err := fx.db.Create(fx.user2Ctx, githubScopeWithGitHubApp, auther)
	require.NoError(t, err)

	// This one should always be invisible to the user tests below.
	_, err = fx.db.Create(fx.adminCtx, adminScope, token)
	require.NoError(t, err)

	t.Run("not found", func(t *testing.T) {
		creds, next, err := fx.db.List(fx.userCtx, UserCredentialsListOpts{
			Scope: UserCredentialScope{
				Domain: "this is not a valid domain",
			},
		})
		assert.NoError(t, err)
		assert.Zero(t, next)
		assert.Empty(t, creds)
	})

	t.Run("user accessing admin", func(t *testing.T) {
		creds, next, err := fx.db.List(fx.userCtx, UserCredentialsListOpts{
			Scope: UserCredentialScope{UserID: fx.admin.ID},
		})
		assert.NoError(t, err)
		assert.Zero(t, next)
		assert.Empty(t, creds)
	})

	for name, tc := range map[string]struct {
		scope UserCredentialScope
		want  *UserCredential
	}{
		"service ID only": {
			scope: UserCredentialScope{
				ExternalServiceID: "https://github.com",
			},
			want: githubCred,
		},
		"service type only": {
			scope: UserCredentialScope{
				ExternalServiceType: "gitlab",
			},
			want: gitlabCred,
		},
		"full scope": {
			scope: githubScope,
			want:  githubCred,
		},
	} {
		t.Run("single match on "+name, func(t *testing.T) {
			creds, next, err := fx.db.List(fx.userCtx, UserCredentialsListOpts{
				Scope: tc.scope,
			})
			assert.NoError(t, err)
			assert.Zero(t, next)
			assert.Equal(t, []*UserCredential{tc.want}, creds)
		})
	}

	// Combinations that return all user credentials.
	for name, opts := range map[string]UserCredentialsListOpts{
		"no options":   {},
		"domain only":  {Scope: UserCredentialScope{Domain: UserCredentialDomainBatches}},
		"user ID only": {Scope: UserCredentialScope{UserID: fx.user.ID}},
		"domain and user ID": {
			Scope: UserCredentialScope{
				Domain: UserCredentialDomainBatches,
				UserID: fx.user.ID,
			},
		},
	} {
		t.Run("multiple matches on "+name, func(t *testing.T) {
			creds, next, err := fx.db.List(fx.userCtx, opts)
			assert.NoError(t, err)
			assert.Zero(t, next)
			assert.Equal(t, []*UserCredential{githubCred, gitlabCred}, creds)
		})

		t.Run("pagination for "+name, func(t *testing.T) {
			o := opts
			o.LimitOffset = &LimitOffset{Limit: 1}
			creds, next, err := fx.db.List(fx.userCtx, o)
			assert.NoError(t, err)
			assert.EqualValues(t, 1, next)
			assert.Equal(t, []*UserCredential{githubCred}, creds)

			o.LimitOffset = &LimitOffset{Limit: 1, Offset: next}
			creds, next, err = fx.db.List(fx.userCtx, o)
			assert.NoError(t, err)
			assert.Zero(t, next)
			assert.Equal(t, []*UserCredential{gitlabCred}, creds)
		})
	}

	t.Run("user with github app credential", func(t *testing.T) {
		creds, next, err := fx.db.List(fx.user2Ctx, UserCredentialsListOpts{
			Scope: UserCredentialScope{
				Domain: UserCredentialDomainBatches,
			},
		})
		assert.NoError(t, err)
		assert.Zero(t, next)
		assert.Equal(t, []*UserCredential{githubAppCred}, creds)
	})
}

func TestUserCredentials_Invalid(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	fx := setUpUserCredentialTest(t, db)
	ctx := fx.internalCtx
	key := fx.key

	t.Run("marshal", func(t *testing.T) {
		_, err := fx.db.Create(ctx, UserCredentialScope{}, &invalidAuth{})
		assert.Error(t, err)
	})

	t.Run("unmarshal", func(t *testing.T) {
		// We'll set up some cases here that just shouldn't happen at all, and
		// make sure they bubble up with errors where we expect. Let's define a
		// helper to make that easier.

		insertRawCredential := func(t *testing.T, domain string, raw string) int64 {
			kid := testEncryptionKeyID(key)
			secret, err := key.Encrypt(ctx, []byte(raw))
			require.NoError(t, err)

			q := sqlf.Sprintf(
				userCredentialsCreateQueryFmtstr,
				domain,
				fx.user.ID,
				"type",
				"id",
				secret,
				kid,
				nil,
				sqlf.Sprintf("id"),
			)

			var id int64
			err = db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&id)
			require.NoError(t, err)

			return id
		}

		for name, id := range map[string]int64{
			"invalid credential type": insertRawCredential(t, "invalid", `{"type":"InvalidType"}`),
			"lying credential type":   insertRawCredential(t, "lying", `{"type":"BasicAuth","username":42}`),
			"malformed JSON":          insertRawCredential(t, "malformed", "this is not valid JSON"),
		} {
			t.Run(name, func(t *testing.T) {
				cred, err := fx.db.GetByID(ctx, id)
				require.NoError(t, err)

				_, err = cred.Authenticator(ctx, ghauth.CreateAuthenticatorForCredentialOpts{})
				assert.Error(t, err)
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

func assertUserCredentialNotFoundError(t *testing.T, want any, have error) {
	t.Helper()

	var e UserCredentialNotFoundErr
	assert.ErrorAs(t, have, &e)
	assert.Len(t, e.args, 1)
	assert.EqualValues(t, want, e.args[0])
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
		&auth.BasicAuthWithSSH{BasicAuth: auth.BasicAuth{Username: "foo", Password: "bar"}, PrivateKey: "private", PublicKey: "public", Passphrase: "pass"},
		&auth.OAuthBearerToken{Token: "abcdef"},
		&auth.OAuthBearerTokenWithSSH{OAuthBearerToken: auth.OAuthBearerToken{Token: "abcdef"}, PrivateKey: "private", PublicKey: "public", Passphrase: "pass"},
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

type testFixture struct {
	internalCtx context.Context
	userCtx     context.Context
	user2Ctx    context.Context
	adminCtx    context.Context

	db  UserCredentialsStore
	key encryption.Key

	user  *types.User
	user2 *types.User
	admin *types.User
}

func setUpUserCredentialTest(t *testing.T, db DB) *testFixture {
	if testing.Short() {
		t.Skip()
	}

	t.Helper()
	ctx := context.Background()
	key := et.TestKey{}

	admin, err := db.Users().Create(ctx, NewUser{
		Email:                 "admin@example.com",
		Username:              "admin",
		Password:              "pw",
		EmailVerificationCode: "c",
	})
	require.NoError(t, err)

	user, err := db.Users().Create(ctx, NewUser{
		Email:                 "a@example.com",
		Username:              "u2",
		Password:              "pw",
		EmailVerificationCode: "c",
	})
	require.NoError(t, err)

	user2, err := db.Users().Create(ctx, NewUser{
		Email:                 "b@example.com",
		Username:              "u3",
		Password:              "pw",
		EmailVerificationCode: "c",
	})
	require.NoError(t, err)

	return &testFixture{
		internalCtx: actor.WithInternalActor(ctx),
		userCtx:     actor.WithActor(ctx, actor.FromUser(user.ID)),
		user2Ctx:    actor.WithActor(ctx, actor.FromUser(user2.ID)),
		adminCtx:    actor.WithActor(ctx, actor.FromUser(admin.ID)),
		key:         key,
		db:          db.UserCredentials(key),
		user:        user,
		admin:       admin,
		user2:       user2,
	}
}

type authFailureTestCase struct {
	user  *types.User
	ctx   context.Context
	setup func(*testing.T)
}

func authFailureTestCases(t *testing.T, fx *testFixture) map[string]authFailureTestCase {
	t.Helper()

	return map[string]authFailureTestCase{
		"user accessing admin": {
			user:  fx.admin,
			ctx:   fx.userCtx,
			setup: func(*testing.T) {},
		},
		"admin accessing user without permission": {
			user: fx.user,
			ctx:  fx.adminCtx,
			setup: func(*testing.T) {
				conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
					AuthzEnforceForSiteAdmins: true,
				}})
				t.Cleanup(func() { conf.Mock(nil) })
			},
		},
		"anonymous accessing user": {
			user:  fx.user,
			ctx:   context.Background(),
			setup: func(*testing.T) {},
		},
	}
}

type invalidAuth struct{}

var _ auth.Authenticator = &invalidAuth{}

func (*invalidAuth) Authenticate(_ *http.Request) error { panic("should not be called") }
func (*invalidAuth) Hash() string                       { panic("should not be called") }

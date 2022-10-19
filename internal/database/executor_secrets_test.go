package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

func TestExecutorSecrets_CreateUpdate(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	fx := setUpUserCredentialTest(t, db)

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
				auth := &auth.BasicAuth{}

				// Attempt to create with the invalid context.
				cred, err := fx.db.Create(tc.ctx, scope, auth)
				assert.Error(t, err)
				assert.Nil(t, cred)

				// Now we'll create a credential so we can test update.
				cred, err = fx.db.Create(fx.internalCtx, scope, auth)
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
	for name, auth := range createUserCredentialAuths(t) {
		t.Run(name, func(t *testing.T) {
			scope := UserCredentialScope{
				Domain:              name,
				UserID:              fx.user.ID,
				ExternalServiceType: extsvc.TypeGitHub,
				ExternalServiceID:   "https://github.com",
			}

			cred, err := fx.db.Create(fx.userCtx, scope, auth)
			assert.NoError(t, err)
			assert.NotNil(t, cred)
			assert.NotZero(t, cred.ID)
			assert.Equal(t, scope.Domain, cred.Domain)
			assert.Equal(t, scope.UserID, cred.UserID)
			assert.Equal(t, scope.ExternalServiceType, cred.ExternalServiceType)
			assert.Equal(t, scope.ExternalServiceID, cred.ExternalServiceID)
			assert.NotZero(t, cred.CreatedAt)
			assert.NotZero(t, cred.UpdatedAt)

			have, err := cred.Authenticator(fx.userCtx)
			assert.NoError(t, err)
			assert.Equal(t, auth, have)

			// Ensure that trying to insert again fails.
			second, err := fx.db.Create(fx.userCtx, scope, auth)
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
}

func TestExecutorSecrets_Delete(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
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

func TestExecutorSecrets_GetByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
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

func TestExecutorSecrets_GetByScope(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
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
}

func TestExecutorSecrets_List(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	fx := setUpUserCredentialTest(t, db)

	githubScope := UserCredentialScope{
		Domain:              UserCredentialDomainBatches,
		UserID:              fx.user.ID,
		ExternalServiceType: "github",
		ExternalServiceID:   "https://github.com",
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
}

func TestExecutorSecretNotFoundError(t *testing.T) {
	err := ExecutorSecretNotFoundErr{}
	if have := errcode.IsNotFound(err); !have {
		t.Error("ExecutorSecretNotFoundErr does not say it represents a not found error")
	}
}

func assertExecutorSecretNotFoundError(t *testing.T, want int64, have error) {
	t.Helper()

	var e ExecutorSecretNotFoundErr
	assert.ErrorAs(t, have, &e)
	assert.EqualValues(t, want, e.id)
}

// func createUserCredentialAuths(t *testing.T) map[string]auth.Authenticator {
// 	t.Helper()

// 	createOAuthClient := func(t *testing.T, token, secret string) *oauth.Client {
// 		t.Helper()

// 		// Generate a random key so we can test different clients are different.
// 		// Note that this is wildly insecure.
// 		key, err := rsa.GenerateKey(rand.Reader, 64)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		return &oauth.Client{
// 			Credentials: oauth.Credentials{
// 				Token:  token,
// 				Secret: secret,
// 			},
// 			PrivateKey: key,
// 		}
// 	}

// 	auths := make(map[string]auth.Authenticator)
// 	for _, a := range []auth.Authenticator{
// 		&auth.OAuthClient{Client: createOAuthClient(t, "abc", "def")},
// 		&auth.BasicAuth{Username: "foo", Password: "bar"},
// 		&auth.BasicAuthWithSSH{BasicAuth: auth.BasicAuth{Username: "foo", Password: "bar"}, PrivateKey: "private", PublicKey: "public", Passphrase: "pass"},
// 		&auth.OAuthBearerToken{Token: "abcdef"},
// 		&auth.OAuthBearerTokenWithSSH{OAuthBearerToken: auth.OAuthBearerToken{Token: "abcdef"}, PrivateKey: "private", PublicKey: "public", Passphrase: "pass"},
// 		&bitbucketserver.SudoableOAuthClient{
// 			Client:   auth.OAuthClient{Client: createOAuthClient(t, "ghi", "jkl")},
// 			Username: "neo",
// 		},
// 		&gitlab.SudoableToken{Token: "mnop", Sudo: "qrs"},
// 	} {
// 		auths[reflect.TypeOf(a).String()] = a
// 	}

// 	return auths
// }

// type testFixture struct {
// 	internalCtx context.Context
// 	userCtx     context.Context
// 	adminCtx    context.Context

// 	db  UserCredentialsStore
// 	key encryption.Key

// 	user  *types.User
// 	admin *types.User
// }

// func setUpUserCredentialTest(t *testing.T, db DB) *testFixture {
// 	if testing.Short() {
// 		t.Skip()
// 	}

// 	t.Helper()
// 	ctx := context.Background()
// 	key := et.TestKey{}

// 	admin, err := db.Users().Create(ctx, NewUser{
// 		Email:                 "admin@example.com",
// 		Username:              "admin",
// 		Password:              "pw",
// 		EmailVerificationCode: "c",
// 	})
// 	require.NoError(t, err)

// 	user, err := db.Users().Create(ctx, NewUser{
// 		Email:                 "a@example.com",
// 		Username:              "u2",
// 		Password:              "pw",
// 		EmailVerificationCode: "c",
// 	})
// 	require.NoError(t, err)

// 	return &testFixture{
// 		internalCtx: actor.WithInternalActor(ctx),
// 		userCtx:     actor.WithActor(ctx, actor.FromUser(user.ID)),
// 		adminCtx:    actor.WithActor(ctx, actor.FromUser(admin.ID)),
// 		key:         key,
// 		db:          db.UserCredentials(key),
// 		user:        user,
// 		admin:       admin,
// 	}
// }

// type authFailureTestCase struct {
// 	user  *types.User
// 	ctx   context.Context
// 	setup func(*testing.T)
// }

// func authFailureTestCases(t *testing.T, fx *testFixture) map[string]authFailureTestCase {
// 	t.Helper()

// 	return map[string]authFailureTestCase{
// 		"user accessing admin": {
// 			user:  fx.admin,
// 			ctx:   fx.userCtx,
// 			setup: func(*testing.T) {},
// 		},
// 		"admin accessing user without permission": {
// 			user: fx.user,
// 			ctx:  fx.adminCtx,
// 			setup: func(*testing.T) {
// 				conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
// 					AuthzEnforceForSiteAdmins: true,
// 				}})
// 				t.Cleanup(func() { conf.Mock(nil) })
// 			},
// 		},
// 		"anonymous accessing user": {
// 			user:  fx.user,
// 			ctx:   context.Background(),
// 			setup: func(*testing.T) {},
// 		},
// 	}
// }

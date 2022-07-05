package gitlaboauth

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSessionIssuerHelper_CreateCodeHostConnection(t *testing.T) {
	createCodeHostConnectionHelper(t, false)
}

func TestSessionIssuerHelper_GetOrCreateUser(t *testing.T) {
	glURL, _ := url.Parse("https://gitlab.com")
	codeHost := extsvc.NewCodeHost(glURL, extsvc.TypeGitLab)
	clientID := "client-id"

	authSaveableUsers := map[string]int32{
		"alice": 1,
		"cindy": 3,
		"dan":   4,
	}

	signupNotAllowed := new(bool)
	signupAllowed := new(bool)
	*signupAllowed = true

	type input struct {
		description     string
		glUser          *gitlab.User
		glUserGroups    []*gitlab.Group
		glUserGroupsErr error
		allowSignup     *bool
		allowGroups     []string
	}

	cases := []struct {
		inputs        []input
		expActor      *actor.Actor
		expErr        bool
		expAuthUserOp *auth.GetAndSaveUserOp
	}{
		{
			inputs: []input{{
				description: "glUser, allowSignup not set, defaults to true -> new user and session created",
				glUser: &gitlab.User{
					ID:       int32(104),
					Username: string("dan"),
					Email:    string("dan@example.com"),
				},
			}},
			expActor: &actor.Actor{UID: 4},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps: database.NewUser{
					Username:        "dan",
					Email:           "dan@example.com",
					EmailIsVerified: true,
				},
				ExternalAccount: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitLab,
					ServiceID:   "https://gitlab.com/",
					ClientID:    clientID,
					AccountID:   "104",
				},
				CreateIfNotExist: true,
			},
		},
		{
			inputs: []input{{
				description: "glUser, allowSignup set to false -> no new user nor session created",
				allowSignup: signupNotAllowed,
				glUser: &gitlab.User{
					ID:       int32(102),
					Username: string("bob"),
					Email:    string("bob@example.com"),
				},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "glUser, allowSignup set to false, allowedGroups list provided -> no new user nor session created",
				allowSignup: signupNotAllowed,
				allowGroups: []string{"group1"},
				glUser: &gitlab.User{
					ID:       int32(102),
					Username: string("bob"),
					Email:    string("bob@example.com"),
				},
				glUserGroups: []*gitlab.Group{
					{FullPath: "group1"},
				},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "glUser, allowSignup set true -> new user and session created",
				allowSignup: signupAllowed,
				glUser: &gitlab.User{
					ID:       int32(103),
					Username: string("cindy"),
					Email:    string("cindy@example.com"),
				},
			}},
			expActor: &actor.Actor{UID: 3},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps: database.NewUser{
					Username:        "cindy",
					Email:           "cindy@example.com",
					EmailIsVerified: true,
				},
				ExternalAccount: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitLab,
					ServiceID:   "https://gitlab.com/",
					ClientID:    clientID,
					AccountID:   "101",
				},
				CreateIfNotExist: true,
			},
		},
		{
			inputs: []input{{
				description: "glUser, allowedGroups not set -> session created",
				glUser: &gitlab.User{
					ID:       int32(101),
					Username: string("alice"),
					Email:    string("alice@example.com"),
				},
			}},
			expActor: &actor.Actor{UID: 1},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps: database.NewUser{
					Username:        "alice",
					Email:           "alice@example.com",
					EmailIsVerified: true,
				},
				ExternalAccount: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitLab,
					ServiceID:   "https://gitlab.com/",
					ClientID:    clientID,
					AccountID:   "101",
				},
				CreateIfNotExist: true,
			},
		},
		{
			inputs: []input{{
				description: "glUser, not in allowed groups -> no session created",
				allowGroups: []string{"group2"},
				glUser: &gitlab.User{
					ID:       int32(101),
					Username: string("alice"),
					Email:    string("alice@example.com"),
				},
				glUserGroups: []*gitlab.Group{
					{FullPath: "group1"},
				},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "glUser, in allowed groups, error getting user groups -> no session created",
				allowGroups: []string{"group1"},
				glUser: &gitlab.User{
					ID:       int32(101),
					Username: string("alice"),
					Email:    string("alice@example.com"),
				},
				glUserGroups: []*gitlab.Group{
					{FullPath: "group1"},
					{FullPath: "group2"},
				},
				glUserGroupsErr: errors.New("boom"),
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "glUser, in allowed groups -> session created",
				allowGroups: []string{"group1"},
				glUser: &gitlab.User{
					ID:       int32(101),
					Username: string("alice"),
					Email:    string("alice@example.com"),
				},
				glUserGroups: []*gitlab.Group{
					{FullPath: "group1"},
					{FullPath: "group2"},
				},
			}},
			expActor: &actor.Actor{UID: 1},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps: database.NewUser{
					Username:        "alice",
					Email:           "alice@example.com",
					EmailIsVerified: true,
				},
				ExternalAccount: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitLab,
					ServiceID:   "https://gitlab.com/",
					ClientID:    clientID,
					AccountID:   "101",
				},
				CreateIfNotExist: true,
			},
		},
		{
			inputs: []input{{
				description: "glUser, not in allowed subgroup -> session not created",
				allowGroups: []string{"group1/subgroup1"},
				glUser: &gitlab.User{
					ID:       int32(101),
					Username: string("alice"),
					Email:    string("alice@example.com"),
				},
				glUserGroups: []*gitlab.Group{
					{FullPath: "group1/subgroup2"},
				},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "glUser, in allowed subgroup  -> session created",
				allowGroups: []string{"group1/subgroup2"},
				glUser: &gitlab.User{
					ID:       int32(101),
					Username: string("alice"),
					Email:    string("alice@example.com"),
				},
				glUserGroups: []*gitlab.Group{
					{FullPath: "group1/subgroup2"},
				},
			}},
			expActor: &actor.Actor{UID: 1},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps: database.NewUser{
					Username:        "alice",
					Email:           "alice@example.com",
					EmailIsVerified: true,
				},
				ExternalAccount: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitLab,
					ServiceID:   "https://gitlab.com/",
					ClientID:    clientID,
					AccountID:   "101",
				},
				CreateIfNotExist: true,
			},
		},
	}

	for _, c := range cases {
		for _, ci := range c.inputs {
			c, ci := c, ci

			t.Run(ci.description, func(t *testing.T) {

				gitlab.MockListGroups = func(ctx context.Context, page int) (groups []*gitlab.Group, hasNextPage bool, err error) {
					return ci.glUserGroups, false, ci.glUserGroupsErr
				}

				var gotAuthUserOp *auth.GetAndSaveUserOp
				getAndSaveUserError := errors.New("auth.GetAndSaveUser error")

				auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (userID int32, safeErrMsg string, err error) {
					if gotAuthUserOp != nil {
						t.Fatal("GetAndSaveUser called more than once")
					}

					op.ExternalAccountData = extsvc.AccountData{}
					gotAuthUserOp = &op

					if uid, ok := authSaveableUsers[op.UserProps.Username]; ok {
						return uid, "", nil
					}

					return 0, "safeErr", getAndSaveUserError
				}

				defer func() {
					auth.MockGetAndSaveUser = nil
					gitlab.MockListGroups = nil
				}()

				ctx := WithUser(context.Background(), ci.glUser)
				s := &sessionIssuerHelper{
					CodeHost:    codeHost,
					clientID:    clientID,
					allowSignup: ci.allowSignup,
					allowGroups: ci.allowGroups,
				}

				tok := &oauth2.Token{AccessToken: "dummy-value-that-isnt-relevant-to-unit-correctness"}
				actr, _, err := s.GetOrCreateUser(ctx, tok, "", "", "")

				if got, exp := actr, c.expActor; !reflect.DeepEqual(got, exp) {
					t.Errorf("expected actor %v, got %v", exp, got)
				}

				if c.expErr && err == nil {
					t.Errorf("expected err %v, but was nil", c.expErr)
				} else if !c.expErr && err != nil {
					t.Errorf("expected no error, but was %v", err)
				}

				if c.expErr && err != getAndSaveUserError {
					if got, exp := gotAuthUserOp, c.expAuthUserOp; !reflect.DeepEqual(got, exp) {
						t.Error(cmp.Diff(got, exp))
					}
				}
			})
		}
	}

}

func TestSessionIssuerHelper_CreateCodeHostConnectionHandlesExistingService(t *testing.T) {
	createCodeHostConnectionHelper(t, true)
}

func createCodeHostConnectionHelper(t *testing.T, serviceExists bool) {
	t.Helper()

	ctx := context.Background()
	db := database.NewMockDB()
	s := &sessionIssuerHelper{db: db}
	t.Run("Unauthenticated request", func(t *testing.T) {
		_, _, err := s.CreateCodeHostConnection(ctx, nil, "")
		assert.Error(t, err)
	})

	mockGitLabCom := newMockProvider(t, db, "gitlabcomclient", "gitlabcomsecret", "https://gitlab.com/")
	providers.MockProviders = []providers.Provider{mockGitLabCom.Provider}
	defer func() { providers.MockProviders = nil }()

	expiry := time.Now().Add(1 * time.Hour)
	tok := &oauth2.Token{
		AccessToken:  "dummy-value-that-isnt-relevant-to-unit-correctness",
		RefreshToken: "some-refresh-token",
		Expiry:       expiry,
	}
	act := &actor.Actor{UID: 1}
	glUser := &gitlab.User{
		ID:       101,
		Username: "alice",
	}

	ctx = actor.WithActor(ctx, act)
	ctx = WithUser(ctx, glUser)
	now := time.Now()

	externalServices := database.NewMockExternalServiceStore()
	externalServices.TransactFunc.SetDefaultReturn(externalServices, nil)
	externalServices.ListFunc.SetDefaultHook(func(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		if !serviceExists {
			return nil, nil
		}
		return []*types.ExternalService{
			{
				Kind:        extsvc.KindGitLab,
				DisplayName: fmt.Sprintf("GitLab (%s)", glUser.Username),
				Config: fmt.Sprintf(`
{
  "url": "%s",
  "token": "%s",
  "token.type": "oauth",
  "token.oauth.refresh": "%s",
  "token.oauth.expiry": %d,
  "projectQuery": ["projects?id_before=0"]
}
`, mockGitLabCom.ServiceID, "a-token-that-should-be-replaced", "old-refresh-token", expiry.Add(-1*time.Hour).Unix()),

				NamespaceUserID: act.UID,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		}, nil
	})
	var got *types.ExternalService
	externalServices.CreateFunc.SetDefaultHook(func(ctx context.Context, confGet func() *conf.Unified, es *types.ExternalService) error {
		got = es
		return nil
	})
	externalServices.UpsertFunc.SetDefaultHook(func(ctx context.Context, svcs ...*types.ExternalService) error {
		require.Len(t, svcs, 1)

		// Tweak timestamps
		svcs[0].CreatedAt = now
		svcs[0].UpdatedAt = now
		got = svcs[0]
		return nil
	})
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	fromCreation, _, err := s.CreateCodeHostConnection(ctx, tok, mockGitLabCom.ConfigID().ID)
	require.NoError(t, err)

	want := &types.ExternalService{
		Kind:        extsvc.KindGitLab,
		DisplayName: fmt.Sprintf("GitLab (%s)", glUser.Username),
		Config: fmt.Sprintf(`
{
  "url": "%s",
  "token": "%s",
  "token.type": "oauth",
  "token.oauth.refresh": "%s",
  "token.oauth.expiry": %d,
  "projectQuery": ["projects?id_before=0"]
}
`, mockGitLabCom.ServiceID, tok.AccessToken, tok.RefreshToken, tok.Expiry.Unix()),
		NamespaceUserID: act.UID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	assert.Equal(t, want, got)
	assert.Equal(t, want, fromCreation)
}

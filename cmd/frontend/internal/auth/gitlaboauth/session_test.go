package gitlaboauth

import (
	"context"
	"net/url"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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
		glUser          *gitlab.AuthUser
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
				glUser: &gitlab.AuthUser{
					ID:       int32(104),
					Username: "dan",
					Email:    "dan@example.com",
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
				glUser: &gitlab.AuthUser{
					ID:       int32(102),
					Username: "bob",
					Email:    "bob@example.com",
				},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "glUser, allowSignup set to false, allowedGroups list provided -> no new user nor session created",
				allowSignup: signupNotAllowed,
				allowGroups: []string{"group1"},
				glUser: &gitlab.AuthUser{
					ID:       int32(102),
					Username: "bob",
					Email:    "bob@example.com",
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
				glUser: &gitlab.AuthUser{
					ID:       int32(103),
					Username: "cindy",
					Email:    "cindy@example.com",
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
				glUser: &gitlab.AuthUser{
					ID:       int32(101),
					Username: "alice",
					Email:    "alice@example.com",
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
				glUser: &gitlab.AuthUser{
					ID:       int32(101),
					Username: "alice",
					Email:    "alice@example.com",
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
				glUser: &gitlab.AuthUser{
					ID:       int32(101),
					Username: "alice",
					Email:    "alice@example.com",
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
				glUser: &gitlab.AuthUser{
					ID:       int32(101),
					Username: "alice",
					Email:    "alice@example.com",
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
				glUser: &gitlab.AuthUser{
					ID:       int32(101),
					Username: "alice",
					Email:    "alice@example.com",
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
				glUser: &gitlab.AuthUser{
					ID:       int32(101),
					Username: "alice",
					Email:    "alice@example.com",
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

				auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
					if gotAuthUserOp != nil {
						t.Fatal("GetAndSaveUser called more than once")
					}

					op.ExternalAccountData = extsvc.AccountData{}
					gotAuthUserOp = &op

					if uid, ok := authSaveableUsers[op.UserProps.Username]; ok {
						return false, uid, "", nil
					}

					return false, 0, "safeErr", getAndSaveUserError
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
				_, actr, _, err := s.GetOrCreateUser(ctx, tok, nil)

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

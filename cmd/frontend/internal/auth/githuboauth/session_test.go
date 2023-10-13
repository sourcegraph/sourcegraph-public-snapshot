package githuboauth

import (
	"context"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	githublogin "github.com/dghubble/gologin/v2/github"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	githubsvc "github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github/githubconvert"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	spew.Config.DisablePointerAddresses = true
	spew.Config.SortKeys = true
	spew.Config.SpewKeys = true
}

func TestSessionIssuerHelper_GetOrCreateUser(t *testing.T) {
	ghURL, _ := url.Parse("https://github.com")
	codeHost := extsvc.NewCodeHost(ghURL, extsvc.TypeGitHub)
	clientID := "client-id"

	// Top-level mock data
	//
	// authSaveableUsers that will be accepted by auth.GetAndSaveUser
	authSaveableUsers := map[string]int32{
		"alice": 1,
	}

	type input struct {
		description     string
		ghUser          *github.User
		ghUserEmails    []*githubsvc.UserEmail
		ghUserOrgs      []*githubsvc.Org
		ghUserTeams     []*githubsvc.Team
		ghUserEmailsErr error
		ghUserOrgsErr   error
		ghUserTeamsErr  error
		allowSignup     bool
		allowOrgs       []string
		allowOrgsMap    map[string][]string
	}
	cases := []struct {
		inputs        []input
		expActor      *actor.Actor
		expErr        bool
		expAuthUserOp *auth.GetAndSaveUserOp
	}{
		{
			inputs: []input{{
				description: "ghUser, verified email -> session created",
				ghUser:      &github.User{ID: github.Int64(101), Login: github.String("alice")},
				ghUserEmails: []*githubsvc.UserEmail{{
					Email:    "alice@example.com",
					Primary:  true,
					Verified: true,
				}},
			}},
			expActor: &actor.Actor{UID: 1},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps:       u("alice", "alice@example.com", true),
				ExternalAccount: acct(extsvc.TypeGitHub, "https://github.com/", clientID, "101"),
			},
		},
		{
			inputs: []input{{
				description: "ghUser, primary email not verified but another is -> no session created",
				ghUser:      &github.User{ID: github.Int64(101), Login: github.String("alice")},
				ghUserEmails: []*githubsvc.UserEmail{{
					Email:    "alice@example1.com",
					Primary:  true,
					Verified: false,
				}, {
					Email:    "alice@example2.com",
					Primary:  false,
					Verified: false,
				}, {
					Email:    "alice@example3.com",
					Primary:  false,
					Verified: true,
				}},
			}},
			expActor: &actor.Actor{UID: 1},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps:       u("alice", "alice@example3.com", true),
				ExternalAccount: acct(extsvc.TypeGitHub, "https://github.com/", clientID, "101"),
			},
		},
		{
			inputs: []input{{
				description: "ghUser, no emails -> no session created",
				ghUser:      &github.User{ID: github.Int64(101), Login: github.String("alice")},
			}, {
				description:     "ghUser, email fetching err -> no session created",
				ghUser:          &github.User{ID: github.Int64(101), Login: github.String("alice")},
				ghUserEmailsErr: errors.New("x"),
			}, {
				description: "ghUser, plenty of emails but none verified -> no session created",
				ghUser:      &github.User{ID: github.Int64(101), Login: github.String("alice")},
				ghUserEmails: []*githubsvc.UserEmail{{
					Email:    "alice@example1.com",
					Primary:  true,
					Verified: false,
				}, {
					Email:    "alice@example2.com",
					Primary:  false,
					Verified: false,
				}, {
					Email:    "alice@example3.com",
					Primary:  false,
					Verified: false,
				}},
			}, {
				description: "no ghUser -> no session created",
			}, {
				description: "ghUser, verified email, unsaveable -> no session created",
				ghUser:      &github.User{ID: github.Int64(102), Login: github.String("bob")},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "ghUser, verified email, not in allowed orgs -> no session created",
				allowOrgs:   []string{"sourcegraph"},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("alice"),
				},
				ghUserEmails: []*githubsvc.UserEmail{{
					Email:    "alice@example.com",
					Primary:  true,
					Verified: true,
				}},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "ghUser, verified email, error getting user orgs -> no session created",
				allowOrgs:   []string{"sourcegraph"},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("alice"),
				},
				ghUserEmails: []*githubsvc.UserEmail{{
					Email:    "alice@example.com",
					Primary:  true,
					Verified: true,
				}},
				ghUserOrgs: []*githubsvc.Org{
					{Login: "sourcegraph"},
					{Login: "example"},
				},
				ghUserOrgsErr: errors.New("boom"),
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description: "ghUser, verified email, allowed orgs -> session created",
				allowOrgs:   []string{"sourcegraph"},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("alice"),
				},
				ghUserEmails: []*githubsvc.UserEmail{{
					Email:    "alice@example.com",
					Primary:  true,
					Verified: true,
				}},
				ghUserOrgs: []*githubsvc.Org{
					{Login: "sourcegraph"},
					{Login: "example"},
				},
			}},
			expActor: &actor.Actor{UID: 1},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps:       u("alice", "alice@example.com", true),
				ExternalAccount: acct(extsvc.TypeGitHub, "https://github.com/", clientID, "101"),
			},
		},
		{
			inputs: []input{{
				description:  "ghUser, verified email, team name matches, org name doesn't match -> no session created",
				allowOrgsMap: map[string][]string{"org1": {"team1"}},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("alice"),
				},
				ghUserEmails: []*githubsvc.UserEmail{{
					Email:    "alice@example.com",
					Primary:  true,
					Verified: true,
				}},
				ghUserTeams: []*githubsvc.Team{
					{Name: "team1", Organization: &githubsvc.Org{Login: "org2"}},
				},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description:  "ghUser, verified email, team name doesn't match, org name matches -> no session created",
				allowOrgsMap: map[string][]string{"org1": {"team1"}},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("alice"),
				},
				ghUserEmails: []*githubsvc.UserEmail{{
					Email:    "alice@example.com",
					Primary:  true,
					Verified: true,
				}},
				ghUserTeams: []*githubsvc.Team{
					{Name: "team2", Organization: &githubsvc.Org{Login: "org1"}},
				},
			}},
			expErr: true,
		},
		{
			inputs: []input{{
				description:  "ghUser, verified email, in allowed org > teams -> session created",
				allowOrgsMap: map[string][]string{"org1": {"team1"}},
				ghUser: &github.User{
					ID:    github.Int64(101),
					Login: github.String("alice"),
				},
				ghUserEmails: []*githubsvc.UserEmail{{
					Email:    "alice@example.com",
					Primary:  true,
					Verified: true,
				}},
				ghUserTeams: []*githubsvc.Team{
					{Name: "team1", Organization: &githubsvc.Org{Login: "org1"}},
				},
			}},
			expActor: &actor.Actor{UID: 1},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps:       u("alice", "alice@example.com", true),
				ExternalAccount: acct(extsvc.TypeGitHub, "https://github.com/", clientID, "101"),
			},
		},
	}
	for _, c := range cases {
		for _, ci := range c.inputs {
			c, ci := c, ci
			t.Run(ci.description, func(t *testing.T) {
				githubsvc.MockGetAuthenticatedUserEmails = func(ctx context.Context) ([]*githubsvc.UserEmail, error) {
					return ci.ghUserEmails, ci.ghUserEmailsErr
				}
				githubsvc.MockGetAuthenticatedUserOrgs.FnMock = func(ctx context.Context) ([]*githubsvc.Org, bool, int, error) {
					return ci.ghUserOrgs, false, 1, ci.ghUserOrgsErr
				}
				githubsvc.MockGetAuthenticatedUserTeams = func(ctx context.Context, page int) ([]*githubsvc.Team, bool, int, error) {
					return ci.ghUserTeams, false, 0, ci.ghUserTeamsErr
				}
				var gotAuthUserOp *auth.GetAndSaveUserOp
				auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
					if gotAuthUserOp != nil {
						t.Fatal("GetAndSaveUser called more than once")
					}
					op.ExternalAccountData = extsvc.AccountData{} // ignore AccountData value
					gotAuthUserOp = &op

					if uid, ok := authSaveableUsers[op.UserProps.Username]; ok {
						return false, uid, "", nil
					}
					return false, 0, "safeErr", errors.New("auth.GetAndSaveUser error")
				}
				defer func() {
					auth.MockGetAndSaveUser = nil
					githubsvc.MockGetAuthenticatedUserEmails = nil
					githubsvc.MockGetAuthenticatedUserTeams = nil
					githubsvc.MockGetAuthenticatedUserOrgs.FnMock = nil
				}()

				ctx := githublogin.WithUser(context.Background(), githubconvert.ConvertUserV55ToV48(ci.ghUser))
				s := &sessionIssuerHelper{
					CodeHost:     codeHost,
					clientID:     clientID,
					allowSignup:  ci.allowSignup,
					allowOrgs:    ci.allowOrgs,
					allowOrgsMap: ci.allowOrgsMap,
				}

				tok := &oauth2.Token{AccessToken: "dummy-value-that-isnt-relevant-to-unit-correctness"}
				_, actr, _, err := s.GetOrCreateUser(ctx, tok, "", "", "")
				if got, exp := actr, c.expActor; !reflect.DeepEqual(got, exp) {
					t.Errorf("expected actor %v, got %v", exp, got)
				}

				if c.expErr && err == nil {
					t.Errorf("expected err %v, but was nil", c.expErr)
				} else if !c.expErr && err != nil {
					t.Errorf("expected no error, but was %v", err)
				}
				if got, exp := gotAuthUserOp, c.expAuthUserOp; !reflect.DeepEqual(got, exp) {
					t.Error(cmp.Diff(got, exp))
				}
			})
		}
	}
}

func TestSessionIssuerHelper_SignupMatchesSecondaryAccount(t *testing.T) {
	githubsvc.MockGetAuthenticatedUserEmails = func(ctx context.Context) ([]*githubsvc.UserEmail, error) {
		return []*githubsvc.UserEmail{
			{
				Email:    "primary@example.com",
				Primary:  true,
				Verified: true,
			},
			{
				Email:    "secondary@example.com",
				Primary:  false,
				Verified: true,
			},
		}, nil
	}
	// We just want to make sure that we end up getting to the secondary email
	auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
		if op.CreateIfNotExist {
			// We should not get here as we should hit the second email address
			// before trying again with creation enabled.
			t.Fatal("Should not get here")
		}
		// Mock the second email address matching
		if op.UserProps.Email == "secondary@example.com" {
			return false, 1, "", nil
		}
		return false, 0, "no match", errors.New("no match")
	}
	defer func() {
		githubsvc.MockGetAuthenticatedUserEmails = nil
		auth.MockGetAndSaveUser = nil
	}()

	ghURL, _ := url.Parse("https://github.com")
	codeHost := extsvc.NewCodeHost(ghURL, extsvc.TypeGitHub)
	clientID := "client-id"
	ghUser := &github.User{
		ID:    github.Int64(101),
		Login: github.String("alice"),
	}

	ctx := githublogin.WithUser(context.Background(), githubconvert.ConvertUserV55ToV48(ghUser))
	s := &sessionIssuerHelper{
		CodeHost:    codeHost,
		clientID:    clientID,
		allowSignup: true,
		allowOrgs:   nil,
	}
	tok := &oauth2.Token{AccessToken: "dummy-value-that-isnt-relevant-to-unit-correctness"}
	_, _, _, err := s.GetOrCreateUser(ctx, tok, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSessionIssuerHelper_SignupFailsWithLastError(t *testing.T) {
	githubsvc.MockGetAuthenticatedUserEmails = func(ctx context.Context) ([]*githubsvc.UserEmail, error) {
		return []*githubsvc.UserEmail{
			{
				Email:    "primary@example.com",
				Primary:  true,
				Verified: true,
			},
			{
				Email:    "secondary@example.com",
				Primary:  false,
				Verified: true,
			},
		}, nil
	}
	errorMessage := "could not create new user account, license limit has been reached"

	// We just want to make sure that we end up getting to the signup part
	auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
		if op.CreateIfNotExist {
			// We should not get here as we should hit the second email address
			// before trying again with creation enabled.
			return false, 0, errorMessage, errors.New(errorMessage)
		}
		return false, 0, "no match", errors.New("no match")
	}
	defer func() {
		githubsvc.MockGetAuthenticatedUserEmails = nil
		auth.MockGetAndSaveUser = nil
	}()

	ghURL, _ := url.Parse("https://github.com")
	codeHost := extsvc.NewCodeHost(ghURL, extsvc.TypeGitHub)
	clientID := "client-id"
	ghUser := &github.User{
		ID:    github.Int64(101),
		Login: github.String("alice"),
	}

	ctx := githublogin.WithUser(context.Background(), githubconvert.ConvertUserV55ToV48(ghUser))
	s := &sessionIssuerHelper{
		CodeHost:    codeHost,
		clientID:    clientID,
		allowSignup: true,
		allowOrgs:   nil,
	}
	tok := &oauth2.Token{AccessToken: "dummy-value-that-isnt-relevant-to-unit-correctness"}
	_, _, _, err := s.GetOrCreateUser(ctx, tok, "", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != errorMessage {
		t.Fatalf("expected error message to be %s, got %s", errorMessage, err.Error())
	}
}

func TestVerifyUserOrgs_UserHasMoreThan100Orgs(t *testing.T) {
	// mock calls to get user orgs
	githubsvc.MockGetAuthenticatedUserOrgs.PagesMock = make(map[int][]*githubsvc.Org, 2)
	githubsvc.MockGetAuthenticatedUserOrgs.PagesMock[1] = generate100Orgs(1)
	githubsvc.MockGetAuthenticatedUserOrgs.PagesMock[2] = generate100Orgs(101)

	defer func() {
		githubsvc.MockGetAuthenticatedUserOrgs.PagesMock = nil
	}()

	s := &sessionIssuerHelper{
		CodeHost:     nil,
		clientID:     "clientID",
		allowSignup:  true,
		allowOrgs:    []string{"1337"},
		allowOrgsMap: nil,
	}

	allowed := s.verifyUserOrgs(context.Background(), nil)
	if allowed {
		t.Fatal("User doesn't have an org he is allowed into, but verifyUserOrgs returned true")
	}

	s.allowOrgs = append(s.allowOrgs, "123")

	allowed = s.verifyUserOrgs(context.Background(), nil)
	if !allowed {
		t.Fatal("User has an org he is allowed into, but verifyUserOrgs returned false")
	}
}

func generate100Orgs(startIdx int) (orgs []*githubsvc.Org) {
	for i := startIdx; i < startIdx+100; i++ {
		orgs = append(orgs, &githubsvc.Org{Login: strconv.Itoa(i)})
	}
	return
}

func u(username, email string, emailIsVerified bool) database.NewUser {
	return database.NewUser{
		Username:        username,
		Email:           email,
		EmailIsVerified: emailIsVerified,
	}
}

func acct(serviceType, serviceID, clientID, accountID string) extsvc.AccountSpec {
	return extsvc.AccountSpec{
		ServiceType: serviceType,
		ServiceID:   serviceID,
		ClientID:    clientID,
		AccountID:   accountID,
	}
}

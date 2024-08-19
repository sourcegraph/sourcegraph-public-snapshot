package bitbucketserveroauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	returnUsername  string
	returnAccountID int
	returnEmail     string
)

func createTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/rest/api/1.0/users/") {
			json.NewEncoder(w).Encode(struct {
				Name         string `json:"name"`
				ID           int    `json:"id"`
				EmailAddress string `json:"emailAddress"`
			}{
				Name:         returnUsername,
				ID:           returnAccountID,
				EmailAddress: returnEmail,
			})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/rest/api/1.0/users") {
			w.Header().Add("X-Ausername", returnUsername)
			return
		}
	}))
}

func TestSessionIssuerHelper_GetOrCreateUser(t *testing.T) {
	ratelimit.SetupForTest(t)

	server := createTestServer()
	defer server.Close()
	bbURL, _ := url.Parse(server.URL)
	clientID := "client-id"

	// Top-level mock data
	//
	// authSaveableUsers that will be accepted by auth.GetAndSaveUser
	authSaveableUsers := map[string]int32{
		"alice": 1,
	}

	type input struct {
		description     string
		bbUser          *bitbucketserver.User
		bbUserEmail     string
		bbUserEmailsErr error
		allowSignup     bool
	}
	cases := []struct {
		inputs        []input
		expActor      *actor.Actor
		expErr        bool
		expAuthUserOp *auth.GetAndSaveUserOp
	}{
		{
			inputs: []input{{
				description: "bbUser, verified email -> session created",
				bbUser:      &bitbucketserver.User{Name: "alice"},
				bbUserEmail: "alice@example.com",
			}},
			expActor: &actor.Actor{UID: 1},
			expAuthUserOp: &auth.GetAndSaveUserOp{
				UserProps:       u("alice", "alice@example.com", true),
				ExternalAccount: acct(extsvc.TypeBitbucketServer, server.URL+"/", clientID, "1234"),
			},
		},
	}
	for _, c := range cases {
		for _, ci := range c.inputs {
			c, ci := c, ci
			t.Run(ci.description, func(t *testing.T) {
				if ci.bbUser != nil {
					returnUsername = ci.bbUser.Name
					returnAccountID = 1234
				}
				returnEmail = ci.bbUserEmail

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
				}()

				ctx := context.Background()
				conf := &schema.BitbucketServerConnection{
					Url: server.URL,
				}
				bbClient, err := bitbucketserver.NewClient(server.URL, conf, httpcli.TestExternalDoer)
				if err != nil {
					t.Fatal(err)
				}
				s := &sessionIssuerHelper{
					logger:      logtest.Scoped(t),
					baseURL:     extsvc.NormalizeBaseURL(bbURL),
					clientKey:   clientID,
					allowSignup: ci.allowSignup,
					client:      bbClient,
				}

				tok := &oauth2.Token{AccessToken: "dummy-value-that-isnt-relevant-to-unit-correctness"}
				_, actr, _, err := s.GetOrCreateUser(ctx, tok, nil)
				if c.expErr && err == nil {
					t.Errorf("expected err %v, but was nil", c.expErr)
				} else if !c.expErr && err != nil {
					t.Errorf("expected no error, but was %v", err)
				}

				if got, exp := actr, c.expActor; !reflect.DeepEqual(got, exp) {
					t.Errorf("expected actor %v, got %v", exp, got)
				}

				if got, exp := gotAuthUserOp, c.expAuthUserOp; !reflect.DeepEqual(got, exp) {
					t.Error(cmp.Diff(got, exp))
				}
			})
		}
	}
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

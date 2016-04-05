// +build exectest

package httpapi_test

import (
	"errors"
	"net/http"
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
)

func TestAuth(t *testing.T) {
	httptestutil.ResetGlobals()

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	user, err := a.Client.Accounts.Create(ctx, &sourcegraph.NewAccount{Login: "u", Email: "u@example.com", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}

	httpClient := &httptestutil.Client{Client: http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("redirect")
		},
	}}

	url := router.Rel.URLTo(router.UserSettingsProfile, "User", "u")
	url = a.Config.Endpoint.URLOrDefault().ResolveReference(url)

	// No auth for an endpoint that requires auth.
	_, err = httpClient.Get(url.String())
	if err == nil {
		t.Error("error expected")
	}

	// OAuth2 tests.
	oauth2Client := func(tok *oauth2.Token) *httptestutil.Client {
		oauth2Client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(tok))
		return &httptestutil.Client{Client: *oauth2Client}
	}

	// Try successful OAuth2 access token auth.
	k := idkey.FromContext(ctx)
	tok, err := accesstoken.New(k, auth.Actor{UID: int(user.UID), ClientID: k.ID}, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := oauth2Client(tok).GetOK(url.String()); err != nil {
		t.Fatal(err)
	}

	// Unsuccessful OAuth2 access token auth.
	badTok := &oauth2.Token{AccessToken: "badtoken", TokenType: "Bearer"}
	_, err = oauth2Client(badTok).Get(url.String())
	if err == nil {
		t.Error("error expected")
	}
}

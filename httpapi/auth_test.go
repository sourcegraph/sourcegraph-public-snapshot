// +build exectest

package httpapi_test

import (
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/fed"
	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
	"sourcegraph.com/sqs/pbtypes"
)

func TestAuth(t *testing.T) {
	httptestutil.ResetGlobals()

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "local"},
		&fed.Flags{IsRoot: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	user, err := a.Client.Accounts.Create(ctx, &sourcegraph.NewAccount{Login: "u", Email: "u@example.com", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}

	httpClient := &httptestutil.Client{Client: *http.DefaultClient}

	// Get server's HTTP endpoint URL.
	serverConfig, err := a.Client.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		t.Fatal(err)
	}
	httpURL, err := url.Parse(serverConfig.HTTPEndpoint)
	if err != nil {
		t.Fatal(err)
	}

	url := httpURL.ResolveReference(&url.URL{Path: "repos"})

	// No auth for an endpoint that requires auth.
	resp, err := httpClient.Get(url.String())
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusUnauthorized; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}

	// Try successful password auth.
	req, _ := http.NewRequest("GET", url.String(), nil)
	req.SetBasicAuth("u", "p")
	if _, err := httpClient.DoOK(req); err != nil {
		t.Fatal(err)
	}

	// Unsuccessful password auth.
	req, _ = http.NewRequest("GET", url.String(), nil)
	req.SetBasicAuth("u", "badpw")
	resp, err = httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusForbidden; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}

	// Now OAuth2 tests.
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
	resp, err = oauth2Client(badTok).Get(url.String())
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusUnauthorized; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}
}

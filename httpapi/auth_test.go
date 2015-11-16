// +build exectest

package httpapi_test

import (
	"net/http"
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/accesstoken"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/fed"
	apirouter "src.sourcegraph.com/sourcegraph/httpapi/router"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/httptestutil"
)

func TestAuth(t *testing.T) {
	httptestutil.ResetGlobals()

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "local", AllowAllLogins: true},
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

	url, err := apirouter.URL(apirouter.Repos, nil)
	if err != nil {
		t.Fatal(err)
	}
	url.Path = "/.api" + url.Path
	url = a.Config.Endpoint.URLOrDefault().ResolveReference(url)

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

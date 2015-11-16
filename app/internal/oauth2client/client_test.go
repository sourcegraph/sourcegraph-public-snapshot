package oauth2client

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-querystring/query"
	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/client/pkg/oauth2client"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/pkg/oauth2util"
)

func TestOAuth2_detectAppURLMismatch(t *testing.T) {
	authutil.ActiveFlags.Source = "oauth"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, mock := apptest.New()

	mock.Ctx = oauth2client.WithClientID(mock.Ctx, "a")

	var calledMetaConfig bool
	mock.Meta.Config_ = func(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.ServerConfig, error) {
		calledMetaConfig = true
		return &sourcegraph.ServerConfig{AppURL: "http://mismatch." + conf.AppURL(ctx).Host}, nil
	}

	// Use mismatched request host.
	u := conf.AppURL(mock.Ctx).ResolveReference(router.Rel.URLTo(router.OAuth2ClientInitiate))
	resp, err := c.GetNoFollowRedirects(u.String())
	if err != nil {
		t.Fatalf("GetNoFollowRedirects: %s", err)
	}

	if want := http.StatusConflict; resp.StatusCode != want {
		t.Errorf("got status %d, want %d", resp.StatusCode, want)
	}

	if !calledMetaConfig {
		t.Error("!calledMetaConfig")
	}
}

func TestOAuth2_detectRedirectURIMismatch(t *testing.T) {
	authutil.ActiveFlags.Source = "oauth"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()
	origFedRoot := fed.Config.RootURLStr
	fed.Config.RootURLStr = "http://oauth.example.com"
	defer func() {
		fed.Config.RootURLStr = origFedRoot
	}()

	c, mock := apptest.New()

	mock.Ctx = oauth2client.WithClientID(mock.Ctx, "a")

	var calledMetaConfig, calledGetRegisteredClientForServer bool
	mock.Meta.Config_ = func(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.ServerConfig, error) {
		calledMetaConfig = true
		return &sourcegraph.ServerConfig{AppURL: conf.AppURL(ctx).String()}, nil
	}
	origGetRegisteredClientForServer := getRegisteredClientForServer
	getRegisteredClientForServer = func(ctx context.Context) (*sourcegraph.RegisteredClient, error) {
		calledGetRegisteredClientForServer = true
		return &sourcegraph.RegisteredClient{}, nil
	}
	defer func() {
		getRegisteredClientForServer = origGetRegisteredClientForServer
	}()

	// Use the correct request host to ensure we don't trigger the
	// "detect app URL mismatch" check (since we are testing another
	// check).
	u := conf.AppURL(mock.Ctx).ResolveReference(router.Rel.URLTo(router.OAuth2ClientInitiate))
	resp, err := c.GetNoFollowRedirects(u.String())
	if err != nil {
		t.Fatalf("GetNoFollowRedirects: %s", err)
	}

	if want := http.StatusInternalServerError; resp.StatusCode != want {
		t.Errorf("got status %d, want %d", resp.StatusCode, want)
	}

	defer resp.Body.Close()

	// Check response is the oauth-success page.
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll: %s", err)
	}
	if want := "misconfiguration"; !strings.Contains(string(respData), want) {
		t.Errorf("got resp %q, want to contain %q", respData, want)
	}

	if !calledMetaConfig {
		t.Error("!calledMetaConfig")
	}
	if !calledGetRegisteredClientForServer {
		t.Error("!calledGetRegisteredClientForServer")
	}
}

func TestOAuthAuthorizeClientState(t *testing.T) {
	st := oauthAuthorizeClientState{Nonce: "123", ReturnTo: "/foo"}
	text, err := st.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if want := "123:/foo"; string(text) != want {
		t.Errorf("got %q, want %q", text, want)
	}

	var st2 oauthAuthorizeClientState
	if err := st2.UnmarshalText([]byte("123:/foo")); err != nil {
		t.Fatal(err)
	}
	if want := (oauthAuthorizeClientState{Nonce: "123", ReturnTo: "/foo"}); st2 != want {
		t.Errorf("got %+v, want %+v", st2, want)
	}
}

func TestOAuth2ClientReceive(t *testing.T) {
	authutil.ActiveFlags.Source = "oauth"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()
	origFedRoot := fed.Config.RootURLStr
	fed.Config.RootURLStr = "http://oauth.example.com"
	defer func() {
		fed.Config.RootURLStr = origFedRoot
	}()

	c, mock := apptest.New()

	mock.Ctx = oauth2client.WithClientID(mock.Ctx, "a")

	var calledAuthGetAccessToken, calledAuthIdentify bool
	mock.Auth.GetAccessToken_ = func(ctx context.Context, op *sourcegraph.AccessTokenRequest) (*sourcegraph.AccessTokenResponse, error) {
		calledAuthGetAccessToken = true
		return &sourcegraph.AccessTokenResponse{AccessToken: "t"}, nil
	}
	mock.Auth.Identify_ = func(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.AuthInfo, error) {
		calledAuthIdentify = true
		return &sourcegraph.AuthInfo{UID: 1, Login: "u"}, nil
	}

	u := router.Rel.URLTo(router.OAuth2ClientReceive)
	q, err := query.Values(oauth2util.ReceiveParams{ClientID: "a", Code: "c", State: "123:/foo"})
	if err != nil {
		t.Fatal(err)
	}
	u.RawQuery = q.Encode()
	req, _ := http.NewRequest("GET", u.String(), nil)
	addNonceCookie(req, "123")
	resp, err := c.DoNoFollowRedirects(req)
	if err != nil {
		t.Fatalf("DoNoFollowRedirects: %s", err)
	}

	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("got status %d, want %d", resp.StatusCode, want)
	}

	if !calledAuthGetAccessToken {
		t.Error("!calledAuthGetAccessToken")
	}
	if !calledAuthIdentify {
		t.Error("!calledAuthIdentify")
	}

	// Check the nonce is DELETED (set to empty and expired) in the response, to prevent reuse.
	nonce, present := nonceFromResponseCookie(resp)
	if !present || nonce != "" {
		t.Errorf("got nonce %q, want empty", nonce)
	}

	defer resp.Body.Close()

	// Check response is the oauth-success page.
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll: %s", err)
	}

	respStr := string(respData)
	if !strings.Contains(respStr, `id="return-oauth"`) {
		t.Errorf("response does not contain return-oauth anchor tag")
	}

	if !strings.Contains(respStr, `data-url="/foo"`) {
		t.Errorf("response does not contain correct redirect url")
	}
}

// TestOAuth2ClientReceive_invalid tests the various cases when the
// OAuth2 state does not match up with the nonce cookie (i.e., when
// someone is tampering with the OAuth2 flow), or when the OAuth2
// client ID is invalid.
func TestOAuth2ClientReceive_invalid(t *testing.T) {
	authutil.ActiveFlags.Source = "oauth"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()
	origFedRoot := fed.Config.RootURLStr
	fed.Config.RootURLStr = "http://oauth.example.com"
	defer func() {
		fed.Config.RootURLStr = origFedRoot
	}()

	c, mock := apptest.New()

	mock.Ctx = oauth2client.WithClientID(mock.Ctx, "a")

	tests := []struct {
		state          string
		clientID       string
		nonceInCookie  string
		wantStatusCode int
	}{
		// Test many permutations of state and nonce, just to be safe.
		{state: "123:/foo", clientID: "a", nonceInCookie: "456", wantStatusCode: http.StatusForbidden},
		{state: "badformat", clientID: "a", nonceInCookie: "", wantStatusCode: http.StatusBadRequest},
		{state: "badformat", clientID: "a", nonceInCookie: "456", wantStatusCode: http.StatusBadRequest},
		{state: ":/foo", clientID: "a", nonceInCookie: "", wantStatusCode: http.StatusForbidden},
		{state: ":/foo", clientID: "a", nonceInCookie: "456", wantStatusCode: http.StatusForbidden},
		{state: ":/foo", clientID: "a", nonceInCookie: "", wantStatusCode: http.StatusForbidden},
		{state: "", clientID: "a", nonceInCookie: "", wantStatusCode: http.StatusBadRequest},

		// Invalid client IDs.
		{clientID: "wrong", wantStatusCode: http.StatusForbidden},
		{clientID: "", wantStatusCode: http.StatusBadRequest},
	}
	for _, test := range tests {
		u := router.Rel.URLTo(router.OAuth2ClientReceive)
		q, err := query.Values(oauth2util.ReceiveParams{ClientID: test.clientID, Code: "c", State: test.state})
		if err != nil {
			t.Fatal(err)
		}
		u.RawQuery = q.Encode()
		req, _ := http.NewRequest("GET", u.String(), nil)
		addNonceCookie(req, test.nonceInCookie) // wrong nonce
		resp, err := c.DoNoFollowRedirects(req)
		if err != nil {
			t.Fatalf("%v: DoNoFollowRedirects: %s", test, err)
		}
		if resp.StatusCode != test.wantStatusCode {
			t.Errorf("%v: got status %d, want %d", test, resp.StatusCode, test.wantStatusCode)
		}
	}
}

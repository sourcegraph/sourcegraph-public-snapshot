package oauth2client

import (
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/internal/returnto"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/client/pkg/oauth2client"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/fed"

	// Register the login handler.
	_ "src.sourcegraph.com/sourcegraph/app/internal/localauth"
)

// TestLogIn_OAuthRedirect tests that if the auth source is "oauth",
// the login handler redirects to the provider and sets a session
// cookie to combat CSRF.
func TestLogIn_OAuthRedirect(t *testing.T) {
	authutil.ActiveFlags.Source = "oauth"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()
	origFedRoot := fed.Config.RootURLStr
	fed.Config.RootURLStr = "http://oauth.example.com"
	defer func() {
		fed.Config.RootURLStr = origFedRoot
	}()

	origGetRegisteredClientForServer := getRegisteredClientForServer
	defer func() {
		getRegisteredClientForServer = origGetRegisteredClientForServer
	}()

	c, mock := apptest.New()

	mock.Ctx = oauth2client.WithClientID(mock.Ctx, "a")

	var jwks []byte
	mock.Ctx, jwks = generateJWKS(t, mock.Ctx)

	for _, method := range []string{"GET", "POST"} {
		var calledMetaConfig, calledGetRegisteredClientForServer bool
		mock.Meta.Config_ = func(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.ServerConfig, error) {
			calledMetaConfig = true
			return &sourcegraph.ServerConfig{AppURL: conf.AppURL(ctx).String()}, nil
		}
		getRegisteredClientForServer = func(ctx context.Context) (*sourcegraph.RegisteredClient, error) {
			calledGetRegisteredClientForServer = true
			return &sourcegraph.RegisteredClient{RedirectURIs: []string{"http://example.com"}}, nil
		}

		// Use the correct request host to ensure we don't trigger the
		// "detect app URL mismatch" check.
		u := conf.AppURL(mock.Ctx).ResolveReference(router.Rel.URLTo(router.OAuth2ClientInitiate))
		returnto.SetOnURL(u, "/foo")
		req, _ := http.NewRequest(method, u.String(), nil)
		resp, err := c.DoNoFollowRedirects(req)
		if err != nil {
			t.Fatalf("%s: DoNoFollowRedirects: %s", method, err)
		}

		if want := http.StatusSeeOther; resp.StatusCode != want {
			t.Errorf("%s: got status %d, want %d", method, resp.StatusCode, want)
			continue
		}

		// Check that a nonce is set in the session to combat CSRF.
		nonce, present := nonceFromResponseCookie(resp)
		if !present || len(nonce) < 10 {
			t.Errorf("%s: got bad nonce %q", method, nonce)
		}

		stateStr, err := (oauthAuthorizeClientState{Nonce: nonce, ReturnTo: "/foo"}).MarshalText()
		if err != nil {
			t.Fatal(err)
		}

		want := "http://oauth.example.com/login/oauth/authorize?client_id=a&jwks=" + url.QueryEscape(string(jwks)) + "&redirect_uri=http%3A%2F%2Fexample.com%2Flogin%2Foauth%2Freceive&response_type=code&state=" + url.QueryEscape(string(stateStr))
		if got := resp.Header.Get("location"); got != want {
			t.Errorf("%s: got Location %q, want %q", method, got, want)
		}

		if !calledMetaConfig {
			t.Error("!calledMetaConfig")
		}
		if !calledGetRegisteredClientForServer {
			t.Error("!calledGetRegisteredClientForServer")
		}
	}
}

func generateJWKS(t *testing.T, ctx context.Context) (context.Context, []byte) {
	var jwks []byte

	// TODO(sqs!): remove the need for this by making it so the
	// client only needs the JWKS, not the full id private key.
	idkey.SetTestEnvironment()
	k, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}
	ctx = idkey.NewContext(ctx, k)

	jwks, err = k.MarshalJWKSPublicKey()
	if err != nil {
		t.Fatal(err)
	}

	return ctx, jwks
}

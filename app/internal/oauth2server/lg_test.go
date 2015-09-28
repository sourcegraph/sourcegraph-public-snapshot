// +build exectest

package oauth2server_test

import (
	"fmt"
	"net/url"
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/surf"
	"sourcegraph.com/sourcegraph/surf/browser"
	"sourcegraph.com/sqs/pbtypes"

	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/sharedsecret"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/fed"
	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil"
)

// TestOAuthAuthorize_lg tests OAuth authorization when anon readers
// are disallowed.
func TestOAuthAuthorize_lg(t *testing.T) {
	t.Parallel()
	testOAuthAuthorizeProcess(t, nil, nil)
}

// TestOAuthAuthorize_lg_anonReadersAllowedOnAS tests OAuth
// authorization when anon readers are allowed on the federation root
// (Authorization Server) Sourcegraph instance.
func TestOAuthAuthorize_lg_anonReadersAllowedOnAS(t *testing.T) {
	t.Parallel()
	testOAuthAuthorizeProcess(t, anonReadersAllowedOnAS, nil)
}

// TestOAuthAuthorize_lg_anonReadersAllowedOnC tests OAuth
// authorization when anon readers are allowed on the client Sourcegraph instance.
func TestOAuthAuthorize_lg_anonReadersAllowedOnC(t *testing.T) {
	t.Parallel()
	testOAuthAuthorizeProcess(t, nil, anonReadersAllowedOnC)
}

// TestOAuthAuthorize_lg_anonReadersAllowedOnBoth tests OAuth
// authorization when anon readers are allowed on both the authorization
// server and client Sourcegraph instances.
func TestOAuthAuthorize_lg_anonReadersAllowedOnBoth(t *testing.T) {
	t.Parallel()
	testOAuthAuthorizeProcess(t, anonReadersAllowedOnAS, anonReadersAllowedOnC)
}

// TestOAuthRegisterClient_lg tests OAuth client registration and
// onboarding when anon readers are disallowed.
func TestOAuthRegisterClient_lg(t *testing.T) {
	t.Parallel()
	testOAuthRegisterClientProcess(t, nil, nil)
}

// TestOAuthRegisterClient_lg_anonReadersAllowedOnAS tests OAuth
// client registration and onboarding when anon readers are allowed on
// the federation root (Authorization Server) Sourcegraph instance.
func TestOAuthRegisterClient_lg_anonReadersAllowedOnAS(t *testing.T) {
	t.Parallel()
	testOAuthRegisterClientProcess(t, anonReadersAllowedOnAS, nil)
}

// TestOAuthRegisterClient_lg_anonReadersAllowedOnC tests OAuth client
// registration and onboarding when anon readers are allowed on the
// client Sourcegraph instance.
func TestOAuthRegisterClient_lg_anonReadersAllowedOnC(t *testing.T) {
	t.Parallel()
	testOAuthRegisterClientProcess(t, nil, anonReadersAllowedOnC)
}

// TestOAuthRegisterClient_lg_anonReadersAllowedOnBoth tests OAuth
// client registration and onboarding when anon readers are allowed on
// both the authorization server and client Sourcegraph instances.
func TestOAuthRegisterClient_lg_anonReadersAllowedOnBoth(t *testing.T) {
	t.Parallel()
	testOAuthRegisterClientProcess(t, anonReadersAllowedOnAS, anonReadersAllowedOnC)
}

// anonReadersAllowedOnC configures s (which is an OAuth2 client) to
// allow anonymous readers.
func anonReadersAllowedOnC(s *testserver.Server) {
	s.Config.ServeFlags = append(s.Config.ServeFlags, &authutil.Flags{AllowAnonymousReaders: true, Source: "oauth"})
}

// anonReadersAllowedOnAS configures s (which is an OAuth2
// Authorization Server) to allow anonymous readers.
func anonReadersAllowedOnAS(s *testserver.Server) {
	s.Config.ServeFlags = append(s.Config.ServeFlags, &authutil.Flags{AllowAnonymousReaders: true, Source: "local", OAuth2AuthServer: true, AllowAllLogins: true})
}

// newOAuthTest starts 2 Sourcegraph servers: the root/AS
// (Authentication Server), and the client.
//
// The caller should call oas.Close() and oc.Close() when done
// (probably in a defer statement).
func newOAuthTest(t *testing.T, oasFunc func(*testserver.Server), ocFunc func(*testserver.Server)) (oas *testserver.Server, ctxOAS context.Context, oc *testserver.Server, ctxOC context.Context) {
	httptestutil.ResetGlobals()

	// "oas" for OAuth authorization server
	oas, ctxOAS = testserver.NewUnstartedServer()
	oas.Config.ServeFlags = append(oas.Config.ServeFlags,
		&fed.Flags{IsRoot: true},
		&authutil.Flags{Source: "local", OAuth2AuthServer: true, AllowAllLogins: true},
	)
	useIPAsHostnameToAvoidSharingCookiesBetweenTheServers(oas)
	if oasFunc != nil {
		oasFunc(oas)
	}
	ctxOAS = oas.Ctx // useIPAsHostnameToAvoidSharingCookiesBetweenTheServers mutates oas.Ctx, so re-read it
	if err := oas.Start(); err != nil {
		t.Fatal(err)
	}

	// "oc" for OAuth client
	oc, ctxOC = testserver.NewUnstartedServer()
	oc.Config.ServeFlags = append(oc.Config.ServeFlags,
		&fed.Flags{RootURLStr: conf.AppURL(ctxOAS).String()},
		&authutil.Flags{Source: "oauth"},
	)
	if ocFunc != nil {
		ocFunc(oc)
	}
	if err := oc.Start(); err != nil {
		t.Fatal(err)
	}

	return oas, ctxOAS, oc, ctxOC
}

// testOAuthAuthorizeProcess tests that the OAuth authorization flow
// works.
//
// Unlike testOAuthRegisterClientProcess, this tests the behavior of
// an already-onboarded (client-registered) Sourcegraph instance.
func testOAuthAuthorizeProcess(t *testing.T, oasFunc func(*testserver.Server), ocFunc func(*testserver.Server)) {
	oas, ctxOAS, oc, ctxOC := newOAuthTest(t, oasFunc, ocFunc)
	defer oas.Close()
	defer oc.Close()

	{
		// Register C as a client of AS before starting C.
		k := idkey.FromContext(ctxOC)
		jwks, err := k.MarshalJWKSPublicKey()
		if err != nil {
			t.Fatal(err)
		}
		_, err = oas.Client.RegisteredClients.Create(sharedsecret.NewContext(ctxOAS, "x"), &sourcegraph.RegisteredClient{
			ID:           k.ID,
			ClientName:   "myclient",
			JWKS:         string(jwks),
			RedirectURIs: []string{conf.AppURL(ctxOC).String()},
			Type:         sourcegraph.RegisteredClientType_SourcegraphServer,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	b := surf.NewBrowser()

	// Initiate login on C.
	u := router.Rel.URLTo(router.LogIn)
	returnto.SetOnURL(u, "/")
	loginURL := conf.AppURL(ctxOC).ResolveReference(u)
	if err := b.Open(loginURL.String()); err != nil {
		t.Fatal(err)
	}

	testOAuthLogin(t, b, oas, ctxOAS, oc, ctxOC)
	testOAuthAuthorize(t, b, oas, ctxOAS, oc, ctxOC)
}

// testOAuthRegisterClientProcess tests that the OAuth client
// registration and authorization flow (for onboarding a new
// Sourcegraph installation) works.
//
// Unlike testOAuthAuthorizeProcess, this tests the onboarding flow.
func testOAuthRegisterClientProcess(t *testing.T, oasFunc func(*testserver.Server), ocFunc func(*testserver.Server)) {
	oas, ctxOAS, oc, ctxOC := newOAuthTest(t, oasFunc, ocFunc)
	defer oas.Close()
	defer oc.Close()

	b := surf.NewBrowser()

	// Initiate login on C.
	confC, err := oc.Client.Meta.Config(ctxOC, &pbtypes.Void{})
	if err != nil {
		t.Fatal(err)
	}
	var initialURL *url.URL
	if confC.AllowAnonymousReaders {
		// If anonymous access is permitted, we must go to C's login
		// page to trigger login.
		u := router.Rel.URLTo(router.LogIn)
		returnto.SetOnURL(u, "/")
		initialURL = conf.AppURL(ctxOC).ResolveReference(u)
	} else {
		// Otherwise, if auth is required on C, we can just go to the
		// homepage.
		initialURL = conf.AppURL(ctxOC).ResolveReference(router.Rel.URLTo(router.Home))
	}

	if err := b.Open(initialURL.String()); err != nil {
		t.Fatal(err)
	}

	// Check that we're redirected to the AS's register client
	// flow. The first step in the flow is the user login page.
	testOAuthLogin(t, b, oas, ctxOAS, oc, ctxOC)
	testOAuthRegisterClient(t, b, oas, ctxOAS, oc, ctxOC)
	testOAuthAuthorize(t, b, oas, ctxOAS, oc, ctxOC)
}

func testOAuthRegisterClient(t *testing.T, b *browser.Browser, oas *testserver.Server, ctxOAS context.Context, oc *testserver.Server, ctxOC context.Context) {
	// Check that we're on the OAuth2 authorization screen on AS.
	wantURL := conf.AppURL(ctxOAS).ResolveReference(router.Rel.URLTo(router.RegisterClient))
	if !strings.HasPrefix(b.Url().String(), wantURL.String()) {
		t.Fatalf("after logging in on AS, got URL %s, want authz page on AS %s", b.Url(), wantURL)
	}

	// The actual test:
	f, err := b.Form("#register-client")
	if err != nil {
		printAppStatus(t, b)
		t.Fatal(err)
	}
	if err := f.Input("ClientName", "Some Company"); err != nil {
		t.Fatal(err)
	}
	if err := f.Submit(); err != nil {
		t.Fatal(err)
	}

	// Make sure it registered the client. Call
	// RegisteredClients.GetCurrent on the AS *but* using the C's
	// client credentials.
	clientAuthedCtx := sourcegraph.WithCredentials(oas.Ctx, sharedsecret.TokenSource(idkey.FromContext(ctxOC)))
	rcl, err := oas.Client.RegisteredClients.GetCurrent(clientAuthedCtx, &pbtypes.Void{})
	if err != nil {
		t.Fatal(err)
	}
	if want := "Some Company"; rcl.ClientName != want {
		t.Fatalf("got ClientName == %q, want %q", rcl.ClientName, want)
	}
}

// Use 127.0.0.1, not localhost, for one host, so that the two app
// servers don't share cookies. Confusing test failures occur if they
// share cookies.
func useIPAsHostnameToAvoidSharingCookiesBetweenTheServers(a *testserver.Server) {
	a.Config.Serve.AppURL = strings.Replace(a.Config.Serve.AppURL, "localhost", "127.0.0.1", 1)
	u, _ := url.Parse(a.Config.Serve.AppURL)
	a.Ctx = conf.WithAppURL(a.Ctx, u)
}

func testOAuthLogin(t *testing.T, b *browser.Browser, oas *testserver.Server, ctxOAS context.Context, oc *testserver.Server, ctxOC context.Context) {
	// Create the user on the AS.
	if _, err := testutil.CreateAccount(t, ctxOAS, "alice"); err != nil {
		t.Fatal(err)
	}

	{
		// Click welcome interstitial if it shows up.
		err := b.Click("#continue-oauth")
		if err != nil && err.Error() != "Element not found matching expr '#continue-oauth'." {
			printAppStatus(t, b)
			t.Fatal(err)
		}
	}

	{
		// Check that we're on the login screen on AS.
		wantASLoginURL := conf.AppURL(ctxOAS).ResolveReference(router.Rel.URLTo(router.LogIn))
		if !strings.HasPrefix(b.Url().String(), wantASLoginURL.String()) {
			t.Fatalf("after initiating login on C, got URL %s, want login page on AS %s", b.Url(), wantASLoginURL)
		}
	}

	{
		// Submit login form on AS.
		f, err := b.Form("form.log-in")
		if err != nil {
			printAppStatus(t, b)
			t.Fatal(err)
		}
		f.Input("Login", "alice")
		f.Input("Password", testutil.Password)
		if err := f.Submit(); err != nil {
			t.Fatal(err)
		}
	}
}

func testOAuthAuthorize(t *testing.T, b *browser.Browser, oas *testserver.Server, ctxOAS context.Context, oc *testserver.Server, ctxOC context.Context) {
	{
		if err := checkAppError(b); err != nil {
			t.Fatal(err)
		}

		// Check that we're on the OAuth2 authorization screen on AS.
		wantAuthzURL := conf.AppURL(ctxOAS).ResolveReference(router.Rel.URLTo(router.OAuth2ServerAuthorize))
		if !strings.HasPrefix(b.Url().String(), wantAuthzURL.String()) {
			t.Fatalf("after logging in on AS, got URL %s, want authz page on AS %s", b.Url(), wantAuthzURL)
		}

		// Authorize C.
		if err := b.Click(".authorize.btn"); err != nil {
			t.Fatal(err)
		}

		// Want to be redirected to C, logged in as the user.
		if wantURL := conf.AppURL(ctxOC); !strings.HasPrefix(b.Url().String(), wantURL.String()) {
			t.Fatalf("after authorizing on AS, got URL %s, want to be back at C %s", b.Url(), wantURL)
		}

		if err := checkAppError(b); err != nil {
			t.Fatal(err)
		}
		if got, want := CurrentUserFromDOM(b), "alice"; got != want {
			t.Errorf("got current user == %q, want %q", got, want)
		}
	}
}

func checkAppError(b *browser.Browser) error {
	if txt := b.Find(".error-message").Text(); txt != "" {
		return fmt.Errorf("app error: %s - at %s", txt, b.Url())
	}
	return nil
}

func printAppStatus(t *testing.T, b *browser.Browser) {
	if err := checkAppError(b); err != nil {
		t.Log(err)
	} else {
		t.Log("at", b.Url())
	}
}

func CurrentUserFromDOM(b *browser.Browser) string {
	login, _ := b.Find("head").Attr("data-current-user-login")
	return login
}

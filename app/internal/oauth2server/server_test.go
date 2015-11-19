package oauth2server_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-querystring/query"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/internal/returnto"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/pkg/oauth2util"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func TestOAuth2ServerAuthorize_notEnabled(t *testing.T) {
	authutil.ActiveFlags = authutil.Flags{Source: "local", OAuth2AuthServer: false}
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, mock := apptest.New()

	mock.Ctx = handlerutil.WithUser(mock.Ctx, &sourcegraph.UserSpec{UID: 1})
	resp, err := c.Get(router.Rel.URLTo(router.OAuth2ServerAuthorize).String())
	if err != nil {
		t.Fatalf("Get: %s", err)
	}

	if want := http.StatusNotFound; resp.StatusCode != want {
		t.Errorf("got status %d, want %d", resp.StatusCode, want)
	}
}

func TestOAuth2ServerAuthorize_notLoggedIn(t *testing.T) {
	authutil.ActiveFlags = authutil.Flags{Source: "local", OAuth2AuthServer: true, AllowAllLogins: true}
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, _ := apptest.New()

	u := router.Rel.URLTo(router.OAuth2ServerAuthorize)
	resp, err := c.GetNoFollowRedirects(u.String())
	if err != nil {
		t.Fatalf("Get: %s", err)
	}

	if want := http.StatusSeeOther; resp.StatusCode != want {
		t.Errorf("got status %d, want %d", resp.StatusCode, want)
	}

	wantLocation := router.Rel.URLTo(router.LogIn)
	q := url.Values{}
	q.Set(returnto.ParamName, u.String())
	wantLocation.RawQuery = q.Encode()
	if got := resp.Header.Get("location"); got != wantLocation.String() {
		t.Errorf("got Location %q, want %q", got, wantLocation)
	}
}

func TestOAuth2ServerAuthorize(t *testing.T) {
	authutil.ActiveFlags = authutil.Flags{Source: "local", OAuth2AuthServer: true, AllowAllLogins: true}
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, mock := apptest.New()

	var calledRegisteredClientsGet, calledAuthGetAuthorizationCode bool
	mock.RegisteredClients.Get_ = func(ctx context.Context, regClient *sourcegraph.RegisteredClientSpec) (*sourcegraph.RegisteredClient, error) {
		calledRegisteredClientsGet = true
		if want := "a"; regClient.ID != want {
			t.Errorf("got client ID == %q, want %q", regClient.ID, want)
		}
		return &sourcegraph.RegisteredClient{ID: "a", RedirectURIs: []string{"http://example.com"}}, nil
	}
	mock.Auth.GetAuthorizationCode_ = func(ctx context.Context, in *sourcegraph.AuthorizationCodeRequest) (*sourcegraph.AuthorizationCode, error) {
		calledAuthGetAuthorizationCode = true
		return &sourcegraph.AuthorizationCode{RedirectURI: "http://example.com/r", Code: "mycode"}, nil
	}

	u := router.Rel.URLTo(router.OAuth2ServerAuthorize)
	q, err := query.Values(oauth2util.AuthorizeParams{ClientID: "a", RedirectURI: "http://example.com/r", State: "s"})
	if err != nil {
		t.Fatal(err)
	}
	u.RawQuery = q.Encode()
	mock.Ctx = handlerutil.WithUser(mock.Ctx, &sourcegraph.UserSpec{UID: 1})
	resp, err := c.Get(u.String())
	if err != nil {
		t.Fatalf("Get: %s", err)
	}

	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("got status %d, want %d", resp.StatusCode, want)
	}

	if !calledRegisteredClientsGet {
		t.Error("!calledRegisteredClientsGet")
	}
	if !calledAuthGetAuthorizationCode {
		t.Error("!calledAuthGetAuthorizationCode")
	}

	// Check that an "Authorize" button directs the user back to the
	// OAuth2 client with the requisite auth info.
	pg, err := readOAuth2ServerAuthorizePage(resp)
	if err != nil {
		t.Fatal(err)
	}

	wantAuthorizeURL := "http://example.com/r?client_id=a&code=mycode&state=s"
	if pg.authorizeURL != wantAuthorizeURL {
		t.Errorf("got page authorize URL %q, want %q", pg.authorizeURL, wantAuthorizeURL)
	}
}

// oauthProviderAuthorizePage describes the "Authorize client" page
// that OAuth2 providers display to users during the OAuth2 flow.
type oauthProviderAuthorizePage struct {
	authorizeURL string
}

func readOAuth2ServerAuthorizePage(resp *http.Response) (*oauthProviderAuthorizePage, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var pg oauthProviderAuthorizePage
	pg.authorizeURL, _ = doc.Find("a.authorize").Attr("href")
	return &pg, nil
}

func TestOAuth2ServerAuthorize_newClientID(t *testing.T) {
	authutil.ActiveFlags = authutil.Flags{Source: "local", OAuth2AuthServer: true, AllowAllLogins: true}
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, mock := apptest.New()

	mock.RegisteredClients.Get_ = func(ctx context.Context, regClient *sourcegraph.RegisteredClientSpec) (*sourcegraph.RegisteredClient, error) {
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	newClientID := "x"
	u := router.Rel.URLTo(router.OAuth2ServerAuthorize)
	q, err := query.Values(oauth2util.AuthorizeParams{ClientID: newClientID})
	if err != nil {
		t.Fatal(err)
	}
	u.RawQuery = q.Encode()
	mock.Ctx = handlerutil.WithUser(mock.Ctx, &sourcegraph.UserSpec{UID: 1})
	resp, err := c.GetNoFollowRedirects(u.String())
	if err != nil {
		t.Fatalf("client ID %q: Get: %s", newClientID, err)
	}
	if want := http.StatusSeeOther; resp.StatusCode != want {
		t.Errorf("client ID %q: got status %d, want %d", newClientID, resp.StatusCode, want)
	}
}

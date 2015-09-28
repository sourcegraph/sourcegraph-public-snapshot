package oauth2server_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-querystring/query"
	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/pkg/oauth2util"
)

func TestServeRegisterClient(t *testing.T) {
	c, mock := apptest.New()

	var called bool
	mock.RegisteredClients.Create_ = func(ctx context.Context, in *sourcegraph.RegisteredClient) (*sourcegraph.RegisteredClient, error) {
		called = true
		if want := "n"; in.ClientName != want {
			t.Errorf(`got ClientName == %q, want %q`, in.ClientName, want)
		}
		return &sourcegraph.RegisteredClient{}, nil
	}

	// URL
	u := router.Rel.URLTo(router.RegisterClient)
	authOpt, err := query.Values(&oauth2util.AuthorizeParams{ClientID: "c", JWKS: "j", RedirectURI: "u"})
	if err != nil {
		t.Fatal(err)
	}
	u.RawQuery = authOpt.Encode()

	// Form
	form := make(url.Values)
	form.Set("ClientName", "n")

	resp, err := c.PostFormNoFollowRedirects(u.String(), form)
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusSeeOther; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}
	wantLoc := router.Rel.URLTo(router.OAuth2ServerAuthorize)
	wantLoc.RawQuery = authOpt.Encode()
	if loc := resp.Header.Get("location"); loc != wantLoc.String() {
		t.Errorf("got Location %q, want %q", loc, wantLoc)
	}

	if !called {
		t.Error("!called")
	}
}

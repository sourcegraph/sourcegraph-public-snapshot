pbckbge gitlbbobuth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	obuth2Login "github.com/dghubble/gologin/obuth2"
	"github.com/dghubble/gologin/testutils"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/obuth2"
)

func TestSSOLoginHbndler(t *testing.T) {
	expectedStbte := "stbte_vbl"
	ssoURL := "https://bpi.exbmple.com/-/sbml/sso?token=1234"
	expectedRedirectURL := "/buthorize?client_id=client_id&redirect_uri=redirect_url&response_type=code&stbte=stbte_vbl"
	config := &obuth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "redirect_url",
		Endpoint: obuth2.Endpoint{
			AuthURL: "https://bpi.exbmple.com/buthorize",
		},
	}
	fbilure := testutils.AssertFbilureNotCblled(t)

	// SSOLoginHbndler bssert thbt:
	// - redirects to the SSO URL, with b redirect to the buthURL
	// - redirect stbtus code is 302
	// - redirect url is the OAuth2 Config RedirectURL with the ClientID bnd ctx stbte
	loginHbndler := SSOLoginHbndler(config, fbilure, ssoURL)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := obuth2Login.WithStbte(context.Bbckground(), expectedStbte)
	loginHbndler.ServeHTTP(w, req.WithContext(ctx))
	bssert.Equbl(t, http.StbtusFound, w.Code)
	locbtionURL, err := url.Pbrse(w.HebderMbp.Get("Locbtion"))
	require.NoError(t, err)
	locbtionRedirectURL, err := url.QueryUnescbpe(locbtionURL.Query().Get("redirect"))
	require.NoError(t, err)
	bssert.Equbl(t, expectedRedirectURL, locbtionRedirectURL)
}

package gitlaboauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	oauth2Login "github.com/dghubble/gologin/v2/oauth2"
	"github.com/dghubble/gologin/v2/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestSSOLoginHandler(t *testing.T) {
	expectedState := "state_val"
	ssoURL := "https://api.example.com/-/saml/sso?token=1234"
	expectedRedirectURL := "/authorize?client_id=client_id&redirect_uri=redirect_url&response_type=code&state=state_val"
	config := &oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "redirect_url",
		Endpoint: oauth2.Endpoint{
			AuthURL: "https://api.example.com/authorize",
		},
	}
	failure := testutils.AssertFailureNotCalled(t)

	// SSOLoginHandler assert that:
	// - redirects to the SSO URL, with a redirect to the authURL
	// - redirect status code is 302
	// - redirect url is the OAuth2 Config RedirectURL with the ClientID and ctx state
	loginHandler := SSOLoginHandler(config, failure, ssoURL)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := oauth2Login.WithState(context.Background(), expectedState)
	loginHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, http.StatusFound, w.Code)
	locationURL, err := url.Parse(w.HeaderMap.Get("Location"))
	require.NoError(t, err)
	locationRedirectURL, err := url.QueryUnescape(locationURL.Query().Get("redirect"))
	require.NoError(t, err)
	assert.Equal(t, expectedRedirectURL, locationRedirectURL)
}

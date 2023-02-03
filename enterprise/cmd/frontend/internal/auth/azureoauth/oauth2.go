package azureoauth

// Adapted from "github.com/golang/oauth2"

// Exchange converts an authorization code into a token.
//
// It is used after a resource provider redirects the user back
// to the Redirect URI (the URL obtained from AuthCodeURL).
//
// The provided context optionally controls which HTTP client is used. See the HTTPClient variable.
//
// The code will be in the *http.Request.FormValue("code"). Before
// calling Exchange, be sure to validate FormValue("state").
//
// Opts may include the PKCE verifier code if previously used in AuthCodeURL.
// See https://www.oauth.com/oauth2-servers/pkce/ for more info.

// func (c *oauth2.Config) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*Token, error) {
// 	v := url.Values{
// 		"grant_type": {"authorization_code"},
// 		"code":       {code},
// 	}
// 	if c.RedirectURL != "" {
// 		v.Set("redirect_uri", c.RedirectURL)
// 	}
// 	for _, opt := range opts {
// 		opt.setValue(v)
// 	}
// 	return retrieveToken(ctx, c, v)
// }

// client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer&client_assertion={0}&grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion={1}&redirect_uri={2}

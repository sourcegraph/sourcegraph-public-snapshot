package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gomodule/oauth1/oauth"
)

// OAuthClient implements OAuth 1 signature authentication for extsvc
// implementations.
type OAuthClient struct{ *oauth.Client }

var _ Authenticator = &OAuthClient{}

func (c *OAuthClient) Authenticate(req *http.Request) error {
	return c.SetAuthorizationHeader(
		req.Header,
		&oauth.Credentials{Token: ""}, // Token must be empty
		req.Method,
		req.URL,
		nil,
	)
}

func (c *OAuthClient) Hash() string {
	sk := sha256.Sum256([]byte(c.Credentials.Secret))
	tk := sha256.Sum256([]byte(c.Credentials.Token))
	return hex.EncodeToString(sk[:]) + hex.EncodeToString(tk[:])
}

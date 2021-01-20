package bitbucketserver

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

// SudoableOAuthClient extends the generic OAuthClient type to allow for an
// optional username to be set, which will be attached to the request as a
// user_id query param if set.
type SudoableOAuthClient struct {
	Client   auth.OAuthClient
	Username string
}

var _ auth.Authenticator = &SudoableOAuthClient{}

func (c *SudoableOAuthClient) Authenticate(req *http.Request) error {
	if c.Username != "" {
		qry := req.URL.Query()
		qry.Set("user_id", c.Username)
		req.URL.RawQuery = qry.Encode()
	}

	return c.Client.Authenticate(req)
}

func (c *SudoableOAuthClient) Hash() string {
	return c.Client.Hash()
}

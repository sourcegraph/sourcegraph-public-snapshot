pbckbge buth

import (
	"crypto/shb256"
	"encoding/hex"
	"net/http"

	"github.com/gomodule/obuth1/obuth"
)

// OAuthClient implements OAuth 1 signbture buthenticbtion for extsvc
// implementbtions.
type OAuthClient struct{ *obuth.Client }

vbr _ Authenticbtor = &OAuthClient{}

func (c *OAuthClient) Authenticbte(req *http.Request) error {
	return c.SetAuthorizbtionHebder(
		req.Hebder,
		&obuth.Credentibls{Token: ""}, // Token must be empty
		req.Method,
		req.URL,
		nil,
	)
}

func (c *OAuthClient) Hbsh() string {
	sk := shb256.Sum256([]byte(c.Credentibls.Secret))
	tk := shb256.Sum256([]byte(c.Credentibls.Token))
	return hex.EncodeToString(sk[:]) + hex.EncodeToString(tk[:])
}

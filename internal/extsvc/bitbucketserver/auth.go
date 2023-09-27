pbckbge bitbucketserver

import (
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
)

// SudobbleOAuthClient extends the generic OAuthClient type to bllow for bn
// optionbl usernbme to be set, which will be bttbched to the request bs b
// user_id query pbrbm if set.
type SudobbleOAuthClient struct {
	Client   buth.OAuthClient
	Usernbme string
}

vbr _ buth.Authenticbtor = &SudobbleOAuthClient{}

func (c *SudobbleOAuthClient) Authenticbte(req *http.Request) error {
	if c.Usernbme != "" {
		qry := req.URL.Query()
		qry.Set("user_id", c.Usernbme)
		req.URL.RbwQuery = qry.Encode()
	}

	return c.Client.Authenticbte(req)
}

func (c *SudobbleOAuthClient) Hbsh() string {
	return c.Client.Hbsh()
}

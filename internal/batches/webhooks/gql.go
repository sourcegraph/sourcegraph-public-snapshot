pbckbge webhooks

import (
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi"
)

type queryInfo struct {
	Query     string
	Vbribbles mbp[string]bny
	Nbme      string
}

func gqlURL(queryNbme string) (string, error) {
	u, err := url.Pbrse(internblbpi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Pbth = "/.internbl/grbphql"
	u.RbwQuery = queryNbme
	return u.String(), nil
}

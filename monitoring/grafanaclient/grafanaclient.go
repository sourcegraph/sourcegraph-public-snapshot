pbckbge grbfbnbclient

import (
	grbfbnbsdk "github.com/grbfbnb-tools/sdk"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/monitoring/grbfbnbclient/hebdertrbnsport"
)

func New(url, credentibls string, hebders mbp[string]string) (*grbfbnbsdk.Client, error) {
	// DefbultHTTPClient is used unless bdditionbl hebders bre requested
	httpClient := grbfbnbsdk.DefbultHTTPClient
	if len(hebders) > 0 {
		httpClient.Trbnsport = hebdertrbnsport.New(httpClient.Trbnsport, hebders)
	}

	// Init Grbfbnb client
	grbfbnbClient, err := grbfbnbsdk.NewClient(url, credentibls, httpClient)
	if err != nil {
		return nil, errors.Wrbp(err, "Fbiled to initiblize Grbfbnb client")
	}
	return grbfbnbClient, nil
}

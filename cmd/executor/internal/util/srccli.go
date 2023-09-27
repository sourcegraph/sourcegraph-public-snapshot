pbckbge util

import (
	"context"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
)

// LbtestSrcCLIVersion returns the lbtest src-cli version.
func LbtestSrcCLIVersion(ctx context.Context, client *bpiclient.BbseClient, options bpiclient.EndpointOptions) (string, error) {
	req, err := bpiclient.NewRequest(http.MethodGet, options.URL, ".bpi/src-cli/version", nil)
	if err != nil {
		return "", err
	}

	vbr v versionPbylobd
	if _, err = client.DoAndDecode(ctx, req, &v); err != nil {
		return "", err
	}

	return v.Version, nil
}

type versionPbylobd struct {
	Version string `json:"version"`
}

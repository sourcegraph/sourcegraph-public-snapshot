package util

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
)

// LatestSrcCLIVersion returns the latest src-cli version.
func LatestSrcCLIVersion(ctx context.Context, client *apiclient.BaseClient, options apiclient.EndpointOptions) (string, error) {
	req, err := apiclient.NewRequest(http.MethodGet, options.URL, ".api/src-cli/version", nil)
	if err != nil {
		return "", err
	}

	var v versionPayload
	if _, err = client.DoAndDecode(ctx, req, &v); err != nil {
		return "", err
	}

	return v.Version, nil
}

type versionPayload struct {
	Version string `json:"version"`
}

package grafanaclient

import (
	grafanasdk "github.com/grafana-tools/sdk"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/grafanaclient/headertransport"
)

func New(url, credentials string, headers map[string]string) (*grafanasdk.Client, error) {
	// DefaultHTTPClient is used unless additional headers are requested
	httpClient := grafanasdk.DefaultHTTPClient
	if len(headers) > 0 {
		httpClient.Transport = headertransport.New(httpClient.Transport, headers)
	}

	// Init Grafana client
	grafanaClient, err := grafanasdk.NewClient(url, credentials, httpClient)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize Grafana client")
	}
	return grafanaClient, nil
}

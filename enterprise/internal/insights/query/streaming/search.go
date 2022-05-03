package streaming

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute/client"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"

	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// Opts contains the search options supported by Search.
type Opts struct {
	Display int
	Trace   bool
	Json    bool
}

// Search calls the streaming search endpoint and uses decoder to decode the
// response body.
func Search(ctx context.Context, query string, decoder streamhttp.FrontendStreamDecoder) error {
	req, err := streamhttp.NewRequest(internalapi.Client.URL+"/.internal", query)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "code-insights-backend")

	resp, err := httpcli.InternalClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	decErr := decoder.ReadAll(resp.Body)
	if decErr != nil {
		return decErr
	}
	return err
}

func ComputeMatchContextStream(ctx context.Context, query string, decoder client.ComputeMatchContextStreamDecoder) error {
	req, err := client.NewMatchContextRequest(internalapi.Client.URL+"/.internal", query)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "code-insights-backend")

	resp, err := httpcli.InternalClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	decErr := decoder.ReadAll(resp.Body)
	if decErr != nil {
		return decErr
	}
	return err
}

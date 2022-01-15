package streaming

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
)

// Opts contains the search options supported by Search.
type Opts struct {
	Display int
	Trace   bool
	Json    bool
}

// Search calls the streaming search endpoint and uses decoder to decode the
// response body.
func Search(ctx context.Context, query string, decoder decoder) error {
	internalURL, err := url.Parse(internalapi.Client.URL)
	if err != nil {
		return err
	}
	// Create request.
	p := ".api/search/stream?q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, "GET", strings.TrimRight(internalURL.String(), "/")+"/"+p, nil)
	if err != nil {
		return err
	}
	// req, err := client.NewHTTPRequest(context.Background(), "GET", , nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")

	// Send request.
	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "Post")
	}
	defer resp.Body.Close()

	// resp, byte, err := client.Do(req)
	// if err != nil {
	// 	return fmt.Errorf("error sending request: %w", err)
	// }
	// defer resp.Body.Close()

	// Process response.
	err = decoder.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error during decoding: %w", err)
	}

	return nil
}

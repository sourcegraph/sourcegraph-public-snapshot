package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func makeRequest[T any](ctx context.Context, q queryInfo, client httpcli.Doer, res T) error {
	reqBody, err := json.Marshal(q)
	if err != nil {
		return errors.Wrap(err, "marshal request body")
	}

	url, err := gqlURL(q.Name)
	if err != nil {
		return errors.Wrap(err, "construct frontend URL")
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return errors.Wrap(err, "construct request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return errors.Wrap(err, "decode response")
	}

	return nil
}

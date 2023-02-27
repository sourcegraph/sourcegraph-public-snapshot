package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func nullable[T any](value T) *T {
	if reflect.ValueOf(value).IsZero() {
		return nil
	}
	return &value
}

func nullableMap[T, U any](value T, mapper func(T) U) *U {
	if reflect.ValueOf(value).IsZero() {
		return nil
	}
	mapped := mapper(value)
	return &mapped
}

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

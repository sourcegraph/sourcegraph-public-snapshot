package dotcom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const ProviderName = "dotcom"

var done_bytes = []byte("done")

const (
	completionsURL = "https://sourcegraph.com/.api/completions/stream"
	codeURL        = "https://sourcegraph.com/.api/completions/code"
)

type dotcomClient struct {
	cli         httpcli.Doer
	accessToken string
}

func NewClient(cli httpcli.Doer, accessToken string) types.CompletionsClient {
	return &dotcomClient{
		cli:         cli,
		accessToken: accessToken,
	}
}

func (a *dotcomClient) Complete(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	reqBody, err := json.Marshal(requestParams)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", codeURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", a.accessToken))

	resp, err := a.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, errors.Errorf("API failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result types.CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "failed to decode payload")
	}
	return &result, nil
}

func (a *dotcomClient) Stream(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	reqBody, err := json.Marshal(requestParams)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", completionsURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", a.accessToken))

	resp, err := a.cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return errors.Errorf("API failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	dec := streamhttp.NewDecoder(resp.Body)
	for dec.Scan() {
		eventBytes := dec.Event()
		// Check if the stream is indicating it is done
		if bytes.Equal(eventBytes, done_bytes) {
			return nil
		}

		var event types.CompletionResponse
		if err := json.Unmarshal(dec.Data(), &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w", err)
		}

		err = sendEvent(event)
		if err != nil {
			return err
		}
	}

	return dec.Err()
}

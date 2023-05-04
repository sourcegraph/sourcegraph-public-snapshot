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

var ProviderName = "dotcom"
var done_bytes = []byte("done")

const api_url = "https://sourcegraph.com/.api/completions/stream"

type dotcomClient struct {
	cli         httpcli.Doer
	accessToken string
	model       string
}

func NewClient(cli httpcli.Doer, accessToken string, model string) types.CompletionsClient {
	return &dotcomClient{
		cli:         cli,
		accessToken: accessToken,
		model:       model,
	}
}

func (a *dotcomClient) Complete(
	ctx context.Context,
	requestParams types.CodeCompletionRequestParameters,
) (*types.CodeCompletionResponse, error) {
	return nil, errors.New("not implemented")
}

func (a *dotcomClient) Stream(
	ctx context.Context,
	requestParams types.ChatCompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	reqBody, err := json.Marshal(requestParams)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", api_url, bytes.NewReader(reqBody))
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
		respBody, _ := io.ReadAll(resp.Body)
		return errors.Errorf("API failed with: %s", string(respBody))
	}

	dec := streamhttp.NewDecoder(resp.Body)
	for dec.Scan() {
		eventBytes := dec.Event()
		// Check if the stream is indicating it is done
		if bytes.Equal(eventBytes, done_bytes) {
			return nil
		}

		var event types.ChatCompletionEvent
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

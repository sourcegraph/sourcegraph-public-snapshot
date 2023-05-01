package passthrough

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

var DONE_BYTES = []byte("done")

const API_URL = "https://sourcegraph.sourcegraph.com/.api/completions/stream"

type passthroughClient struct {
	cli         httpcli.Doer
	accessToken string
	model       string
	url         string
}

func NewPassthoughClient(cli httpcli.Doer, url string, accessToken string, model string) types.CompletionsClient {
	return &passthroughClient{
		cli:         cli,
		accessToken: accessToken,
		model:       model,
		url:         url,
	}
}

func (a *passthroughClient) Complete(
	ctx context.Context,
	requestParams types.CodeCompletionRequestParameters,
) (*types.CodeCompletionResponse, error) {
	return nil, errors.New("not implemented")
}

func (a *passthroughClient) Stream(
	ctx context.Context,
	requestParams types.ChatCompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {

	reqBody, err := json.Marshal(requestParams)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", a.url, bytes.NewReader(reqBody))
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
		if bytes.Equal(eventBytes, DONE_BYTES) {
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

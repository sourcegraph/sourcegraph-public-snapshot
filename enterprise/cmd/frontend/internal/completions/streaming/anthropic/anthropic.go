package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const API_URL = "https://api.anthropic.com/v1/complete"
const CLIENT_ID = "sourcegraph/1.0"

var DONE_BYTES = []byte("[DONE]")
var STOP_SEQUENCES = []string{HUMAN_PROMPT}

type AnthropicCompletionsRequestParameters struct {
	Prompt            string   `json:"prompt"`
	Temperature       float32  `json:"temperature"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences"`
	TopK              int      `json:"top_k"`
	TopP              float32  `json:"top_p"`
	Model             string   `json:"model"`
	Stream            bool     `json:"stream"`
}

type anthropicCompletionStreamClient struct {
	cli         httpcli.Doer
	accessToken string
	model       string
}

func NewAnthropicCompletionStreamClient(cli httpcli.Doer, accessToken string, model string) types.CompletionStreamClient {
	return &anthropicCompletionStreamClient{
		cli:         cli,
		accessToken: accessToken,
		model:       model,
	}
}

func (a *anthropicCompletionStreamClient) Stream(
	ctx context.Context,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return err
	}

	payload := AnthropicCompletionsRequestParameters{
		Stream:            true,
		StopSequences:     STOP_SEQUENCES,
		Model:             a.model,
		Temperature:       requestParams.Temperature,
		MaxTokensToSample: requestParams.MaxTokensToSample,
		TopP:              requestParams.TopP,
		TopK:              requestParams.TopK,
		Prompt:            prompt,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", API_URL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	// Mimic headers set by the official Anthropic client:
	// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client", CLIENT_ID)
	req.Header.Set("X-API-Key", a.accessToken)

	resp, err := a.cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return errors.Errorf("Anthropic API failed with: %s", string(respBody))
	}

	dec := NewDecoder(resp.Body)
	for dec.Scan() {
		data := dec.Data()

		// Check for special sentinel value used by the Anthropic API to
		// indicate that the stream is done.
		if bytes.Equal(data, DONE_BYTES) {
			return nil
		}

		// Gracefully skip over any data that isn't JSON-like. Anthropic's API sometimes sends
		// non-documented data over the stream, like timestamps.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event types.CompletionEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w", err)
		}

		err = sendEvent(event)
		if err != nil {
			return err
		}
	}

	return dec.Err()
}

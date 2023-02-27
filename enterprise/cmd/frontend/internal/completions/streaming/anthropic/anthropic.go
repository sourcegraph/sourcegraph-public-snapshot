package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const API_URL = "https://api.anthropic.com/v1/complete"
const CLIENT_ID = "sourcegraph/1.0"
const ANTHROPIC_MODEL = "claude-v1"

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

func AnthropicCompletionStream(
	ctx context.Context,
	accessToken string,
	requestParams types.CompletionRequestParameters,
) (<-chan types.CompletionEvent, <-chan error, error) {
	eventsChannel := make(chan types.CompletionEvent, 8)
	errChannel := make(chan error, 1)

	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return nil, nil, err
	}

	payload := AnthropicCompletionsRequestParameters{
		Stream:            true,
		StopSequences:     STOP_SEQUENCES,
		Model:             ANTHROPIC_MODEL,
		Temperature:       requestParams.Temperature,
		MaxTokensToSample: requestParams.MaxTokensToSample,
		TopP:              requestParams.TopP,
		TopK:              requestParams.TopK,
		Prompt:            prompt,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("POST", API_URL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(ctx)

	// Mimic headers set by the official Anthropic client:
	// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client", CLIENT_ID)
	req.Header.Set("X-API-Key", accessToken)

	go func() {
		resp, err := httpcli.ExternalDoer.Do(req)
		if err != nil {
			errChannel <- err
			return
		}
		defer resp.Body.Close()
		defer close(eventsChannel)
		defer close(errChannel)

		dec := NewDecoder(resp.Body)
		for dec.Scan() {
			data := dec.Data()

			// Check for special sentinel value used by the Anthropic API to
			// indicate that the stream is done.
			if bytes.Equal(data, DONE_BYTES) {
				return
			}

			// Gracefully skip over any data that isn't JSON-like. Anthropic's API sometimes sends
			// non-documented data over the stream, like timestamps.
			if !bytes.HasPrefix(data, []byte("{")) {
				continue
			}

			var event types.CompletionEvent
			if err := json.Unmarshal(data, &event); err != nil {
				errChannel <- errors.Errorf("failed to decode event payload: %w", err)
				return
			}
			eventsChannel <- event
		}

		errChannel <- dec.Err()
	}()

	return eventsChannel, errChannel, nil
}

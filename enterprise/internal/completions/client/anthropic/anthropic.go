package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

const ProviderName = "anthropic"

func NewClient(cli httpcli.Doer, accessToken string, model string) types.CompletionsClient {
	return &anthropicClient{
		cli:         cli,
		accessToken: accessToken,
		model:       model,
	}
}

type anthropicClient struct {
	cli         httpcli.Doer
	accessToken string
	model       string
}

const apiURL = "https://api.anthropic.com/v1/complete"
const clientID = "sourcegraph/1.0"

var stopSequences = []string{HUMAN_PROMPT}

var allowedClientSpecifiedModels = map[string]struct{}{
	"claude-instant-v1.0": {},
}

func (a *anthropicClient) Complete(
	ctx context.Context,
	requestParams types.CodeCompletionRequestParameters,
) (*types.CodeCompletionResponse, error) {
	var model string
	if _, isAllowed := allowedClientSpecifiedModels[requestParams.Model]; isAllowed {
		model = requestParams.Model
	} else {
		model = a.model
	}
	payload := AnthropicCompletionsRequestParameters{
		Stream:            false,
		StopSequences:     requestParams.StopSequences,
		Model:             model,
		Temperature:       float32(requestParams.Temperature),
		MaxTokensToSample: requestParams.MaxTokensToSample,
		TopP:              float32(requestParams.TopP),
		TopK:              requestParams.TopK,
		Prompt:            requestParams.Prompt,
	}

	resp, err := a.makeRequest(ctx, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response types.CodeCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (a *anthropicClient) Stream(
	ctx context.Context,
	requestParams types.ChatCompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return err
	}

	payload := AnthropicCompletionsRequestParameters{
		Stream:            true,
		StopSequences:     stopSequences,
		Model:             a.model,
		Temperature:       requestParams.Temperature,
		MaxTokensToSample: requestParams.MaxTokensToSample,
		TopP:              requestParams.TopP,
		TopK:              requestParams.TopK,
		Prompt:            prompt,
	}

	resp, err := a.makeRequest(ctx, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := NewDecoder(resp.Body)
	for dec.Scan() {
		if ctx.Err() != nil && ctx.Err() == context.Canceled {
			return nil
		}

		data := dec.Data()
		// Gracefully skip over any data that isn't JSON-like. Anthropic's API sometimes sends
		// non-documented data over the stream, like timestamps.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event types.ChatCompletionEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w - body: %s", err, string(data))
		}

		err = sendEvent(event)
		if err != nil {
			return err
		}
	}

	return dec.Err()
}

func (a *anthropicClient) makeRequest(ctx context.Context, payload AnthropicCompletionsRequestParameters) (*http.Response, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	// Mimic headers set by the official Anthropic client:
	// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client", clientID)
	req.Header.Set("X-API-Key", a.accessToken)

	resp, err := a.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, errors.Errorf("Anthropic API failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return resp, nil
}

package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewClient(cli httpcli.Doer, apiURL, accessToken string) types.CompletionsClient {
	return &anthropicClient{
		cli:         cli,
		accessToken: accessToken,
		apiURL:      apiURL,
	}
}

const (
	clientID = "sourcegraph/1.0"
)

type anthropicClient struct {
	cli         httpcli.Doer
	accessToken string
	apiURL      string
}

func (a *anthropicClient) Complete(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	resp, err := a.makeRequest(ctx, requestParams, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response anthropicCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &types.CompletionResponse{
		Completion: response.Completion,
		StopReason: response.StopReason,
	}, nil
}

func (a *anthropicClient) Stream(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	resp, err := a.makeRequest(ctx, requestParams, true)
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

		var event anthropicCompletionResponse
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w - body: %s", err, string(data))
		}

		err = sendEvent(types.CompletionResponse{
			Completion: event.Completion,
			StopReason: event.StopReason,
		})
		if err != nil {
			return err
		}
	}

	return dec.Err()
}

func (a *anthropicClient) makeRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	prompt, err := GetPrompt(requestParams.Messages)
	if err != nil {
		return nil, err
	}
	// Backcompat: Remove this code once enough clients are upgraded and we drop the
	// Prompt field on requestParams.
	if prompt == "" {
		prompt = requestParams.Prompt
	}

	if len(requestParams.StopSequences) == 0 {
		requestParams.StopSequences = []string{HUMAN_PROMPT}
	}

	payload := anthropicCompletionsRequestParameters{
		Stream:            stream,
		StopSequences:     requestParams.StopSequences,
		Model:             requestParams.Model,
		Temperature:       requestParams.Temperature,
		MaxTokensToSample: requestParams.MaxTokensToSample,
		TopP:              requestParams.TopP,
		TopK:              requestParams.TopK,
		Prompt:            prompt,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.apiURL, bytes.NewReader(reqBody))
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
	// Set the API version so responses are in the expected format.
	// NOTE: When changing this here, Cody Gateway currently overwrites this header
	// with 2023-01-01, so it will not be respected in Gateway usage and we will
	// have to fall back to the old parser, or implement a mechanism on the Gateway
	// side that understands the version header we send here and switch out the parser.
	req.Header.Set("anthropic-version", "2023-01-01")

	resp, err := a.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("Anthropic", resp)
	}

	return resp, nil
}

type anthropicCompletionsRequestParameters struct {
	Prompt            string   `json:"prompt"`
	Temperature       float32  `json:"temperature"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences"`
	TopK              int      `json:"top_k"`
	TopP              float32  `json:"top_p"`
	Model             string   `json:"model"`
	Stream            bool     `json:"stream"`
}

type anthropicCompletionResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}

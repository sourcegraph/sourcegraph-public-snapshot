package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	messages, err := ToAnthropicMessages(requestParams.Messages)
	if err != nil {
		return nil, err
	}

	// TODO:
	// this needs backward-compat with eventually overwritten config values ending in `/v1/complete`
	fmt.Println("%+v", a.apiURL)

	payload := anthropicMessagesRequestParameters{
		Messages:      messages,
		Stream:        stream,
		StopSequences: requestParams.StopSequences,
		Model:         requestParams.Model,
		Temperature:   requestParams.Temperature,
		MaxTokens:     requestParams.MaxTokensToSample,
		TopP:          requestParams.TopP,
		TopK:          requestParams.TopK,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	// Only run this branch if we're not talking to cody gateway
	// Convert the eventual first message from `system` to a top-level system prompt
	payload.System = "" // prevent the upstream API from setting this
	if len(payload.Messages) > 0 && payload.Messages[0].Role == types.SYSTEM_MESSAGE_SPEAKER {
		payload.System = payload.Messages[0].Content[0].Text
		payload.Messages = payload.Messages[1:]
	}

	// Mimic headers set by the official Anthropic client:
	// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client", clientID)
	req.Header.Set("X-API-Key", a.accessToken)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("Anthropic", resp)
	}

	return resp, nil
}

type anthropicMessagesRequestParameters struct {
	Messages      []anthropicMessage `json:"messages,omitempty"`
	Model         string             `json:"model"`
	Temperature   float32            `json:"temperature,omitempty"`
	TopP          float32            `json:"top_p,omitempty"`
	TopK          int                `json:"top_k,omitempty"`
	Stream        bool               `json:"stream,omitempty"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	MaxTokens     int                `json:"max_tokens,omitempty"`

	// These are not accepted from the client an instead are only used to talk to the upstream LLM
	// APIs directly (these do NOT need to be set when talking to Cody Gateway)
	System string `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string                    `json:"role"` // "user", "assistant", or "system" (only allowed for the first message)
	Content []anthropicMessageContent `json:"content"`
}

type anthropicMessageContent struct {
	Type string `json:"type"` // "text" or "image" (not yet supported)
	Text string `json:"text"`
}
type anthropicCompletionResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}

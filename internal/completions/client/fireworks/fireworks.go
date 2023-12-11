package fireworks

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"net/http"
	"strings"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string) types.CompletionsClient {
	return &fireworksClient{
		cli:         cli,
		accessToken: accessToken,
		endpoint:    endpoint,
	}
}

type fireworksClient struct {
	cli         httpcli.Doer
	accessToken string
	endpoint    string
}

func (c *fireworksClient) Complete(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	// TODO: If we add support for other features, Cody Gateway must also be updated.
	if feature != types.CompletionsFeatureCode {
		return nil, errors.Newf("%q for Fireworks is currently not supported")
	}

	resp, err := c.makeRequest(ctx, requestParams, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response fireworksResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		// Empty response.
		return &types.CompletionResponse{}, nil
	}

	return &types.CompletionResponse{
		Completion: response.Choices[0].Text,
		StopReason: response.Choices[0].FinishReason,
		Logprobs:   response.Choices[0].Logprobs,
	}, nil
}

func (c *fireworksClient) Stream(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	logprobsInclude := uint8(0)
	requestParams.Logprobs = &logprobsInclude

	// HACK: Cody Gateway expects model names in <provider>/<model> format, but if we're connecting directly to
	// the Fireworks API, we need to strip the "fireworks" provider prefix
	if components := strings.Split(requestParams.Model, "/"); components[0] == "fireworks" {
		requestParams.Model = strings.Join(components[1:], "/")
	}

	resp, err := c.makeRequest(ctx, requestParams, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := NewDecoder(resp.Body)
	var content string
	var accumulatedLogprobs *types.Logprobs
	for dec.Scan() {
		if ctx.Err() != nil && ctx.Err() == context.Canceled {
			return nil
		}

		data := dec.Data()
		// Gracefully skip over any data that isn't JSON-like.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event fireworksResponse
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w - body: %s", err, string(data))
		}

		if len(event.Choices) > 0 {
			content += event.Choices[0].Text
			accumulatedLogprobs = accumulatedLogprobs.Append(event.Choices[0].Logprobs)
			ev := types.CompletionResponse{
				Completion: content,
				StopReason: event.Choices[0].FinishReason,
				Logprobs:   accumulatedLogprobs,
			}
			err = sendEvent(ev)
			if err != nil {
				return err
			}
		}
	}

	return dec.Err()
}

func (c *fireworksClient) makeRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	// For compatibility reasons with other models, we expect to find the prompt
	// in the first and only message
	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return nil, err
	}

	payload := fireworksRequest{
		Model:       requestParams.Model,
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		N:           1,
		Stream:      stream,
		MaxTokens:   int32(requestParams.MaxTokensToSample),
		Stop:        requestParams.StopSequences,
		Echo:        false,
		Prompt:      prompt,
		Logprobs:    requestParams.Logprobs,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("Fireworks", resp)
	}

	return resp, nil
}

// fireworksRequest captures all known fields from https://fireworksai.readme.io/reference/createcompletion.
type fireworksRequest struct {
	Prompt      string   `json:"prompt"`
	Model       string   `json:"model"`
	MaxTokens   int32    `json:"max_tokens,omitempty"`
	Temperature float32  `json:"temperature,omitempty"`
	TopP        float32  `json:"top_p,omitempty"`
	N           int32    `json:"n,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
	Echo        bool     `json:"echo,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	Logprobs    *uint8   `json:"logprobs,omitempty"`
}

// response for a non streaming request
type fireworksResponse struct {
	Choices []struct {
		Text         string          `json:"text"`
		Index        int             `json:"index"`
		FinishReason string          `json:"finish_reason"`
		Logprobs     *types.Logprobs `json:"logprobs"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		TotalTokens      int `json:"total_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

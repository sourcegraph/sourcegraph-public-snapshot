package azureopenai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string) types.CompletionsClient {
	return &azureCompletionClient{
		cli:         cli,
		accessToken: accessToken,
		endpoint:    endpoint,
	}
}

type azureCompletionClient struct {
	cli         httpcli.Doer
	accessToken string
	endpoint    string
}

func (c *azureCompletionClient) Complete(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	var resp *http.Response
	var err error
	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()
	if feature == types.CompletionsFeatureCode {
		resp, err = c.makeCompletionRequest(ctx, requestParams, false)
	} else {
		resp, err = c.makeRequest(ctx, requestParams, false)
	}
	if err != nil {
		return nil, err
	}

	var response openaiResponse
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
	}, nil
}

func (c *azureCompletionClient) Stream(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	var resp *http.Response
	var err error

	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()
	if feature == types.CompletionsFeatureCode {
		resp, err = c.makeCompletionRequest(ctx, requestParams, true)
	} else {
		resp, err = c.makeRequest(ctx, requestParams, true)
	}
	if err != nil {
		return err
	}

	dec := openai.NewDecoder(resp.Body)
	var content string
	for dec.Scan() {
		if ctx.Err() != nil && ctx.Err() == context.Canceled {
			return nil
		}

		data := dec.Data()
		// Gracefully skip over any data that isn't JSON-like.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event openaiResponse
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w - body: %s", err, string(data))
		}

		if len(event.Choices) > 0 {
			if feature == types.CompletionsFeatureCode {
				content += event.Choices[0].Text
			} else {
				content += event.Choices[0].Delta.Content
			}
			ev := types.CompletionResponse{
				Completion: content,
				StopReason: event.Choices[0].FinishReason,
			}
			err = sendEvent(ev)
			if err != nil {
				return err
			}
		}
	}

	return dec.Err()
}

func (c *azureCompletionClient) makeRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	payload := azureChatCompletionsRequestParameters{
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		N:           1,
		Stream:      stream,
		MaxTokens:   requestParams.MaxTokensToSample,
		Stop:        requestParams.StopSequences,
	}
	for _, m := range requestParams.Messages {
		var role string
		switch m.Speaker {
		case types.HUMAN_MESSAGE_SPEAKER:
			role = "user"
		case types.ASISSTANT_MESSAGE_SPEAKER:
			role = "assistant"
		default:
			role = strings.ToLower(role)
		}
		payload.Messages = append(payload.Messages, message{
			Role:    role,
			Content: m.Text,
		})
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configured endpoint")
	}
	q := url.Query()
	q.Add("api-version", "2023-05-15")
	url.RawQuery = q.Encode()
	url.Path = fmt.Sprintf("/openai/deployments/%s/chat/completions", requestParams.Model)

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.accessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("AzureOpenAI", resp)
	}

	return resp, nil
}

func (c *azureCompletionClient) makeCompletionRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return nil, err
	}

	payload := azureCompletionsRequestParameters{
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		N:           1,
		Stream:      stream,
		MaxTokens:   requestParams.MaxTokensToSample,
		Stop:        requestParams.StopSequences,
		Prompt:      prompt,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configured endpoint")
	}
	q := url.Query()
	q.Add("api-version", "2023-05-15")
	url.RawQuery = q.Encode()
	url.Path = fmt.Sprintf("/openai/deployments/%s/completions", requestParams.Model)

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.accessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("AzureOpenAI", resp)
	}

	return resp, nil
}

type azureChatCompletionsRequestParameters struct {
	Messages         []message          `json:"messages"`
	Temperature      float32            `json:"temperature,omitempty"`
	TopP             float32            `json:"top_p,omitempty"`
	N                int                `json:"n,omitempty"`
	Stream           bool               `json:"stream,omitempty"`
	Stop             []string           `json:"stop,omitempty"`
	MaxTokens        int                `json:"max_tokens,omitempty"`
	PresencePenalty  float32            `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32            `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float32 `json:"logit_bias,omitempty"`
	User             string             `json:"user,omitempty"`
}

type azureCompletionsRequestParameters struct {
	Prompt           string             `json:"prompt"`
	Temperature      float32            `json:"temperature,omitempty"`
	TopP             float32            `json:"top_p,omitempty"`
	N                int                `json:"n,omitempty"`
	Stream           bool               `json:"stream,omitempty"`
	Stop             []string           `json:"stop,omitempty"`
	MaxTokens        int                `json:"max_tokens,omitempty"`
	PresencePenalty  float32            `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32            `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float32 `json:"logit_bias,omitempty"`
	Suffix           string             `json:"suffix,omitempty"`
	User             string             `json:"user,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type azure struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openaiChoiceDelta struct {
	Content string `json:"content"`
}

type openaiChoice struct {
	Delta        openaiChoiceDelta `json:"delta"`
	Role         string            `json:"role"`
	Text         string            `json:"text"`
	FinishReason string            `json:"finish_reason"`
}

type openaiResponse struct {
	// Usage is only available for non-streaming requests.
	Usage   azure          `json:"usage"`
	Model   string         `json:"model"`
	Choices []openaiChoice `json:"choices"`
}

func getPrompt(messages []types.Message) (string, error) {
	if len(messages) != 1 {
		return "", errors.New("Expected to receive exactly one message with the prompt")
	}

	return messages[0].Text, nil
}

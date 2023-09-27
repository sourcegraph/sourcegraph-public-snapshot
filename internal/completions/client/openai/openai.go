package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string) types.CompletionsClient {
	return &openAIChatCompletionStreamClient{
		cli:         cli,
		accessToken: accessToken,
		endpoint:    endpoint,
	}
}

type openAIChatCompletionStreamClient struct {
	cli         httpcli.Doer
	accessToken string
	endpoint    string
}

func (c *openAIChatCompletionStreamClient) Complete(
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
	defer resp.Body.Close()

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

func (c *openAIChatCompletionStreamClient) Stream(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	resp, err := c.makeRequest(ctx, requestParams, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := NewDecoder(resp.Body)
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
			content += event.Choices[0].Delta.Content
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

func (c *openAIChatCompletionStreamClient) makeRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	// TODO(sqs): make CompletionRequestParameters non-anthropic-specific
	payload := openAIChatCompletionsRequestParameters{
		Model:       requestParams.Model,
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		// TODO(sqs): map requestParams.TopK to openai
		N:         1,
		Stream:    stream,
		MaxTokens: requestParams.MaxTokensToSample,
		// TODO: Our clients are currently heavily biased towards Anthropic,
		// so the stop sequences we send might not actually be very useful
		// for OpenAI.
		Stop: requestParams.StopSequences,
	}
	for _, m := range requestParams.Messages {
		// TODO(sqs): map these 'roles' to openai system/user/assistant
		var role string
		switch m.Speaker {
		case types.HUMAN_MESSAGE_SPEAKER:
			role = "user"
		case types.ASISSTANT_MESSAGE_SPEAKER:
			role = "assistant"
			//
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
	url.Path = "v1/chat/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
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
		return nil, types.NewErrStatusNotOK("OpenAI", resp)
	}

	return resp, nil
}

func (c *openAIChatCompletionStreamClient) makeCompletionRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
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

	payload := openAICompletionsRequestParameters{
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
	url.Path = "v1/completions"

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

type openAIChatCompletionsRequestParameters struct {
	Model            string             `json:"model"`
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

type openAICompletionsRequestParameters struct {
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

type openaiUsage struct {
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
	Usage   openaiUsage    `json:"usage"`
	Model   string         `json:"model"`
	Choices []openaiChoice `json:"choices"`
}

func getPrompt(messages []types.Message) (string, error) {
	if len(messages) != 1 {
		return "", errors.New("Expected to receive exactly one message with the prompt")
	}

	return messages[0].Text, nil
}

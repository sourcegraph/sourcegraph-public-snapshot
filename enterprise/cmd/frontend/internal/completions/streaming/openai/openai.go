package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const apiURL = "https://api.openai.com/v1/chat/completions"

var doneBytes = []byte("[DONE]")

type OpenAIChatCompletionsRequestParameters struct {
	Model            string             `json:"model"`
	Messages         []Message          `json:"messages"`
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

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatCompletionStreamClient struct {
	cli         httpcli.Doer
	accessToken string
	model       string
}

func NewOpenAIClient(cli httpcli.Doer, accessToken string, model string) types.CompletionsClient {
	return &openAIChatCompletionStreamClient{
		cli:         cli,
		accessToken: accessToken,
		model:       model,
	}
}

func (a *openAIChatCompletionStreamClient) Complete(ctx context.Context, requestParams types.CodeCompletionRequestParameters) (*types.CodeCompletionResponse, error) {
	return nil, errors.New("openAIChatCompletionStreamClient.Complete: unimplemented")
}

func (a *openAIChatCompletionStreamClient) Stream(
	ctx context.Context,
	requestParams types.ChatCompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	// TODO(sqs): make CompletionRequestParameters non-anthropic-specific
	payload := OpenAIChatCompletionsRequestParameters{
		Model:       a.model,
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		// TODO(sqs): map requestParams.TopK to openai
		N:         1,
		Stream:    true,
		MaxTokens: requestParams.MaxTokensToSample,
	}
	for _, m := range requestParams.Messages {
		// TODO(sqs): map these 'roles' to openai system/user/assistant
		var role string
		switch m.Speaker {
		case types.HUMAN_MESSAGE_SPEAKER:
			role = "user"
		case types.ASISSTANT_MESSAGE_SPEAKER:
			role = "assistant"
		default:
			role = strings.ToLower(role)
		}
		payload.Messages = append(payload.Messages, Message{
			Role:    role,
			Content: m.Text,
		})
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	resp, err := a.cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return errors.Errorf("OpenAI API failed with: %s", string(respBody))
	}

	dec := NewDecoder(resp.Body)
	var content []string
	for dec.Scan() {
		data := dec.Data()

		if bytes.Equal(data, doneBytes) {
			return nil
		}

		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w", err)
		}

		if len(event.Choices) > 0 && event.Choices[0].FinishReason == nil {
			content = append(content, event.Choices[0].Delta.Content)
			err = sendEvent(types.ChatCompletionEvent{Completion: strings.Join(content, "")})
			if err != nil {
				return err
			}
		}
	}

	return dec.Err()
}

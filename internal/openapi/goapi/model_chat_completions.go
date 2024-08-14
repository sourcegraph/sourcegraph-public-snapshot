package goapi

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// The types in this file are generated from the OpenAI API spec:
// https://github.com/openai/openai-openapi/blob/master/openapi.yaml
// We intentionally don't customize the types because the main purpose
// of this endpoint is to be compatible with OpenAI clients.
// The structs and fields have no docs in this file but you can find the
// descriptions in the OpenAPI spec. The goal is to document this API properly
// on sourcegraph.com/docs using the same descriptions as in the OpenAPI spec.

type CreateChatCompletionRequest struct {
	Messages []ChatCompletionRequestMessage `json:"messages"`
	// IMPORTANT: we only accept the ModelRef syntax here, see internal/modelconfig/types/refs.go
	Model            string                                  `json:"model"`
	FrequencyPenalty *float64                                `json:"frequency_penalty,omitempty"`
	LogitBias        *map[string]int                         `json:"logit_bias,omitempty"`
	Logprobs         *bool                                   `json:"logprobs,omitempty"`
	TopLogprobs      *int                                    `json:"top_logprobs,omitempty"`
	MaxTokens        *int                                    `json:"max_tokens,omitempty"`
	N                *int                                    `json:"n,omitempty"`
	PresencePenalty  *float64                                `json:"presence_penalty,omitempty"`
	ResponseFormat   *ResponseFormat                         `json:"response_format,omitempty"`
	Seed             *int64                                  `json:"seed,omitempty"`
	ServiceTier      *string                                 `json:"service_tier,omitempty"`
	Stop             CreateChatCompletionRequestStopProperty `json:"stop,omitempty"`
	Stream           *bool                                   `json:"stream,omitempty"`
	StreamOptions    *ChatCompletionStreamOptions            `json:"stream_options,omitempty"`
	Temperature      *float32                                `json:"temperature,omitempty"`
	TopP             *float32                                `json:"top_p,omitempty"`
	User             *string                                 `json:"user,omitempty"`
}

// CreateChatCompletionRequestStopProperty is equivalent to the type `string | string[]` in TypeScript.
type CreateChatCompletionRequestStopProperty struct {
	Stop []string `json:"stop,omitempty"`
}

func (c *CreateChatCompletionRequestStopProperty) UnmarshalJSON(data []byte) error {
	var raw interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	switch v := raw.(type) {
	case string:
		c.Stop = []string{v}
	case []interface{}:
		var stops []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				stops = append(stops, s)
			} else {
				return errors.Newf("invalid type for stop item. Expected string, got %T", item)
			}
		}
		c.Stop = stops
	default:
		return errors.Newf("invalid type for stop. Expected string or []interface{}, got %T", v)
	}

	return nil

}

type ChatCompletionRequestMessage struct {
	Content ChatCompletionRequestContentProperty `json:"content"`
	Role    string                               `json:"role"`
	Name    *string                              `json:"name,omitempty"`
}

// ChatCompletionRequestContentProperty is a property of a message that can be a
// string or an array of ChatCompletionRequestMessageContentPartText When it's a
// string, it's a single part message where the string is the `Text` property of
// the part.
type ChatCompletionRequestContentProperty struct {
	Parts []ChatCompletionRequestMessageContentPartText `json:"parts,omitempty"`
}

type ChatCompletionRequestMessageContentPartText struct {
	Type string `json:"type" enum:"text"`
	Text string `json:"text"`
}

func (c *ChatCompletionRequestContentProperty) UnmarshalJSON(data []byte) error {
	var raw interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	switch v := raw.(type) {
	case string:
		c.Parts = []ChatCompletionRequestMessageContentPartText{
			{
				Type: "text",
				Text: v,
			},
		}
	case []interface{}:
		var parts []ChatCompletionRequestMessageContentPartText
		err = json.Unmarshal(data, &parts)
		if err != nil {
			return err
		}
		c.Parts = parts
	default:
		return errors.Newf("invalid type for ChatCompletionRequestContent. Expected string or []interface{}, got %T", v)
	}

	return nil
}

type ResponseFormat struct {
	Type ResponseFormatType `json:"type"`
}

type ResponseFormatType string

const (
	ResponseFormatTypeText       ResponseFormatType = "text"
	ResponseFormatTypeJSONObject ResponseFormatType = "json_object"
)

func (r *ResponseFormatType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch ResponseFormatType(s) {
	case ResponseFormatTypeText, ResponseFormatTypeJSONObject:
		*r = ResponseFormatType(s)
		return nil
	default:
		return errors.Newf("invalid ResponseFormatType: %s", s)
	}
}

type ChatCompletionStreamOptions struct {
	IncludeUsage *bool `json:"include_usage,omitempty"`
}

type CreateChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	ServiceTier       *string                `json:"service_tier,omitempty"`
	SystemFingerprint string                 `json:"system_fingerprint"`
	Object            string                 `json:"object"`
	Usage             CompletionUsage        `json:"usage"`
}

type ChatCompletionChoice struct {
	FinishReason string                        `json:"finish_reason"`
	Index        int                           `json:"index"`
	Message      ChatCompletionResponseMessage `json:"message"`
	Logprobs     *ChatCompletionLogprobs       `json:"logprobs,omitempty"`
}

type ChatCompletionResponseMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type ChatCompletionLogprobs struct {
	Content []ChatCompletionTokenLogprob `json:"content"`
}

type ChatCompletionTokenLogprob struct {
	Token       string                  `json:"token"`
	Logprob     float64                 `json:"logprob"`
	Bytes       []int                   `json:"bytes"`
	TopLogprobs []ChatCompletionLogprob `json:"top_logprobs"`
}

type ChatCompletionLogprob struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
	Bytes   []int   `json:"bytes"`
}

type CompletionUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

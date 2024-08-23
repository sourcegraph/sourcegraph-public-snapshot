//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package azopenai

import (
	"encoding/json"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// Models for methods that return streaming response

// GetCompletionsStreamOptions contains the optional parameters for the [Client.GetCompletionsStream] method.
type GetCompletionsStreamOptions struct {
	// placeholder for future optional parameters
}

// GetCompletionsStreamResponse is the response from [Client.GetCompletionsStream].
type GetCompletionsStreamResponse struct {
	// CompletionsStream returns the stream of completions. Token limits and other settings may limit the number of completions returned by the service.
	CompletionsStream *EventReader[Completions]
}

// GetChatCompletionsStreamOptions contains the optional parameters for the [Client.GetChatCompletionsStream] method.
type GetChatCompletionsStreamOptions struct {
	// placeholder for future optional parameters
}

// GetChatCompletionsStreamResponse is the response from [Client.GetChatCompletionsStream].
type GetChatCompletionsStreamResponse struct {
	// ChatCompletionsStream returns the stream of completions. Token limits and other settings may limit the number of chat completions returned by the service.
	ChatCompletionsStream *EventReader[ChatCompletions]
}

// ImageGenerationsDataItem contains the results of image generation.
//
// The field that's set will be based on [ImageGenerationOptions.ResponseFormat] and
// are mutually exclusive.
type ImageGenerationsDataItem struct {
	// Base64Data is set to image data, encoded as a base64 string, if [ImageGenerationOptions.ResponseFormat]
	// was set to [ImageGenerationResponseFormatB64JSON].
	Base64Data *string `json:"b64_json"`

	// URL is the address of a generated image if [ImageGenerationOptions.ResponseFormat] was set
	// to [ImageGenerationResponseFormatURL].
	URL *string `json:"url"`
}

// ContentFilterResponseError is an error as a result of a request being filtered.
type ContentFilterResponseError struct {
	azcore.ResponseError

	// ContentFilterResults contains Information about the content filtering category, if it has been detected.
	ContentFilterResults *ContentFilterResults
}

// ContentFilterResults are the content filtering results for a [ContentFilterResponseError].
type ContentFilterResults struct {
	// Describes language attacks or uses that include pejorative or discriminatory language with reference to a person or identity
	// group on the basis of certain differentiating attributes of these groups
	// including but not limited to race, ethnicity, nationality, gender identity and expression, sexual orientation, religion,
	// immigration status, ability status, personal appearance, and body size.
	Hate *ContentFilterResult `json:"hate"`

	// Describes language related to physical actions intended to purposely hurt, injure, or damage one’s body, or kill oneself.
	SelfHarm *ContentFilterResult `json:"self_harm"`

	// Describes language related to anatomical organs and genitals, romantic relationships, acts portrayed in erotic or affectionate
	// terms, physical sexual acts, including those portrayed as an assault or a
	// forced sexual violent act against one’s will, prostitution, pornography, and abuse.
	Sexual *ContentFilterResult `json:"sexual"`

	// Describes language related to physical actions intended to hurt, injure, damage, or kill someone or something; describes
	// weapons, etc.
	Violence *ContentFilterResult `json:"violence"`
}

// Unwrap returns the inner error for this error.
func (e *ContentFilterResponseError) Unwrap() error {
	return &e.ResponseError
}

func newContentFilterResponseError(resp *http.Response) error {
	respErr := runtime.NewResponseError(resp).(*azcore.ResponseError)

	if respErr.ErrorCode != "content_filter" {
		return respErr
	}

	body, err := runtime.Payload(resp)

	if err != nil {
		return err
	}

	var envelope *struct {
		Error struct {
			InnerError struct {
				ContentFilterResults *ContentFilterResults `json:"content_filter_result"`
			} `json:"innererror"`
		}
	}

	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}

	return &ContentFilterResponseError{
		ResponseError:        *respErr,
		ContentFilterResults: envelope.Error.InnerError.ContentFilterResults,
	}
}

// AzureChatExtensionOptions provides Azure specific options to extend ChatCompletions.
type AzureChatExtensionOptions struct {
	// Extensions is a slice of extensions to the chat completions endpoint, like Azure Cognitive Search.
	Extensions []AzureChatExtensionConfiguration
}

// Error implements the error interface for type Error.
// Note that the message contents are not contractual and can change over time.
func (e *Error) Error() string {
	if e.message == nil {
		return ""
	}

	return *e.message
}

// ChatCompletionsToolChoice controls which tool is used for this ChatCompletions call.
// You can choose between:
// - [ChatCompletionsToolChoiceAuto] means the model can pick between generating a message or calling a function.
// - [ChatCompletionsToolChoiceNone] means the model will not call a function and instead generates a message
// - Use the [NewChatCompletionsToolChoice] function to specify a specific tool.
type ChatCompletionsToolChoice struct {
	value any
}

var (
	// ChatCompletionsToolChoiceAuto means the model can pick between generating a message or calling a function.
	ChatCompletionsToolChoiceAuto *ChatCompletionsToolChoice = &ChatCompletionsToolChoice{value: "auto"}

	// ChatCompletionsToolChoiceNone means the model will not call a function and instead generates a message.
	ChatCompletionsToolChoiceNone *ChatCompletionsToolChoice = &ChatCompletionsToolChoice{value: "none"}
)

// NewChatCompletionsToolChoice creates a ChatCompletionsToolChoice for a specific tool.
func NewChatCompletionsToolChoice[T ChatCompletionsToolChoiceFunction](v T) *ChatCompletionsToolChoice {
	return &ChatCompletionsToolChoice{value: v}
}

// ChatCompletionsToolChoiceFunction can be used to force the model to call a particular function.
type ChatCompletionsToolChoiceFunction struct {
	// Name is the name of the function to call.
	Name string
}

// MarshalJSON implements the json.Marshaller interface for type ChatCompletionsToolChoiceFunction.
func (tf ChatCompletionsToolChoiceFunction) MarshalJSON() ([]byte, error) {
	type jsonInnerFunc struct {
		Name string `json:"name"`
	}

	type jsonFormat struct {
		Type     string        `json:"type"`
		Function jsonInnerFunc `json:"function"`
	}

	return json.Marshal(jsonFormat{
		Type:     "function",
		Function: jsonInnerFunc(tf),
	})
}

// MarshalJSON implements the json.Marshaller interface for type ChatCompletionsToolChoice.
func (tc ChatCompletionsToolChoice) MarshalJSON() ([]byte, error) {
	return json.Marshal(tc.value)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ChatCompletionsToolChoice.
func (tc *ChatCompletionsToolChoice) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &tc.value)
}

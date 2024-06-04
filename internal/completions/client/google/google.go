package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string) types.CompletionsClient {
	return &googleCompletionStreamClient{
		cli:         cli,
		accessToken: accessToken,
		endpoint:    endpoint,
	}
}

type googleCompletionStreamClient struct {
	cli         httpcli.Doer
	accessToken string
	endpoint    string
}

func (c *googleCompletionStreamClient) Complete(
	ctx context.Context,
	feature types.CompletionsFeature,
	_ types.CompletionsVersion,
	requestParams types.CompletionRequestParameters,
	logger log.Logger,
) (*types.CompletionResponse, error) {
	var resp *http.Response
	var err error
	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()

	if feature == types.CompletionsFeatureCode {
		return nil, errors.Newf("feature %q is currently not supported for Google", feature)
	}

	resp, err = c.makeRequest(ctx, requestParams, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response googleResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Candidates) == 0 {
		// Empty response.
		return &types.CompletionResponse{}, nil
	}

	if len(response.Candidates[0].Content.Parts) == 0 {
		// Empty response.
		return &types.CompletionResponse{}, nil
	}

	return &types.CompletionResponse{
		Completion: response.Candidates[0].Content.Parts[0].Text,
		StopReason: response.Candidates[0].FinishReason,
	}, nil
}

func (c *googleCompletionStreamClient) Stream(
	ctx context.Context,
	feature types.CompletionsFeature,
	_ types.CompletionsVersion,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
	logger log.Logger,
) error {
	var resp *http.Response
	var err error

	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()

	if feature == types.CompletionsFeatureCode {
		return errors.Newf("feature %q is currently not supported for Google", feature)
	}

	resp, err = c.makeRequest(ctx, requestParams, true)
	if err != nil {
		return err
	}

	dec := NewDecoder(resp.Body)
	var content string
	var ev types.CompletionResponse

	for dec.Scan() {
		if ctx.Err() != nil && ctx.Err() == context.Canceled {
			return nil
		}

		data := dec.Data()
		// Gracefully skip over any data that isn't JSON-like.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event googleResponse
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w - body: %s", err, string(data))
		}

		if len(event.Candidates) > 0 && len(event.Candidates[0].Content.Parts) > 0 {
			content += event.Candidates[0].Content.Parts[0].Text

			ev = types.CompletionResponse{
				Completion: content,
				StopReason: event.Candidates[0].FinishReason,
			}
			err = sendEvent(ev)
			if err != nil {
				return err
			}
		}
	}
	if dec.Err() != nil {
		return dec.Err()
	}

	return nil
}

// In the latest API Docs, the model name and API key must be used in the API endpoint URL.
// Ref: https://ai.google.dev/gemini-api/docs/get-started/tutorial?lang=rest#gemini_and_content_based_apis
func (c *googleCompletionStreamClient) getAPIURL(requestParams types.CompletionRequestParameters, stream bool) string {
	rpc := "generateContent"
	sseSuffix := ""

	if stream {
		rpc = "streamGenerateContent"
		sseSuffix = "&alt=sse"
	}

	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:%s?key=%s%s", requestParams.Model, rpc, c.accessToken, sseSuffix)
}

// makeRequest formats the request and calls the chat/completions endpoint for code_completion requests
func (c *googleCompletionStreamClient) makeRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	// Ensure TopK and TopP are non-negative
	requestParams.TopK = max(0, requestParams.TopK)
	requestParams.TopP = max(0, requestParams.TopP)

	// Generate the prompt
	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return nil, err
	}

	payload := googleRequest{
		Contents: prompt,
		GenerationConfig: googleGenerationConfig{
			Temperature:     requestParams.Temperature,
			TopP:            requestParams.TopP,
			TopK:            requestParams.TopK,
			MaxOutputTokens: requestParams.MaxTokensToSample,
			StopSequences:   requestParams.StopSequences,
		},
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	apiURL := c.getAPIURL(requestParams, stream)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("Google", resp)
	}

	return resp, nil
}

type googleRequest struct {
	Contents         []googleContentMessage `json:"contents"`
	GenerationConfig googleGenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []googleSafetySettings `json:"safetySettings,omitempty"`
}

type googleContentMessage struct {
	Role  string                     `json:"role"`
	Parts []googleContentMessagePart `json:"parts"`
}

type googleContentMessagePart struct {
	Text string `json:"text"`
}

// Configuration options for model generation and outputs.
// Ref: https://ai.google.dev/api/rest/v1/GenerationConfig
type googleGenerationConfig struct {
	Temperature     float32  `json:"temperature,omitempty"`     // request.Temperature
	TopP            float32  `json:"topP,omitempty"`            // request.TopP
	TopK            int      `json:"topK,omitempty"`            // request.TopK
	StopSequences   []string `json:"stopSequences,omitempty"`   // request.StopSequences
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"` // request.MaxTokensToSample
	CandidateCount  int      `json:"candidateCount,omitempty"`  // request.CandidateCount
}

type googleResponse struct {
	Model      string `json:"model"`
	Candidates []struct {
		Content      googleContentMessage
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`

	UsageMetadata  googleUsage            `json:"usageMetadata"`
	SafetySettings []googleSafetySettings `json:"safetySettings,omitempty"`
}

// Safety setting, affecting the safety-blocking behavior.
// Ref: https://ai.google.dev/gemini-api/docs/safety-settings
type googleSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type googleUsage struct {
	PromptTokenCount int `json:"promptTokenCount"`
	// Use the same name we use elsewhere (completion instead of candidates)
	CompletionTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

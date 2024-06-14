package google

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

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

	// NOTE:  Candidates can be used to get multiple completions when CandidateCount is set,
	// which is not currently supported by Cody. For now, we only return the first completion.
	return &types.CompletionResponse{
		Completion: response.Candidates[0].Content.Parts[0].Text,
		StopReason: response.Candidates[0].StopReason,
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
				StopReason: event.Candidates[0].StopReason,
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
		Model:    requestParams.Model,
		Stream:   stream,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Vertex AI API requires an Authorization header with the access token.
	// Ref: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/gemini#sample-requests
	if !isDefaultAPIEndpoint(apiURL) {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("Google", resp)
	}

	return resp, nil
}

// In the latest API Docs, the model name and API key must be used with the default API endpoint URL.
// Ref: https://ai.google.dev/gemini-api/docs/get-started/tutorial?lang=rest#gemini_and_content_based_apis
func (c *googleCompletionStreamClient) getAPIURL(requestParams types.CompletionRequestParameters, stream bool) *url.URL {
	apiURL, err := url.Parse(c.endpoint)
	if err != nil {
		apiURL = &url.URL{
			Scheme: "https",
			Host:   defaultAPIHost,
			Path:   defaultAPIPath,
		}
	}

	apiURL.Path = path.Join(apiURL.Path, requestParams.Model) + ":" + getgRPCMethod(stream)

	// We need to append the API key to the default API endpoint URL.
	if isDefaultAPIEndpoint(apiURL) {
		query := apiURL.Query()
		query.Set("key", c.accessToken)
		if stream {
			query.Set("alt", "sse")
		}
		apiURL.RawQuery = query.Encode()
	}

	return apiURL
}

// getgRPCMethod returns the gRPC method name based on the stream flag.
func getgRPCMethod(stream bool) string {
	if stream {
		return "streamGenerateContent"
	}
	return "generateContent"
}

// isDefaultAPIEndpoint checks if the given API endpoint URL is the default API endpoint.
// The default API endpoint is determined by the defaultAPIHost constant.
func isDefaultAPIEndpoint(endpoint *url.URL) bool {
	return endpoint.Host == defaultAPIHost
}

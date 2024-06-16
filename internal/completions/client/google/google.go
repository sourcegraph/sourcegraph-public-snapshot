package google

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/sourcegraph/log"
	"golang.org/x/oauth2/google"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	Gemini    ModelFamily = "gemini"
	Anthropic ModelFamily = "anthropic"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string) types.CompletionsClient {
	modelFamily := Gemini
	if strings.Contains(endpoint, "anthropic") {
		modelFamily = Anthropic
	}
	return &googleCompletionStreamClient{
		cli:         cli,
		accessToken: accessToken,
		endpoint:    endpoint,
		modelFamily: modelFamily,
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

	resp, err = c.makeAnthopicRequest(ctx, requestParams, false)
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

	// NOTE: Candidates can be used to get multiple completions when CandidateCount is set,
	// which is not currently supported by Cody. For now, we only return the first completion.
	return &types.CompletionResponse{
		Completion: response.Candidates[0].Content.Parts[0].Text,
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

	if c.modelFamily == Anthropic {
		resp, err = c.makeAnthopicRequest(ctx, requestParams, true)
		if err != nil {
			return err
		}

		reader := bufio.NewReader(resp.Body)
		var (
			event             []byte
			emptyMessageCount uint
			totalCompletion   string
			sentEvent         bool
		)
		for {
			rawLine, readErr := reader.ReadBytes('\n')
			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					break
				}
				return readErr
			}

			noSpaceLine := bytes.TrimSpace(rawLine)
			if len(noSpaceLine) == 0 {
				continue
			}

			if bytes.HasPrefix(noSpaceLine, []byte("event:")) {
				event = bytes.TrimSpace(bytes.TrimPrefix(noSpaceLine, []byte("event:")))
				continue
			}

			if bytes.HasPrefix(noSpaceLine, []byte("data:")) {
				data := bytes.TrimPrefix(noSpaceLine, []byte("data:"))
				eventType := string(event)

				switch eventType {
				case "message_start":
					// Handle message_start event
					var d anthropicStreamingResponse
					if err := json.Unmarshal(data, &d); err != nil {
						return err
					}
					// Process message_start event if needed
					continue

				case "content_block_delta":
					// Handle content_block_delta event
					var d anthropicStreamingResponse
					if err := json.Unmarshal(data, &d); err != nil {
						return err
					}
					totalCompletion += d.Delta.Text
					sentEvent = true
					err = sendEvent(types.CompletionResponse{
						Completion: totalCompletion,
					})
					if err != nil {
						return err
					}
					continue
				case "message_delta":
					// Handle message_delta event
					var d anthropicStreamingResponseTextBucket
					if err := json.Unmarshal(data, &d); err != nil {
						return err
					}
					// Process message_delta event if needed
					continue

				case "message_stop":
					// Handle message_stop event
					// Process message_stop event if needed
					continue

				default:
					// Handle other events if needed
					continue
				}
			}

			emptyMessageCount++
			if emptyMessageCount > 100 { // Adjust the limit as needed
				return errors.New("too many empty stream messages")
			}
		}

		if !sentEvent {
			return errors.New("stream closed with no events")
		}

		return nil
	} else {
		resp, err = c.makeGeminiRequest(ctx, requestParams, true)
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
}

// makeRequest formats the request and calls the chat/completions endpoint for code_completion requests
func (c *googleCompletionStreamClient) makeGeminiRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	// Ensure TopK and TopP are non-negative
	requestParams.TopK = max(0, requestParams.TopK)
	requestParams.TopP = max(0, requestParams.TopP)

	// Generate the prompt
	prompt, err := getGeminiPrompt(requestParams.Messages)
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

// DecodeBase64 decodes a base64 string
func decodeBase64(encoded string) ([]byte, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return decodedBytes, nil
}

// makeRequest formats the request and calls the chat/completions endpoint for code_completion requests
func (c *googleCompletionStreamClient) makeAnthopicRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	// Generate the prompt
	prompt, err := getAnthropicPrompt(requestParams.Messages)
	if err != nil {
		return nil, err
	}

	payload := anthropicRequest{
		Messages:         prompt,
		MaxTokens:        requestParams.MaxTokensToSample,
		Stream:           true,
		AnthropicVersion: "vertex-2023-10-16",
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	apiURL := c.getAPIURL(requestParams, stream)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	serviceAccountInfo, err := decodeBase64(c.accessToken)
	if err != nil {
		return nil, err
	}

	scopes := "https://www.googleapis.com/auth/cloud-platform"
	creds, err := google.CredentialsFromJSON(ctx, serviceAccountInfo, scopes)
	if err != nil {
		return nil, err
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	// Vertex AI API requires an Authorization header with the access token.
	// Ref: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/gemini#sample-requests
	// TODO: Weird Oauth2 thingy here
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

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

	apiURL.Path = path.Join(apiURL.Path, requestParams.Model) + ":" + getgRPCMethod(stream, c.modelFamily)

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
func getgRPCMethod(stream bool, modelFamily ModelFamily) string {
	if stream {
		if modelFamily == Anthropic {
			return "streamRawPredict"
		}
		return "streamGenerateContent"
	}
	return "generateContent"
}

// isDefaultAPIEndpoint checks if the given API endpoint URL is the default API endpoint.
// The default API endpoint is determined by the defaultAPIHost constant.
func isDefaultAPIEndpoint(endpoint *url.URL) bool {
	return endpoint.Host == defaultAPIHost
}

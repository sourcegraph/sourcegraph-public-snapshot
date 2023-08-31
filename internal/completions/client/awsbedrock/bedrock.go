package awsbedrock

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/sourcegraph/sourcegraph/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string) types.CompletionsClient {
	return &awsBedrockAnthropicCompletionStreamClient{
		cli:         cli,
		accessToken: accessToken,
		endpoint:    endpoint,
	}
}

const (
	clientID = "sourcegraph/1.0"
)

type awsBedrockAnthropicCompletionStreamClient struct {
	cli         httpcli.Doer
	accessToken string
	endpoint    string
}

func awsConfigOptsForKeyConfig(endpoint string, accessToken string) []func(*config.LoadOptions) error {
	configOpts := []func(*config.LoadOptions) error{}
	if endpoint != "" {
		configOpts = append(configOpts, config.WithRegion(endpoint))
	}
	if accessToken != "" {
		parts := strings.SplitN(accessToken, ":", 3)
		if len(parts) == 2 {
			configOpts = append(configOpts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(parts[0], parts[1], "")))
		} else if len(parts) == 3 {
			configOpts = append(configOpts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(parts[0], parts[1], parts[2])))
		}
	}
	return configOpts
}

func (c *awsBedrockAnthropicCompletionStreamClient) Complete(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	resp, err := c.makeRequest(ctx, requestParams, false)
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

func (a *awsBedrockAnthropicCompletionStreamClient) Stream(
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

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("unexpected status code from API: %d", resp.StatusCode)
	}

	var totalCompletion string
	deco := eventstream.NewDecoder()
	buf := make([]byte, 0, 1024*1024)
	for {
		m, err := deco.Decode(resp.Body, buf)
		if ctx.Err() != nil && ctx.Err() == context.Canceled {
			return nil
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		type payload struct {
			Bytes []byte `json:"bytes"`
		}

		var p payload
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			return err
		}

		data := p.Bytes

		fmt.Printf("received event: %s\n", string(data))

		// Gracefully skip over any data that isn't JSON-like. Anthropic's API sometimes sends
		// non-documented data over the stream, like timestamps.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event anthropicCompletionResponse
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w - body: %s", err, string(data))
		}

		// Collect the whole completion, AWS already uses the new Anthropic API
		// that sends partial completion results, but our clients still expect
		// a fill completion to be returned.
		totalCompletion += event.Completion

		err = sendEvent(types.CompletionResponse{
			Completion: totalCompletion,
			StopReason: event.StopReason,
		})
		if err != nil {
			return err
		}
	}
}

func (c *awsBedrockAnthropicCompletionStreamClient) makeRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	defaultConfig, err := config.LoadDefaultConfig(ctx, awsConfigOptsForKeyConfig(c.endpoint, c.accessToken)...)
	if err != nil {
		return nil, errors.Wrap(err, "loading aws config")
	}

	creds, err := defaultConfig.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving aws credentials")
	}

	if requestParams.TopK == -1 {
		requestParams.TopK = 0
	}

	if requestParams.TopP == -1 {
		requestParams.TopP = 0
	}

	prompt, err := anthropic.GetPrompt(requestParams.Messages)
	if err != nil {
		return nil, err
	}
	// Backcompat: Remove this code once enough clients are upgraded and we drop the
	// Prompt field on requestParams.
	if prompt == "" {
		prompt = requestParams.Prompt
	}

	if len(requestParams.StopSequences) == 0 {
		requestParams.StopSequences = []string{anthropic.HUMAN_PROMPT}
	}

	if requestParams.MaxTokensToSample == 0 {
		requestParams.MaxTokensToSample = 300
	}

	payload := anthropicCompletionsRequestParameters{
		StopSequences:     requestParams.StopSequences,
		Temperature:       requestParams.Temperature,
		MaxTokensToSample: requestParams.MaxTokensToSample,
		TopP:              requestParams.TopP,
		TopK:              requestParams.TopK,
		Prompt:            prompt,
		AnthropicVersion:  "bedrock-2023-05-31",
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	apiURL := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("bedrock.%s.amazonaws.com", c.endpoint),
		Path:   fmt.Sprintf("/model/%s/invoke-with-response-stream", requestParams.Model),
	}

	if !stream {
		apiURL.Path = fmt.Sprintf("/model/%s/invoke", requestParams.Model)
	}

	fmt.Printf("talking to API at %s: %s\n", apiURL.String(), string(reqBody))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(reqBody)
	if err := v4.NewSigner().SignHTTP(ctx, creds, req, hex.EncodeToString(hash[:]), "bedrock", c.endpoint, time.Now()); err != nil {
		return nil, errors.Wrap(err, "signing request")
	}

	req.Header.Set("Cache-Control", "no-cache")
	if !stream {
		req.Header.Set("Accept", "application/json")
	} else {
		req.Header.Set("Accept", "application/vnd.amazon.eventstream")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client", clientID)
	req.Header.Set("X-Amzn-Bedrock-Accept", "*/*")
	// Don't store the prompt in the prompt history.
	req.Header.Set("X-Amzn-Bedrock-Save", "false")

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("AWS Bedrock", resp)
	}

	return resp, nil
}

type anthropicCompletionsRequestParameters struct {
	Prompt            string   `json:"prompt"`
	Temperature       float32  `json:"temperature,omitempty"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	TopP              float32  `json:"top_p,omitempty"`
	// Model             string   `json:"model"`
	// Stream            bool     `json:"stream"`
	AnthropicVersion string `json:"anthropic_version"`
}

type anthropicCompletionResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}

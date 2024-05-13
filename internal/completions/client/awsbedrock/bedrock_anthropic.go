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
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go/logging"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type awsBedrockAnthropicCompletionStreamClient struct {
	cli          httpcli.Doer
	accessToken  string
	endpoint     string
	tokenManager tokenusage.Manager
}

type awsTitanAnthropicCompletionStreamClient struct {
	cli          httpcli.Doer
	accessToken  string
	endpoint     string
	tokenManager tokenusage.Manager
}

func (c *awsBedrockAnthropicCompletionStreamClient) Complete(
	ctx context.Context,
	feature types.CompletionsFeature,
	version types.CompletionsVersion,
	requestParams types.CompletionRequestParameters,
	logger log.Logger,
) (*types.CompletionResponse, error) {
	resp, err := c.makeRequest(ctx, requestParams, version, false)
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}
	defer resp.Body.Close()

	var response bedrockAnthropicNonStreamingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "decoding response")
	}
	completion := ""
	for _, content := range response.Content {
		completion += content.Text
	}

	err = c.tokenManager.UpdateTokenCountsFromModelUsage(response.Usage.InputTokens, response.Usage.OutputTokens, "anthropic/"+requestParams.Model, string(feature), tokenusage.AwsBedrock)
	if err != nil {
		return nil, err
	}
	return &types.CompletionResponse{
		Completion: completion,
		StopReason: response.StopReason,
	}, nil
}

func (a *awsBedrockAnthropicCompletionStreamClient) Stream(
	ctx context.Context,
	feature types.CompletionsFeature,
	version types.CompletionsVersion,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
	logger log.Logger,
) error {
	resp, err := a.makeRequest(ctx, requestParams, version, true)
	if err != nil {
		return errors.Wrap(err, "making request")
	}
	defer resp.Body.Close()
	var sentEvent bool

	// totalCompletion is the complete completion string, bedrock already uses
	// the new incremental Anthropic API, but our clients still expect a full
	// response in each event.
	var totalCompletion string
	var inputPromptTokens int
	dec := eventstream.NewDecoder(func(do *eventstream.DecoderOptions) {
		do.LogMessages = true
		do.Logger = logging.NewStandardLogger(os.Stderr)
	})
	// Allocate a 10 MB buffer for decoding.
	buf := make([]byte, 0, 10_000*1024)
	for {
		m, err := dec.Decode(resp.Body, buf)

		// Exit early on context cancellation.
		if ctx.Err() != nil && ctx.Err() == context.Canceled {
			return nil
		}

		// AWS's event stream decoder returns EOF once completed, so return.
		if err == io.EOF {
			if !sentEvent {
				return errors.New("stream closed with no events")
			}
			return nil
		}

		// For any other error, return.
		if err != nil {
			return err
		}

		// Unmarshal the event payload from the stream.
		var p awsEventStreamPayload
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			return errors.Wrap(err, "unmarshaling event payload")
		}

		data := p.Bytes

		// Gracefully skip over any data that isn't JSON-like. Anthropic's API sometimes sends
		// non-documented data over the stream, like timestamps.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event bedrockAnthropicStreamingResponse
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w - body: %s", err, string(data))
		}
		stopReason := ""
		switch event.Type {
		case "message_start":
			if event.Message != nil && event.Message.Usage != nil {
				inputPromptTokens = event.Message.Usage.InputTokens
			}
			continue
		case "content_block_delta":
			if event.Delta != nil {
				totalCompletion += event.Delta.Text
			}
		case "message_delta":
			if event.Delta != nil {
				stopReason = event.Delta.StopReason
				err = a.tokenManager.UpdateTokenCountsFromModelUsage(inputPromptTokens, event.Usage.OutputTokens, "anthropic/"+requestParams.Model, string(feature), tokenusage.AwsBedrock)
				if err != nil {
					logger.Warn("Failed to count tokens with the token manager %w ", log.Error(err))
				}
			}
		default:
			continue
		}
		sentEvent = true
		err = sendEvent(types.CompletionResponse{
			Completion: totalCompletion,
			StopReason: stopReason,
		})
		if err != nil {
			return errors.Wrap(err, "sending event")
		}
	}
}

func (c *awsBedrockAnthropicCompletionStreamClient) makeRequest(ctx context.Context, requestParams types.CompletionRequestParameters, version types.CompletionsVersion, stream bool) (*http.Response, error) {
	defaultConfig, err := config.LoadDefaultConfig(ctx, awsConfigOptsForKeyConfig(c.endpoint, c.accessToken)...)
	if err != nil {
		return nil, errors.Wrap(err, "loading aws config")
	}

	if requestParams.TopK == -1 {
		requestParams.TopK = 0
	}

	if requestParams.TopP == -1 {
		requestParams.TopP = 0
	}

	if requestParams.MaxTokensToSample == 0 {
		requestParams.MaxTokensToSample = 300
	}

	creds, err := defaultConfig.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving aws credentials")
	}

	convertedMessages := requestParams.Messages
	stopSequences := removeWhitespaceOnlySequences(requestParams.StopSequences)
	if version == types.CompletionsVersionLegacy {
		convertedMessages = types.ConvertFromLegacyMessages(convertedMessages)
	}

	messages, err := toAnthropicMessages(convertedMessages)
	if err != nil {
		return nil, err
	}

	// Convert the first message from `system` to a top-level system prompt
	system := "" // prevent the upstream API from setting this
	if len(messages) > 0 && messages[0].Role == types.SYSTEM_MESSAGE_SPEAKER {
		system = messages[0].Content[0].Text
		messages = messages[1:]
	}

	payload := bedrockAnthropicCompletionsRequestParameters{
		StopSequences:    stopSequences,
		Temperature:      requestParams.Temperature,
		MaxTokens:        requestParams.MaxTokensToSample,
		TopP:             requestParams.TopP,
		TopK:             requestParams.TopK,
		Messages:         messages,
		System:           system,
		AnthropicVersion: "bedrock-2023-05-31",
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling request body")
	}
	apiURL, err := url.Parse(c.endpoint)
	if err != nil || apiURL.Scheme == "" {
		apiURL = &url.URL{
			Scheme: "https",
			Host:   fmt.Sprintf("bedrock-runtime.%s.amazonaws.com", defaultConfig.Region),
		}
	}

	if stream {
		apiURL.Path = fmt.Sprintf("/model/%s/invoke-with-response-stream", requestParams.Model)
	} else {
		apiURL.Path = fmt.Sprintf("/model/%s/invoke", requestParams.Model)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	// Sign the request with AWS credentials.
	hash := sha256.Sum256(reqBody)
	if err := v4.NewSigner().SignHTTP(ctx, creds, req, hex.EncodeToString(hash[:]), "bedrock", defaultConfig.Region, time.Now()); err != nil {
		return nil, errors.Wrap(err, "signing request")
	}

	req.Header.Set("Cache-Control", "no-cache")
	if stream {
		req.Header.Set("Accept", "application/vnd.amazon.eventstream")
	} else {
		req.Header.Set("Accept", "application/json")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client", clientID)
	req.Header.Set("X-Amzn-Bedrock-Accept", "*/*")
	// Don't store the prompt in the prompt history.
	req.Header.Set("X-Amzn-Bedrock-Save", "false")

	// Make the request.
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "make request to bedrock")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("AWS Bedrock", resp)
	}

	return resp, nil
}

type bedrockAnthropicNonStreamingResponse struct {
	Content    []bedrockAnthropicMessageContent      `json:"content"`
	StopReason string                                `json:"stop_reason"`
	Usage      bedrockAnthropicMessagesResponseUsage `json:"usage"`
}

// AnthropicMessagesStreamingResponse captures all relevant-to-us fields from each relevant SSE event from https://docs.anthropic.com/claude/reference/messages_post.
type bedrockAnthropicStreamingResponse struct {
	Type         string                                       `json:"type"`
	Delta        *bedrockAnthropicStreamingResponseTextBucket `json:"delta"`
	ContentBlock *bedrockAnthropicStreamingResponseTextBucket `json:"content_block"`
	Usage        *bedrockAnthropicMessagesResponseUsage       `json:"usage"`
	Message      *bedrockAnthropicStreamingResponseMessage    `json:"message"`
}

type bedrockAnthropicStreamingResponseMessage struct {
	Usage *bedrockAnthropicMessagesResponseUsage `json:"usage"`
}

type bedrockAnthropicMessagesResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type bedrockAnthropicStreamingResponseTextBucket struct {
	Text       string `json:"text"`        // for event `content_block_delta`
	StopReason string `json:"stop_reason"` // for event `message_delta`
}

type bedrockAnthropicCompletionsRequestParameters struct {
	Messages      []bedrockAnthropicMessage `json:"messages,omitempty"`
	Temperature   float32                   `json:"temperature,omitempty"`
	TopP          float32                   `json:"top_p,omitempty"`
	TopK          int                       `json:"top_k,omitempty"`
	Stream        bool                      `json:"stream,omitempty"`
	StopSequences []string                  `json:"stop_sequences,omitempty"`
	MaxTokens     int                       `json:"max_tokens,omitempty"`

	// These are not accepted from the client an instead are only used to talk to the upstream LLM
	// APIs directly (these do NOT need to be set when talking to Cody Gateway)
	System           string `json:"system,omitempty"`
	AnthropicVersion string `json:"anthropic_version"`
}

type bedrockAnthropicMessage struct {
	Role    string                           `json:"role"` // "user", "assistant", or "system" (only allowed for the first message)
	Content []bedrockAnthropicMessageContent `json:"content"`
}

type bedrockAnthropicMessageContent struct {
	Type string `json:"type"` // "text" or "image" (not yet supported)
	Text string `json:"text"`
}

func toAnthropicMessages(messages []types.Message) ([]bedrockAnthropicMessage, error) {
	anthropicMessages := make([]bedrockAnthropicMessage, 0, len(messages))

	for i, message := range messages {
		speaker := message.Speaker
		text := message.Text

		anthropicRole := message.Speaker

		switch speaker {
		case types.SYSTEM_MESSAGE_SPEAKER:
			if i != 0 {
				return nil, errors.New("system role can only be used in the first message")
			}
		case types.ASSISTANT_MESSAGE_SPEAKER:
		case types.HUMAN_MESSAGE_SPEAKER:
			anthropicRole = "user"
		default:
			return nil, errors.Errorf("unexpected role: %s", text)
		}

		if text == "" {
			return nil, errors.New("message content cannot be empty")
		}

		anthropicMessages = append(anthropicMessages, bedrockAnthropicMessage{
			Role:    anthropicRole,
			Content: []bedrockAnthropicMessageContent{{Text: text, Type: "text"}},
		})
	}

	return anthropicMessages, nil
}

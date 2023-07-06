package completions

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

const openAIURL = "https://api.openai.com/v1/chat/completions"

func NewOpenAIHandler(
	logger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	accessToken string,
	orgID string,
	allowedModels []string,
) http.Handler {
	return makeUpstreamHandler(
		logger,
		eventLogger,
		rs,
		rateLimitNotifier,
		string(conftypes.CompletionsProviderNameOpenAI),
		openAIURL,
		allowedModels,
		upstreamHandlerMethods[openaiRequest]{
			transformBody: func(body *openaiRequest) {
				// We don't want to let users generate multiple responses, as this would
				// mess with rate limit counting.
				if body.N > 1 {
					body.N = 1
				}
				// Null the user field, we don't want to allow users to specify it:
				body.User = ""
				// TODO: We can forward the actor ID here later if we want?
			},
			getRequestMetadata: func(body openaiRequest) (promptCharacterCount int, model string, additionalMetadata map[string]any) {
				for _, m := range body.Messages {
					promptCharacterCount += len(m.Content)
				}
				return promptCharacterCount, body.Model, map[string]any{"stream": body.Stream}
			},
			transformRequest: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", "Bearer "+accessToken)
				if orgID != "" {
					r.Header.Set("OpenAI-Organization", orgID)
				}
			},
			parseResponse: func(body openaiRequest, r io.Reader) (completionCharacterCount int) {
				// Try to parse the request we saw, if it was non-streaming, we can simply parse
				// it as JSON.
				if !body.Stream {
					var res openaiResponse
					if err := json.NewDecoder(r).Decode(&res); err != nil {
						logger.Error("failed to parse OpenAI response as JSON", log.Error(err))
						return 0
					}
					if len(res.Choices) > 0 {
						// TODO: Later, we should look at the usage field.
						return len(res.Choices[0].Content)
					}
					return 0
				}

				// Otherwise, we have to parse the event stream.
				dec := openai.NewDecoder(r)
				var finalCompletion string
				// Consume all the messages, but we only care about the last completion data.
				for dec.Scan() {
					data := dec.Data()

					// Gracefully skip over any data that isn't JSON-like.
					if !bytes.HasPrefix(data, []byte("{")) {
						continue
					}

					var event openaiResponse
					if err := json.Unmarshal(data, &event); err != nil {
						logger.Error("failed to decode event payload", log.Error(err), log.String("body", string(data)))
						continue
					}
					if len(event.Choices) > 0 {
						finalCompletion += event.Choices[0].Delta.Content
					}
				}

				if err := dec.Err(); err != nil {
					logger.Error("failed to decode OpenAI streaming response", log.Error(err))
				}
				return len(finalCompletion)
			},
			validateRequest: func(_ openaiRequest) error {
				return nil
			},
		},
	)
}

type openaiRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type openaiRequest struct {
	Model            string                 `json:"model"`
	Messages         []openaiRequestMessage `json:"messages"`
	Temperature      float32                `json:"temperature,omitempty"`
	TopP             float32                `json:"top_p,omitempty"`
	N                int                    `json:"n,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	PresencePenalty  float32                `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32                `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float32     `json:"logit_bias,omitempty"`
	User             string                 `json:"user,omitempty"`
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
	Content      string            `json:"content"`
	FinishReason string            `json:"finish_reason"`
}

type openaiResponse struct {
	// Usage is only available for non-streaming requests.
	Usage   openaiUsage    `json:"usage"`
	Model   string         `json:"model"`
	Choices []openaiChoice `json:"choices"`
}

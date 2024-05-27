package completions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/shared/config"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/fireworks"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func NewGoogleHandler(baseLogger log.Logger, eventLogger events.Logger, rs limiter.RedisStore, rateLimitNotifier notify.RateLimitNotifier, httpClient httpcli.Doer, config config.GoogleConfig, promptRecorder PromptRecorder, upstreamConfig UpstreamHandlerConfig) http.Handler {
	return makeUpstreamHandler[googleRequest](
		baseLogger,
		eventLogger,
		rs,
		rateLimitNotifier,
		httpClient,
		string(conftypes.CompletionsProviderNameGoogle),
		config.AllowedModels,
		&GoogleHandlerMethods{config: config},
		promptRecorder,
		upstreamConfig,
	)
}

type googleContentMessage struct {
	Role  string `json:"role"`
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
}

type googleRequest struct {
	Model            string                 `json:"model"`
	Contents         []googleContentMessage `json:"contents"`
	GenerationConfig struct {
		MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
	} `json:"generationConfig,omitempty"`

	Stream bool `json:"stream,omitempty"`
	stream bool
}

func (r googleRequest) ShouldStream() bool {
	return r.stream
}

func (r googleRequest) GetModel() string {
	return r.Model
}

func (r googleRequest) BuildPrompt() string {
	var sb strings.Builder
	for _, m := range r.Contents {
		for _, t := range m.Parts {
			sb.WriteString(t.Text + "\n")
		}
	}
	return sb.String()
}

type googleUsage struct {
	PromptTokens int `json:"promptTokenCount"`
	// Use the same name we use elsewhere (completion instead of candidates)
	CompletionTokens int `json:"candidatesTokenCount"`
	TotalTokens      int `json:"totalTokenCount"`
}

type googleResponse struct {
	// Usage is only available for non-streaming requests.
	UsageMetadata googleUsage                              `json:"usageMetadata"`
	Model         string                                   `json:"model"`
	Candidates    []struct{ Content googleContentMessage } `json:"candidates"`
}

type GoogleHandlerMethods struct {
	config config.GoogleConfig
}

func (g *GoogleHandlerMethods) getAPIURL(_ codygateway.Feature, req googleRequest) string {
	rpc := "generateContent"
	sseSuffix := ""
	if req.ShouldStream() {
		rpc = "streamGenerateContent"
		sseSuffix = "&alt=sse"
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:%s?key=%s%s", req.Model, rpc, g.config.AccessToken, sseSuffix)
	fmt.Println(url)
	return url
}

func (*GoogleHandlerMethods) validateRequest(_ context.Context, _ log.Logger, feature codygateway.Feature, _ googleRequest) error {
	if feature == codygateway.FeatureCodeCompletions {
		return errors.Newf("feature %q is currently not supported for Google", feature)
	}
	return nil
}

func (g *GoogleHandlerMethods) shouldFlagRequest(_ context.Context, _ log.Logger, req googleRequest) (*flaggingResult, error) {
	result, err := isFlaggedRequest(
		nil, /* tokenizer, meaning token counts aren't considered when for flagging consideration. */
		flaggingRequest{
			FlattenedPrompt: req.BuildPrompt(),
			MaxTokens:       int(req.GenerationConfig.MaxOutputTokens),
		},
		makeFlaggingConfig(g.config.FlaggingConfig))
	return result, err
}

func (*GoogleHandlerMethods) transformBody(body *googleRequest, _ string) {
	body.stream = body.Stream
	body.Stream = false
}

func (*GoogleHandlerMethods) getRequestMetadata(body googleRequest) (model string, additionalMetadata map[string]any) {
	return body.Model, map[string]any{"stream": body.ShouldStream()}
}

func (o *GoogleHandlerMethods) transformRequest(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
}

func (*GoogleHandlerMethods) parseResponseAndUsage(logger log.Logger, reqBody googleRequest, r io.Reader) (promptUsage, completionUsage usageStats) {
	// First, extract prompt usage details from the request.
	promptUsage.characters = len(reqBody.BuildPrompt())

	// Try to parse the request we saw, if it was non-streaming, we can simply parse
	// it as JSON.
	if !reqBody.ShouldStream() {
		var res googleResponse

		if err := json.NewDecoder(r).Decode(&res); err != nil {
			logger.Error("failed to parse fireworks response as JSON", log.Error(err))
			return promptUsage, completionUsage
		}
		fmt.Println(res)
		promptUsage.tokens = res.UsageMetadata.PromptTokens
		completionUsage.tokens = res.UsageMetadata.CompletionTokens
		if len(res.Candidates) > 0 {
			// TODO: Later, we should look at the usage field.
			completionUsage.characters = len(res.Candidates[0].Content.Parts[0].Text)
		}
		return promptUsage, completionUsage
	}

	// Otherwise, we have to parse the event stream.
	//
	// TODO: Does fireworks streaming include usage data?
	// Unclear in the API currently: https://readme.fireworks.ai/reference/createcompletion
	// For now, just count character usage, and set token counts to
	// -1 as sentinel values.
	promptUsage.tokens = -1
	promptUsage.tokenizerTokens = -1
	completionUsage.tokens = -1
	completionUsage.tokenizerTokens = -1

	dec := fireworks.NewDecoder(r)
	// Consume all the messages, but we only care about the last completion data.
	for dec.Scan() {
		data := dec.Data()

		// Gracefully skip over any data that isn't JSON-like.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event fireworksResponse
		if err := json.Unmarshal(data, &event); err != nil {
			logger.Error("failed to decode event payload", log.Error(err), log.String("body", string(data)))
			continue
		}

		if len(event.Choices) > 0 {
			completionUsage.characters += len(event.Choices[0].Text)
		}
		// These are only included in the last message, so we're not worried about overwriting
		if event.Usage.PromptTokens > 0 {
			promptUsage.tokens = event.Usage.PromptTokens
		}
		if event.Usage.CompletionTokens > 0 {
			completionUsage.tokens = event.Usage.CompletionTokens
		}
	}
	if err := dec.Err(); err != nil {
		logger.Error("failed to decode Google streaming response", log.Error(err))
	}
	if completionUsage.tokens == -1 || promptUsage.tokens == -1 {
		logger.Warn("did not extract token counts from Google streaming response", log.Int("prompt-tokens", promptUsage.tokens), log.Int("completion-tokens", completionUsage.tokens))
	}

	return promptUsage, completionUsage
}

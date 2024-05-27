package completions

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/shared/config"
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
}

func (r googleRequest) ShouldStream() bool {
	return true
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
	PromptTokenCount int `json:"promptTokenCount"`
	// Use the same name we use elsewhere (completion instead of candidates)
	CompletionTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
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
	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:%s?key=%s%s", req.Model, rpc, g.config.AccessToken, sseSuffix)
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

func (*GoogleHandlerMethods) transformBody(_ *googleRequest, _ string) {
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
			logger.Error("failed to parse Google response as JSON", log.Error(err))
			return promptUsage, completionUsage
		}
		promptUsage.tokens = res.UsageMetadata.PromptTokenCount
		completionUsage.tokens = res.UsageMetadata.CompletionTokenCount
		if len(res.Candidates) > 0 {
			// TODO: Later, we should look at the usage field.
			completionUsage.characters = len(res.Candidates[0].Content.Parts[0].Text)
		}
		return promptUsage, completionUsage
	}

	// Otherwise, we have to parse the event stream.
	promptUsage.tokens = -1
	promptUsage.tokenizerTokens = -1
	completionUsage.tokens = -1
	completionUsage.tokenizerTokens = -1

	promptTokens, completionTokens, err := parseGoogleTokenUsage(r, logger)
	if err != nil {
		logger.Error("failed to decode Google streaming response", log.Error(err))
	}
	if completionUsage.tokens == -1 || promptUsage.tokens == -1 {
		logger.Warn("did not extract token counts from Google streaming response", log.Int("prompt-tokens", promptUsage.tokens), log.Int("completion-tokens", completionUsage.tokens))
	}
	promptUsage.tokens, completionUsage.tokens = promptTokens, completionTokens
	return promptUsage, completionUsage
}

const maxPayloadSize = 10 * 1024 * 1024 // 10mb

func parseGoogleTokenUsage(r io.Reader, logger log.Logger) (promptTokens int, completionTokens int, err error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 4096), maxPayloadSize)
	// bufio.ScanLines, except we look for \r\n\r\n which separate events.
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, []byte("\r\n\r\n")); i >= 0 {
			return i + 4, data[:i], nil
		}
		if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
			return i + 2, data[:i], nil
		}
		// If we're at EOF, we have a final, non-terminated event. This should
		// be empty.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
	scanner.Split(split)
	// skip to the last event
	var lastEvent []byte
	for scanner.Scan() {
		lastEvent = scanner.Bytes()
	}
	var res googleResponse
	if err := json.NewDecoder(bytes.NewReader(lastEvent[5:])).Decode(&res); err != nil {
		logger.Error("failed to parse Google response as JSON", log.Error(err))
		return -1, -1, err
	}
	return res.UsageMetadata.PromptTokenCount, res.UsageMetadata.CompletionTokenCount, nil
}

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

type GoogleHandlerMethods struct {
	config config.GoogleConfig
}

func (r googleRequest) ShouldStream() bool {
	return r.Stream
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

func (g *GoogleHandlerMethods) getAPIURL(feature codygateway.Feature, req googleRequest) string {
	rpc := "generateContent"
	sseSuffix := ""
	// If we're streaming, we need to use the stream endpoint.
	if req.ShouldStream() {
		rpc = "streamGenerateContent"
		sseSuffix = "&alt=sse"
	}
	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:%s?key=%s%s", req.Model, rpc, g.config.AccessToken, sseSuffix)
}

func (*GoogleHandlerMethods) validateRequest(_ context.Context, _ log.Logger, feature codygateway.Feature, _ googleRequest) error {
	if feature == codygateway.FeatureEmbeddings {
		return errors.Newf("feature %q is currently not supported for Google", feature)
	}
	return nil
}

func (g *GoogleHandlerMethods) shouldFlagRequest(ctx context.Context, logger log.Logger, req googleRequest) (*flaggingResult, error) {
	result, err := isFlaggedRequest(
		nil, // tokenizer, meaning token counts aren't considered when for flagging consideration.
		flaggingRequest{
			ModelName:       req.Model,
			FlattenedPrompt: req.BuildPrompt(),
			MaxTokens:       req.GenerationConfig.MaxOutputTokens,
		},
		makeFlaggingConfig(g.config.FlaggingConfig))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Used to modify the request body before it is sent to upstream.
func (*GoogleHandlerMethods) transformBody(gr *googleRequest, _ string) {
	// Remove Stream from the request body before sending it to Google.
	gr.Stream = false
}

func (*GoogleHandlerMethods) getRequestMetadata(body googleRequest) (model string, additionalMetadata map[string]any) {
	return body.Model, map[string]any{"stream": body.ShouldStream()}
}

func (o *GoogleHandlerMethods) transformRequest(downstreamRequest, upstreamRequest *http.Request) {
	upstreamRequest.Header.Set("Content-Type", "application/json")
}

func (*GoogleHandlerMethods) parseResponseAndUsage(logger log.Logger, reqBody googleRequest, r io.Reader, isStreamRequest bool) (promptUsage, completionUsage usageStats) {
	// First, extract prompt usage details from the request.
	promptUsage.characters = len(reqBody.BuildPrompt())
	// Try to parse the request we saw, if it was non-streaming, we can simply parse
	// it as JSON.
	if !isStreamRequest {
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
	promptUsage.tokens, completionUsage.tokens = -1, -1
	promptUsage.tokenizerTokens, completionUsage.tokenizerTokens = -1, -1
	promptTokens, completionTokens, err := parseGoogleTokenUsage(r, logger)
	if err != nil {
		logger.Error("failed to decode Google streaming response", log.Error(err))
	}
	promptUsage.tokens, completionUsage.tokens = promptTokens, completionTokens
	if completionUsage.tokens == -1 || promptUsage.tokens == -1 {
		logger.Warn("did not extract token counts from Google streaming response", log.Int("prompt-tokens", promptUsage.tokens), log.Int("completion-tokens", completionUsage.tokens))
	}
	return promptUsage, completionUsage
}

const maxPayloadSize = 10 * 1024 * 1024 // 10mb

func parseGoogleTokenUsage(r io.Reader, logger log.Logger) (promptTokens int, completionTokens int, err error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 4096), maxPayloadSize)
	scanner.Split(bufio.ScanLines)

	var lastNonEmptyLine []byte

	// Find the last non-empty line in the stream.
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) > 0 {
			lastNonEmptyLine = line
		}
	}

	if bytes.HasPrefix(bytes.TrimSpace(lastNonEmptyLine), []byte("data: ")) {
		event := lastNonEmptyLine[5:]
		var res googleResponse
		if err := json.NewDecoder(bytes.NewReader(event)).Decode(&res); err != nil {
			logger.Error("failed to parse Google response as JSON", log.Error(err))
			return -1, -1, err
		}
		return res.UsageMetadata.PromptTokenCount, res.UsageMetadata.CompletionTokenCount, nil
	}

	logger.Warn("no Google response found", log.Error(err))
	return -1, -1, nil
}

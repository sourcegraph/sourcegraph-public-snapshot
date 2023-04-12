package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-api/internal/streaming/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-api/internal/streaming/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-api/internal/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewCompletionsStreamHandler is an http handler which streams back completions results.
func NewCompletionsStreamHandler(logger log.Logger) http.Handler {
	return &completionHandler{callback: func(
		ctx context.Context,
		w http.ResponseWriter,
		completionStreamClient types.CompletionStreamClient,
		requestParams types.CompletionRequestParameters,
	) error {
		eventWriter, err := streamhttp.NewWriter(w)
		if err != nil {
			return err
		}
		// Always send a final done event so clients know the stream is shutting down.
		defer eventWriter.Event("done", map[string]any{})

		writeCompletion := func(event types.CompletionEvent) error {
			return eventWriter.Event("completion", event)
		}
		if err := completionStreamClient.Stream(ctx, requestParams, writeCompletion); err != nil {
			logger.Error("error while streaming completions", log.Error(err))
			eventWriter.Event("error", map[string]string{"error": err.Error()})
		}

		return nil
	}}
}

// NewCompletionsFinalHandler is an http handler which sends back the final completion result.
func NewCompletionsFinalHandler(logger log.Logger) http.Handler {
	return &completionHandler{callback: func(
		ctx context.Context,
		w http.ResponseWriter,
		completionStreamClient types.CompletionStreamClient,
		requestParams types.CompletionRequestParameters,
	) error {
		var lastEvent types.CompletionEvent
		writeCompletion := func(event types.CompletionEvent) error {
			lastEvent = event
			return nil
		}
		if err := completionStreamClient.Stream(ctx, requestParams, writeCompletion); err != nil {
			return err
		}

		_ = json.NewEncoder(w).Encode(lastEvent)
		return nil
	}}
}

type completionHandler struct {
	callback func(
		ctx context.Context,
		w http.ResponseWriter,
		completionStreamClient types.CompletionStreamClient,
		requestParams types.CompletionRequestParameters,
	) error
}

const maxRequestDuration = time.Minute

func (h *completionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), maxRequestDuration)
	defer cancel()

	// Gate via feature flags on Sourcegraph.com
	if envvar.SourcegraphDotComMode() && !cody.IsCodyExperimentalFeatureFlagEnabled(ctx) {
		http.Error(w, "cody experimental feature flag is not enabled for current user", http.StatusUnauthorized)
		return
	}

	if err := h.handle(ctx, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *completionHandler) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
	tr, ctx := trace.New(ctx, "completions.ServeStream", "Completions")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	completionsConfig := conf.Get().Completions
	if completionsConfig == nil || !completionsConfig.Enabled {
		return errors.New("completions are not enabled")
	}

	completionStreamClient, err := getCompletionStreamClient(completionsConfig.Provider, completionsConfig.AccessToken, completionsConfig.Model)
	if err != nil {
		return err
	}

	var requestParams types.CompletionRequestParameters
	if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
		return err
	}

	return h.callback(ctx, w, completionStreamClient, requestParams)
}

func getCompletionStreamClient(provider string, accessToken string, model string) (types.CompletionStreamClient, error) {
	switch provider {
	case "anthropic":
		return anthropic.NewAnthropicCompletionStreamClient(httpcli.ExternalDoer, accessToken, model), nil
	case "openai":
		return openai.NewOpenAIChatCompletionsStreamClient(httpcli.ExternalDoer, accessToken, model), nil
	default:
		return nil, errors.Newf("unknown completion stream provider: %s", provider)
	}
}

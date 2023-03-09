package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/streaming/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const maxRequestDuration = time.Minute

// NewCompletionsStreamHandler is an http handler which streams back completions results.
func NewCompletionsStreamHandler(logger log.Logger) http.Handler {
	return &streamHandler{logger: logger}
}

type streamHandler struct {
	logger log.Logger
}

func getCompletionStreamClient(provider string, accessToken string, model string) (types.CompletionStreamClient, error) {
	switch provider {
	case "anthropic":
		return anthropic.NewAnthropicCompletionStreamClient(httpcli.ExternalDoer, accessToken, model), nil
	default:
		return nil, errors.Newf("unknown completion stream provider: %s", provider)
	}
}

func (h *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	completionsConfig := conf.Get().Completions
	if completionsConfig == nil || !completionsConfig.Enabled {
		http.Error(w, "completions are not configured or disabled", http.StatusInternalServerError)
		return
	}

	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}

	var requestParams types.CompletionRequestParameters
	if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
		http.Error(w, "could not decode request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), maxRequestDuration)
	defer cancel()

	var err error
	tr, ctx := trace.New(ctx, "completions.ServeStream", "Completions")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	completionStreamClient, err := getCompletionStreamClient(completionsConfig.Provider, completionsConfig.AccessToken, completionsConfig.Model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	eventWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Always send a final done event so clients know the stream is shutting down.
	defer eventWriter.Event("done", map[string]any{})

	err = completionStreamClient.Stream(ctx, requestParams, func(event types.CompletionEvent) error { return eventWriter.Event("completion", event) })
	if err != nil {
		h.logger.Error("error while streaming completions", log.Error(err))
		eventWriter.Event("error", map[string]string{"error": err.Error()})
		return
	}
}

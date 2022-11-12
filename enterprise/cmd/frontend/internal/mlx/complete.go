package mlx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type completeHandler struct {
	logger log.Logger
	db     database.DB
}

type CompleteParams struct {
	Prompt         string
	NumCompletions int
	MaxTokens      string
}

type CompleteResult struct {
	Choices []CompletionChoice
}

type CompletionChoice struct {
	Text string
}

func (h *completeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx, cancel := context.WithTimeout(r.Context(), maxRequestDuration)
	defer cancel()

	var params CompleteParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tr, ctx := trace.New(ctx, "mlx.complete.ServeHTTP", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	eventWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Always send a final done event so clients know the stream is shutting down.
	defer eventWriter.Event("done", map[string]any{})

	// Log events to trace
	eventWriter.StatHook = eventStreamOTHook(tr.LogFields)

	_ = eventWriter.Event("progress", CompleteResult{
		Choices: []CompletionChoice{
			{Text: "Hello, world!"},
		},
	})
}

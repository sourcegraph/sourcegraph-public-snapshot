package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/response"
	llmproxy "github.com/sourcegraph/sourcegraph/enterprise/internal/llm-proxy"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type bodyTransformer[T any] func(*T)
type requestTransformer func(*http.Request)
type requestMetadataRetriever[T any] func(T) (promptCharacterCount int, model string, additionalMetadata map[string]any)
type responseParser[T any] func(T, io.Reader) (completionCharacterCount int)

func makeUpstreamHandler[ReqT any](logger log.Logger, eventLogger events.Logger, upstreamAPIURL string, bodyTrans bodyTransformer[ReqT], rmr requestMetadataRetriever[ReqT], reqTrans requestTransformer, respParser responseParser[ReqT]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())

		// Parse the request body.
		var body ReqT
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			response.JSONError(logger, w, http.StatusBadRequest, errors.Wrap(err, "failed to parse request body"))
			return
		}

		bodyTrans(&body)

		// Re-marshal the payload for upstream to unset metadata and remove any properties
		// not known to us.
		upstreamPayload, err := json.Marshal(body)
		if err != nil {
			response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to marshal request body"))
			return
		}

		req, err := http.NewRequest(http.MethodPost, upstreamAPIURL, bytes.NewReader(upstreamPayload))
		if err != nil {
			response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to create request"))
			return
		}

		// Run the request transformer.
		reqTrans(req)

		{
			metadata := map[string]any{}
			promptCharacterCount, model, am := rmr(body)
			for k, v := range am {
				metadata[k] = v
			}
			metadata["prompt_character_count"] = promptCharacterCount
			metadata["model"] = model
			err = eventLogger.LogEvent(
				events.Event{
					Name:       llmproxy.EventNameCompletionsStarted,
					Source:     act.Source.Name(),
					Identifier: act.ID,
					Metadata:   metadata,
				},
			)
			if err != nil {
				logger.Error("failed to log event", log.Error(err))
			}
		}

		upstreamStarted := time.Now()
		var (
			upstreamStatusCode       int = -1
			completionCharacterCount int = -1
		)
		defer func() {
			err := eventLogger.LogEvent(
				events.Event{
					Name:       llmproxy.EventNameCompletionsFinished,
					Source:     act.Source.Name(),
					Identifier: act.ID,
					Metadata: map[string]any{
						"upstream_request_duration_ms": time.Since(upstreamStarted).Milliseconds(),
						"upstream_status_code":         upstreamStatusCode,
						"completion_character_count":   completionCharacterCount,
					},
				},
			)
			if err != nil {
				logger.Error("failed to log event", log.Error(err))
			}
		}()

		resp, err := httpcli.ExternalDoer.Do(req)
		if err != nil {
			response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to make request to Anthropic"))
			return
		}
		defer func() { _ = resp.Body.Close() }()

		// Forward upstream http headers.
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		// Forward status code.
		upstreamStatusCode = resp.StatusCode
		w.WriteHeader(resp.StatusCode)

		// Set up a buffer to capture the response as it's streamed and sent to the client.
		var responseBuf bytes.Buffer
		respBody := io.TeeReader(resp.Body, &responseBuf)
		// Forward response to client.
		_, _ = io.Copy(w, respBody)

		if upstreamStatusCode >= 200 && upstreamStatusCode < 300 {
			// Pass reader to response transformer to capture token counts.
			completionCharacterCount = respParser(body, &responseBuf)
		}
	})
}

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
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/anthropic"
	llmproxy "github.com/sourcegraph/sourcegraph/enterprise/internal/llm-proxy"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/complete"

func newAnthropicHandler(logger log.Logger, eventLogger events.Logger, anthropicAccessToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())

		// Parse the request body. This supports all fields from the public Anthropic documentation.
		var body anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			response.JSONError(logger, w, http.StatusBadRequest, errors.Wrap(err, "failed to parse request body"))
			return
		}

		// Null the metadata field, we don't want to allow users to specify it:
		body.Metadata = nil
		// TODO: We can forward the actor ID here later if we want?

		// Re-marshal the payload for upstream to unset metadata and remove any properties
		// not known to us.
		upstreamPayload, err := json.Marshal(body)
		if err != nil {
			response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to marshal request body"))
			return
		}

		ar, err := http.NewRequest(http.MethodPost, anthropicAPIURL, bytes.NewReader(upstreamPayload))
		if err != nil {
			response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to create request"))
			return
		}

		err = eventLogger.LogEvent(
			events.Event{
				Name:       llmproxy.EventNameCompletionsStarted,
				Source:     act.Source.Name(),
				Identifier: act.ID,
				Metadata: map[string]any{
					"prompt_character_count": len(body.Prompt),
					"model":                  body.Model,
					"stream":                 body.Stream,
				},
			},
		)
		if err != nil {
			logger.Error("failed to log event", log.Error(err))
		}

		// Mimic headers set by the official Anthropic client:
		// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
		ar.Header.Set("Cache-Control", "no-cache")
		ar.Header.Set("Accept", "application/json")
		ar.Header.Set("Content-Type", "application/json")
		ar.Header.Set("Client", "sourcegraph-llm-proxy/1.0")
		ar.Header.Set("X-API-Key", anthropicAccessToken)

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

		resp, err := httpcli.ExternalDoer.Do(ar)
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

		// Now try to parse the request we saw, if it was non-streaming, we can simply parse
		// it as JSON.
		if !body.Stream {
			var res anthropicResponse
			if err := json.NewDecoder(&responseBuf).Decode(&res); err != nil {
				logger.Error("failed to parse anthropic response as JSON", log.Error(err))
				return
			}
			completionCharacterCount = len(res.Completion)
		} else {
			// Otherwise, we have to parse the event stream from anthropic.
			dec := anthropic.NewDecoder(&responseBuf)
			var lastCompletion string
			// Consume all the messages, but we only care about the last completion data.
			for dec.Scan() {
				data := dec.Data()

				// Gracefully skip over any data that isn't JSON-like. Anthropic's API sometimes sends
				// non-documented data over the stream, like timestamps.
				if !bytes.HasPrefix(data, []byte("{")) {
					continue
				}

				var event anthropicResponse
				if err := json.Unmarshal(data, &event); err != nil {
					logger.Error("failed to decode event payload", log.Error(err), log.String("body", string(data)))
					continue
				}
				lastCompletion = event.Completion
			}

			if err := dec.Err(); err != nil {
				logger.Error("failed to decode Anthropic streaming response", log.Error(err))
			} else {
				completionCharacterCount = len(lastCompletion)
			}
		}
	})
}

// anthropicRequest captures all known fields from https://console.anthropic.com/docs/api/reference.
type anthropicRequest struct {
	Prompt            string   `json:"prompt"`
	Model             string   `json:"model"`
	MaxTokensToSample int32    `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences"`
	Stream            bool     `json:"stream"`
	Temperature       float32  `json:"temperature"`
	TopK              int32    `json:"top_k"`
	TopP              int32    `json:"top_p"`
	Metadata          any      `json:"metadata"`
}

// anthropicResponse captures all relevant-to-us fields from https://console.anthropic.com/docs/api/reference.
type anthropicResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}

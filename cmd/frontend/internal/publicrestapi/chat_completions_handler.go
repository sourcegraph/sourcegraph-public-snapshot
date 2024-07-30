package publicrestapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	completions "github.com/sourcegraph/sourcegraph/internal/completions/types"
)

// chatCompletionsHandler implements the REST endpoint /chat/completions
type chatCompletionsHandler struct {
	logger sglog.Logger

	// apiHandler is the underlying implemenation of the Sourcegraph /.api/completions/stream endpoint.
	// We access this endpoint via HTTP to keep a single source-of-truth about LLM completions.
	// The goal with this OpenAI endpoint is compatibility, not optimal performance. Ideally, we
	// would have an in-house service we can use instead of going via HTTP but using HTTP
	// simplifies a lof of things (including testing).
	apiHandler http.Handler
}

var _ http.Handler = (*chatCompletionsHandler)(nil)

func (h *chatCompletionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var chatCompletionRequest CreateChatCompletionRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "io.ReadAll(r.Body) failed", http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(io.NopCloser(bytes.NewBuffer(body)))

	if err := decoder.Decode(&chatCompletionRequest); err != nil {
		http.Error(w, "decoder.Decode(body) failed", http.StatusInternalServerError)
		return
	}

	if chatCompletionRequest.Stream != nil && *chatCompletionRequest.Stream {
		http.Error(w, "stream is not supported", http.StatusBadRequest)
		return
	}

	for _, message := range chatCompletionRequest.Messages {
		if message.Role == "system" {
			http.Error(w, "system role is not supported", http.StatusBadRequest)
			return
		}
	}

	sgReq := transformToSGRequest(chatCompletionRequest)
	sgResp, err := h.forwardToAPIHandler(sgReq, r)
	if err != nil {
		http.Error(w, "failed to forward request to apiHandler "+err.Error(), http.StatusInternalServerError)
		return
	}

	chatCompletionResponse := transformToOpenAIResponse(sgResp, chatCompletionRequest)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(chatCompletionResponse); err != nil {
		h.logger.Error("writing /chat/completions response body", log.Error(err))
	}

}

func transformToSGRequest(openAIReq CreateChatCompletionRequest) completions.CompletionRequestParameters {
	maxTokens := 16 // Default in OpenAI openapi.yaml spec
	if openAIReq.MaxTokens != nil {
		maxTokens = *openAIReq.MaxTokens
	}

	var temperature float32
	if openAIReq.Temperature != nil {
		temperature = *openAIReq.Temperature
	}

	var topP float32
	if openAIReq.TopP != nil {
		topP = *openAIReq.TopP
	}
	stream := false // TODO: reject error when stream is true
	return completions.CompletionRequestParameters{
		MaxTokensToSample: maxTokens,
		Messages:          transformMessages(openAIReq.Messages),
		RequestedModel:    completions.TaintedModelRef(openAIReq.Model),
		Temperature:       temperature,
		TopP:              topP,
		Stream:            &stream,
		StopSequences:     openAIReq.Stop.Stop,
	}
}

func transformMessages(messages []ChatCompletionRequestMessage) []completions.Message {
	// Transform OpenAI messages to Sourcegraph format
	transformed := make([]completions.Message, len(messages))
	for i, msg := range messages {
		text := ""
		for _, part := range msg.Content.Parts {
			text += part.Text
		}
		speaker := msg.Role
		if speaker == "user" {
			speaker = "human"
		}
		transformed[i] = completions.Message{
			Speaker: speaker,
			Text:    text,
		}
	}
	return transformed
}

func (h *chatCompletionsHandler) forwardToAPIHandler(sgReq completions.CompletionRequestParameters, r *http.Request) (*completions.CompletionResponse, error) {
	// Create a new request to /.api/completions
	reqBody, err := json.Marshal(sgReq)
	if err != nil {
		return nil, errors.Newf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(r.Context(), "POST",
		"/.api/completions/stream?api-version=1&client-name=openai-rest-api&client-version=6.0.0",
		bytes.NewBuffer(reqBody))

	if err != nil {
		return nil, errors.Newf("failed to create request: %v", err)
	}

	// Set headers from the original request
	for headerName, values := range r.Header {
		for _, headerValue := range values {
			if headerName == "Authorization" && strings.HasPrefix(headerValue, "Bearer ") {
				// The OpenAI API expects the Authorization header to be in the
				// format "Bearer <token>" and we use the formatting "token <token>" internally.
				req.Header.Add(headerName, "token "+strings.TrimPrefix(headerValue, "Bearer "))
			} else {
				req.Header.Add(headerName, headerValue)
			}
		}
	}

	// Use a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Serve the request using the provided apiHandler
	h.apiHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		// TODO: properly return error matching OpenAI spec.
		return nil, errors.Newf("handler returned unexpected status code: got %v want %v, response body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	// Parse the response body
	var sgResp completions.CompletionResponse
	responseBytes := rr.Body.Bytes()
	err = json.Unmarshal(responseBytes, &sgResp)
	if err != nil {
		return nil, errors.Newf("failed to unmarshal response body %s: %v", string(responseBytes), err)
	}

	return &sgResp, nil
}

func transformToOpenAIResponse(sgResp *completions.CompletionResponse, openAIReq CreateChatCompletionRequest) CreateChatCompletionResponse {
	// Transform Sourcegraph response to OpenAI format
	return CreateChatCompletionResponse{
		ID:      "chat-" + generateUUID(), // You'll need to implement generateUUID
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   openAIReq.Model,
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChatCompletionResponseMessage{
					Role:    "assistant",
					Content: sgResp.Completion,
				},
				FinishReason: sgResp.StopReason,
			},
		},
		Usage: CompletionUsage{
			// You might need to implement token counting logic here
			// or get this information from the Sourcegraph response
		},
	}
}

var MockUUID = ""

func generateUUID() string {
	if MockUUID != "" {
		return MockUUID
	}
	uuid := uuid.New()
	return uuid.String()
}

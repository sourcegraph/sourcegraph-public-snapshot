package publicrestapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/modelconfig"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	completions "github.com/sourcegraph/sourcegraph/internal/completions/types"
	types "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

// chatCompletionsHandler implements the REST endpoint /chat/completions
type chatCompletionsHandler struct {
	logger sglog.Logger

	// apiHandler is the underlying implementation of the Sourcegraph /.api/completions/stream endpoint.
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
		http.Error(w, fmt.Sprintf("io.ReadAll: %v", err), http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(io.NopCloser(bytes.NewBuffer(body)))

	modelConfigSvc := modelconfig.Get()
	currentModelConfig, err := modelConfigSvc.Get()
	if err != nil {
		http.Error(w, fmt.Sprintf("modelConfigSvc.Get: %v", err), http.StatusInternalServerError)
		return
	}

	if err := decoder.Decode(&chatCompletionRequest); err != nil {
		http.Error(w, fmt.Sprintf("decoder.Decode: %v", err), http.StatusInternalServerError)
		return
	}

	if errorMsg := validateChatCompletionRequest(chatCompletionRequest); errorMsg != "" {
		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	if errorMsg := validateRequestedModel(chatCompletionRequest, currentModelConfig); errorMsg != "" {
		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	sgReq := transformToSGRequest(chatCompletionRequest)
	sgResp, err := h.forwardToAPIHandler(sgReq, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to forward request to apiHandler: %v", err), http.StatusInternalServerError)
		return
	}

	chatCompletionResponse := transformToOpenAIResponse(sgResp, chatCompletionRequest)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(chatCompletionResponse); err != nil {
		h.logger.Error("writing /chat/completions response body", log.Error(err))
	}

}

// validateRequestedModel checks that are only use the modelref syntax
// (${ProviderID}::${APIVersionID}::${ModelID}).  If the user passes the old
// syntax `${ProviderID}/${ModelID}`, then we try to return a helpful error
// message suggesting to use the new modelref syntax.
func validateRequestedModel(chatCompletionRequest CreateChatCompletionRequest, modelConfig *types.ModelConfiguration) string {
	closestModelRef := ""
	for _, model := range modelConfig.Models {
		if string(model.ModelRef) == chatCompletionRequest.Model {
			return ""
		}
		if model.DisplayName == chatCompletionRequest.Model || model.ModelName == chatCompletionRequest.Model {
			closestModelRef = string(model.ModelRef)
		} else if chatCompletionRequest.Model == fmt.Sprintf("%s/%s", model.ModelRef.ProviderID(), model.ModelRef.ModelID()) {
			closestModelRef = string(model.ModelRef)
		}
	}
	didYouMean := ""
	if closestModelRef != "" {
		didYouMean = fmt.Sprintf(" (similar to %s)", closestModelRef)
	}
	return fmt.Sprintf("model %s is not supported%s", chatCompletionRequest.Model, didYouMean)
}

func validateChatCompletionRequest(chatCompletionRequest CreateChatCompletionRequest) string {

	if chatCompletionRequest.N != nil && *chatCompletionRequest.N != 1 {
		return "n must be nil or 1"
	}

	if chatCompletionRequest.Stream != nil && *chatCompletionRequest.Stream {
		return "stream is not supported"
	}

	if chatCompletionRequest.Seed != nil {
		return "seed is not supported"
	}

	if chatCompletionRequest.ServiceTier != nil {
		return "service_tier is not supported"
	}

	if chatCompletionRequest.ResponseFormat != nil {
		return "response_format is not supported"
	}

	if chatCompletionRequest.StreamOptions != nil {
		return "stream_options is not supported"
	}

	if chatCompletionRequest.User != nil {
		return "user is not supported"
	}

	for _, message := range chatCompletionRequest.Messages {
		if message.Role == "system" {
			return "system role is not supported"
		}
	}

	return ""
}

func transformToSGRequest(openAIReq CreateChatCompletionRequest) completions.CodyCompletionRequestParameters {
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
	return completions.CodyCompletionRequestParameters{
		CompletionRequestParameters: completions.CompletionRequestParameters{
			MaxTokensToSample: maxTokens,
			Messages:          transformMessages(openAIReq.Messages),
			RequestedModel:    completions.TaintedModelRef(openAIReq.Model),
			Temperature:       temperature,
			TopP:              topP,
			Stream:            &stream,
			StopSequences:     openAIReq.Stop.Stop,
		},
		Fast: false,
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

func (h *chatCompletionsHandler) forwardToAPIHandler(sgReq completions.CodyCompletionRequestParameters, r *http.Request) (*completions.CompletionResponse, error) {
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
	return CreateChatCompletionResponse{
		ID:      "chat-" + generateUUID(),
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
		Usage: CompletionUsage{},
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

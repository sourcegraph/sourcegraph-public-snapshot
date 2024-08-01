package publicrestapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	sglog "github.com/sourcegraph/log"
)

// serveOpenAIChatCompletions is a handler for the OpenAI /v1/chat/completions endpoint.
func serveOpenAIChatCompletions(logger sglog.Logger, apiHandler http.Handler) func(w http.ResponseWriter, r *http.Request) (err error) {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		// Parse OpenAI request

		logger.Info("Received OpenAI chat completion request",
			sglog.String("method", r.Method),
			sglog.String("url", r.URL.String()),
			sglog.String("remote_addr", r.RemoteAddr),
			sglog.String("user_agent", r.UserAgent()),
		)

		start := time.Now()
		defer func() {
			logger.Info("Completed OpenAI chat completion request",
				sglog.Duration("duration", time.Since(start)),
				sglog.Error(err),
			)
		}()

		var openAIReq CreateChatCompletionRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		decoder := json.NewDecoder(io.NopCloser(bytes.NewBuffer(body)))

		if err := decoder.Decode(&openAIReq); err != nil {
			fmt.Println("decodeError", err)
			logger.Error("failed to decode OpenAI request", sglog.Error(err))
			return err
		}

		// Transform to /.api/completions/stream format
		sgReq := transformToSGRequest(openAIReq)

		// Forward request to apiHandler
		sgResp, err := forwardToAPIHandler(apiHandler, sgReq, r)
		if err != nil {
			logger.Error("failed to forward request to apiHandler", sglog.Error(err))
			return err
		}

		// Transform response to OpenAI format
		openAIResp := transformToOpenAIResponse(sgResp, openAIReq)

		// Send response
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(openAIResp)
	}
}

func transformToSGRequest(openAIReq CreateChatCompletionRequest) map[string]interface{} {
	// Transform OpenAI request to Sourcegraph format
	// You'll need to map fields appropriately
	return map[string]interface{}{
		"maxTokensToSample": openAIReq.MaxTokens,
		"messages":          transformMessages(openAIReq.Messages),
		"model":             openAIReq.Model,
		"temperature":       openAIReq.Temperature,
		"stream":            false, // TODO: implement streaming
		// Add other fields as needed
	}
}

func transformMessages(messages []ChatCompletionRequestMessage) []map[string]string {
	// Transform OpenAI messages to Sourcegraph format
	transformed := make([]map[string]string, len(messages))
	for i, msg := range messages {
		text := ""
		for _, part := range msg.Content.Parts {
			text += part.Text
		}
		speaker := msg.Role
		if speaker == "user" {
			speaker = "human"
		}
		transformed[i] = map[string]string{
			"speaker": speaker,
			"text":    text,
		}
	}
	return transformed
}

func forwardToAPIHandler(apiHandler http.Handler, sgReq map[string]interface{}, r *http.Request) (map[string]interface{}, error) {
	// Create a new request to /.api/completions
	reqBody, err := json.Marshal(sgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(r.Context(), "POST",
		"/.api/completions/stream?api-version=1&client-name=openai-rest-api&client-version=6.0.0",
		bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers from the original request
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Use a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Serve the request using the provided apiHandler
	apiHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		// TODO: properly return error matching OpenAI spec.
		return nil, fmt.Errorf("handler returned unexpected status code: got %v want %v, response body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	// Parse the response body
	var sgResp map[string]interface{}
	responseBytes := rr.Body.Bytes()
	err = json.Unmarshal(responseBytes, &sgResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body %s: %v", string(responseBytes), err)
	}

	return sgResp, nil
}

func transformToOpenAIResponse(sgResp map[string]interface{}, openAIReq CreateChatCompletionRequest) CreateChatCompletionResponse {
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
					Content: sgResp["completion"].(string),
				},
				FinishReason: sgResp["stopReason"].(string),
			},
		},
		Usage: CompletionUsage{
			// You might need to implement token counting logic here
			// or get this information from the Sourcegraph response
		},
	}
}

func generateUUID() string {
	// TODO
	return "12345678-1234-1234-1234-123456789012"
}

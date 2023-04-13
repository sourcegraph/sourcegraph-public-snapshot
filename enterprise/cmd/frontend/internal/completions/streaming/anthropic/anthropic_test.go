package anthropic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/types"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func linesToResponse(lines []string) []byte {
	responseBytes := []byte{}
	for _, line := range lines {
		responseBytes = append(responseBytes, []byte(fmt.Sprintf("data: %s", line))...)
		responseBytes = append(responseBytes, []byte("\r\n\r\n")...)
	}
	return responseBytes
}

func getMockClient(responseBody []byte) types.CompletionsClient {
	return NewAnthropicClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(responseBody))}, nil
		},
	}, "", "")
}

func TestValidAnthropicStream(t *testing.T) {
	var mockAnthropicResponseLines = []string{
		`{"completion": "Sure!"}`,
		`{"completion": "Sure! The Fibonacci sequence is defined as:\n\nF0 = 0\nF1 = 1\nFn = Fn-1 + Fn-2\n\nSo in Python, you can write it like this:\ndef fibonacci(n):\n    if n < 2:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)\n\nOr iteratively:\ndef fibonacci(n):\n    a, b = 0, 1\n    for i in range(n):\n        a, b = b, a + b\n    return a\n\nSo for example:\nprint(fibonacci(8))  # 21"}`,
		`2023.28.2 8:54`, // To test skipping over non-JSON data.
		`{"completion": "Sure! The Fibonacci sequence is defined as:\n\nF0 = 0\nF1 = 1\nFn = Fn-1 + Fn-2\n\nSo in Python, you can write it like this:\ndef fibonacci(n):\n    if n < 2:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)\n\nOr iteratively:\ndef fibonacci(n):\n    a, b = 0, 1\n    for i in range(n):\n        a, b = b, a + b\n    return a\n\nSo for example:\nprint(fibonacci(8))  # 21\n\nThe iterative"}`,
		"[DONE]",
	}

	mockClient := getMockClient(linesToResponse(mockAnthropicResponseLines))
	events := []types.ChatCompletionEvent{}
	err := mockClient.Stream(context.Background(), types.ChatCompletionRequestParameters{}, func(event types.ChatCompletionEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	autogold.ExpectFile(t, events)
}

func TestInvalidAnthropicStream(t *testing.T) {
	var mockAnthropicInvalidResponseLines = []string{`{]`}

	mockClient := getMockClient(linesToResponse(mockAnthropicInvalidResponseLines))
	err := mockClient.Stream(context.Background(), types.ChatCompletionRequestParameters{}, func(event types.ChatCompletionEvent) error { return nil })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	assert.Contains(t, err.Error(), "failed to decode event payload")
}

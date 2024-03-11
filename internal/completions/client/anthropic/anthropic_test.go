package anthropic

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func linesToResponse(lines []string, separator string) []byte {
	responseBytes := []byte{}
	for _, line := range lines {
		responseBytes = append(responseBytes, []byte(line)...)
		responseBytes = append(responseBytes, []byte(separator)...)
	}
	return responseBytes
}

func getMockClient(responseBody []byte, messagesApi bool) types.CompletionsClient {
	apiURL := "https://api.anthropic.com/v1/complete"
	if messagesApi {
		apiURL = "https://api.anthropic.com/v1/messages"
	}
	return NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(responseBody))}, nil
		},
	}, apiURL, "", false)
}

func TestValidAnthropicStream(t *testing.T) {
	var mockAnthropicResponseLines = []string{
		`data: {"completion": "Sure!"}`,
		`data: {"completion": "Sure! The Fibonacci sequence is defined as:\n\nF0 = 0\nF1 = 1\nFn = Fn-1 + Fn-2\n\nSo in Python, you can write it like this:\ndef fibonacci(n):\n    if n < 2:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)\n\nOr iteratively:\ndef fibonacci(n):\n    a, b = 0, 1\n    for i in range(n):\n        a, b = b, a + b\n    return a\n\nSo for example:\nprint(fibonacci(8))  # 21"}`,
		`data: 2023.28.2 8:54`, // To test skipping over non-JSON data.
		`data: {"completion": "Sure! The Fibonacci sequence is defined as:\n\nF0 = 0\nF1 = 1\nFn = Fn-1 + Fn-2\n\nSo in Python, you can write it like this:\ndef fibonacci(n):\n    if n < 2:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)\n\nOr iteratively:\ndef fibonacci(n):\n    a, b = 0, 1\n    for i in range(n):\n        a, b = b, a + b\n    return a\n\nSo for example:\nprint(fibonacci(8))  # 21\n\nThe iterative"}`,
		"data: [DONE]",
	}

	mockClient := getMockClient(linesToResponse(mockAnthropicResponseLines, "\r\n\r\n"), false)
	events := []types.CompletionResponse{}
	err := mockClient.Stream(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{}, func(event types.CompletionResponse) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	autogold.ExpectFile(t, events)
}

func TestValidAnthropicMessagesStream(t *testing.T) {
	var mockAnthropicMessagesResponseLines = []string{
		`event: message_start
		data: {"type": "message_start", "message": {"id": "msg_1nZdL29xx5MUA1yADyHTEsnR8uuvGzszyY", "type": "message", "role": "assistant", "content": [], "model": "claude-3-opus-20240229", "stop_reason": null, "stop_sequence": null, "usage": {"input_tokens": 25, "output_tokens": 1}}}`,
		`event: content_block_start
		data: {"type": "content_block_start", "index":0, "content_block": {"type": "text", "text": ""}}`,
		`event: ping
		data: {"type": "ping"}`,
		`event: content_block_delta
		data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "He"}}`,
		`event: content_block_delta
		data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "llo"}}`,
		`event: content_block_delta
		data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "!"}}`,
		`event: content_block_stop
		data: {"type": "content_block_stop", "index": 0}`,
		`event: message_delta
		data: {"type": "message_delta", "delta": {"stop_reason": "end_turn", "stop_sequence":null, "usage":{"output_tokens": 15}}}`,
		`event: message_stop
		data: {"type": "message_stop"}`,
	}

	mockClient := getMockClient(linesToResponse(mockAnthropicMessagesResponseLines, "\n\n"), true)
	events := []types.CompletionResponse{}
	stream := true
	err := mockClient.Stream(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{
		Messages: []types.Message{
			{Speaker: "human", Text: "Servus!"},
		},
		Stream: &stream,
	}, func(event types.CompletionResponse) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	autogold.ExpectFile(t, events)
}

func TestInvalidAnthropicStream(t *testing.T) {
	var mockAnthropicInvalidResponseLines = []string{`data:{]`}

	mockClient := getMockClient(linesToResponse(mockAnthropicInvalidResponseLines, "\r\n\r\n"), false)
	err := mockClient.Stream(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{}, func(event types.CompletionResponse) error { return nil })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	assert.Contains(t, err.Error(), "failed to decode event payload")
}

func TestErrStatusNotOK(t *testing.T) {
	mockClient := NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(bytes.NewReader([]byte("oh no, please slow down!"))),
			}, nil
		},
	}, "", "", false)

	t.Run("Complete", func(t *testing.T) {
		resp, err := mockClient.Complete(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{})
		require.Error(t, err)
		assert.Nil(t, resp)

		autogold.Expect("Anthropic: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
		_, ok := types.IsErrStatusNotOK(err)
		assert.True(t, ok)
	})

	t.Run("Stream", func(t *testing.T) {
		err := mockClient.Stream(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{}, func(event types.CompletionResponse) error { return nil })
		require.Error(t, err)

		autogold.Expect("Anthropic: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
		_, ok := types.IsErrStatusNotOK(err)
		assert.True(t, ok)
	})
}

func TestCompleteApiToMessages(t *testing.T) {
	var response *http.Request
	mockClient := NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			response = r
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(bytes.NewReader([]byte("oh no, please slow down!"))),
			}, nil
		},
	}, "https://api.anthropic.com/v1/messages", "", false)
	messages := []types.Message{
		{Speaker: "human", Text: "¡Hola!"},
		// /complete prompts can have human messages without an assistant response. These should
		// be ignored.
		{Speaker: "assistant", Text: ""},
		{Speaker: "human", Text: "Servus!"},
		// /complete prompts might end with an empty assistant message
		{Speaker: "assistant"},
	}

	t.Run("Complete", func(t *testing.T) {
		resp, err := mockClient.Complete(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{Messages: messages})
		require.Error(t, err)
		assert.Nil(t, resp)

		assert.NotNil(t, response)
		body, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		autogold.Expect(body).Equal(t, []byte(`{"messages":[{"role":"user","content":[{"type":"text","text":"Servus!"}]}],"model":""}`))
	})

	t.Run("Stream", func(t *testing.T) {
		stream := true
		err := mockClient.Stream(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{Messages: messages, Stream: &stream}, func(event types.CompletionResponse) error { return nil })
		require.Error(t, err)

		assert.NotNil(t, response)
		body, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		autogold.Expect(body).Equal(t, []byte(`{"messages":[{"role":"user","content":[{"type":"text","text":"Servus!"}]}],"model":"","stream":true}`))
	})
}

func TestMessagesApiToComplete(t *testing.T) {
	var response *http.Request
	mockClient := NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			response = r
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(bytes.NewReader([]byte("oh no, please slow down!"))),
			}, nil
		},
	}, "https://api.anthropic.com/v1/complete", "", false)
	messages := []types.Message{
		// /messages responses can have a system message
		{Speaker: "system", Text: "You are an Austrian emperor."},
		{Speaker: "human", Text: "Servus!"},
		// No ending `assistant` message
	}

	t.Run("Complete", func(t *testing.T) {
		resp, err := mockClient.Complete(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{Messages: messages})
		require.Error(t, err)
		assert.Nil(t, resp)

		assert.NotNil(t, response)
		body, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		autogold.Expect(body).Equal(t, []byte(`{"prompt":"\n\nHuman: You are an Austrian emperor.\n\nAssistant: Ok.\n\nHuman: Servus!\n\nAssistant:","temperature":0,"max_tokens_to_sample":0,"stop_sequences":["\n\nHuman:"],"top_k":0,"top_p":0,"model":"","stream":false}`))
	})

	t.Run("Stream", func(t *testing.T) {
		stream := true
		err := mockClient.Stream(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{Messages: messages, Stream: &stream}, func(event types.CompletionResponse) error { return nil })
		require.Error(t, err)

		assert.NotNil(t, response)
		body, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		autogold.Expect(body).Equal(t, []byte(`{"prompt":"\n\nHuman: You are an Austrian emperor.\n\nAssistant: Ok.\n\nHuman: Servus!\n\nAssistant:","temperature":0,"max_tokens_to_sample":0,"stop_sequences":["\n\nHuman:"],"top_k":0,"top_p":0,"model":"","stream":true}`))
	})
}

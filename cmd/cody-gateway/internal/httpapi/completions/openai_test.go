package completions

import (
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
)

func TestOpenAIRequestGetTokenCount(t *testing.T) {
	logger := logtest.Scoped(t)

	t.Run("streaming", func(t *testing.T) {
		req := openaiRequest{Stream: true}
		r := strings.NewReader(openaiStreamingResponse)
		handler := &OpenAIHandlerMethods{}
		promptUsage, completionUsage := handler.parseResponseAndUsage(logger, req, r, true)

		assert.Equal(t, 427, promptUsage.tokens)
		assert.Equal(t, 12, completionUsage.tokens)
	})

	t.Run("non-streaming", func(t *testing.T) {
		req := openaiRequest{Stream: false}
		r := strings.NewReader(openaiNonStreamingResponse)
		handler := &OpenAIHandlerMethods{}
		promptUsage, completionUsage := handler.parseResponseAndUsage(logger, req, r, false)

		assert.Equal(t, 12, promptUsage.tokens)
		assert.Equal(t, 9, completionUsage.tokens)
	})
}

var openaiStreamingResponse = `
data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":"Hello"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":"!"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" How"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" can"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" I"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" assist"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" you"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" with"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" your"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" coding"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" today"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":"?"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{},"logprobs":null,"finish_reason":"stop"}]}

data: {"id":"chatcmpl-8elzR9LfyjxFKon8GoRMdY4BVj6Zx","object":"chat.completion.chunk","created":1704728285,"model":"gpt-4-0613","system_fingerprint":null,"choices":[],"usage":{"prompt_tokens":427,"completion_tokens":12,"total_tokens":439}}
`

var openaiNonStreamingResponse = `{"id": "chatcmpl-8emDGlSur24VWoRtKriWO8XpuuFQi","object": "chat.completion","created": 1704729142,"model": "gpt-4-0613","choices": [{"index": 0,"message": {"role": "assistant","content": "Hello! How can I assist you today?"},"logprobs": null,"finish_reason": "stop"}],"usage": {"prompt_tokens": 12,"completion_tokens": 9,"total_tokens": 21},"system_fingerprint": null}`

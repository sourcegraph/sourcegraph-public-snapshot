package awsbedrock

type bedrockAnthropicNonStreamingResponse struct {
	Content    []bedrockAnthropicMessageContent      `json:"content"`
	StopReason string                                `json:"stop_reason"`
	Usage      bedrockAnthropicMessagesResponseUsage `json:"usage"`
}

// AnthropicMessagesStreamingResponse captures all relevant-to-us fields from each relevant SSE event.
// See: https://docs.anthropic.com/claude/reference/messages_post
type bedrockAnthropicStreamingResponse struct {
	Type         string                                       `json:"type"`
	Delta        *bedrockAnthropicStreamingResponseTextBucket `json:"delta"`
	ContentBlock *bedrockAnthropicStreamingResponseTextBucket `json:"content_block"`
	Usage        *bedrockAnthropicMessagesResponseUsage       `json:"usage"`
	Message      *bedrockAnthropicStreamingResponseMessage    `json:"message"`
}

type bedrockAnthropicStreamingResponseMessage struct {
	Usage *bedrockAnthropicMessagesResponseUsage `json:"usage"`
}

type bedrockAnthropicMessagesResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type bedrockAnthropicStreamingResponseTextBucket struct {
	Text       string `json:"text"`        // for event `content_block_delta`
	StopReason string `json:"stop_reason"` // for event `message_delta`
}

type bedrockAnthropicCompletionsRequestParameters struct {
	Messages      []bedrockAnthropicMessage `json:"messages,omitempty"`
	Temperature   float32                   `json:"temperature,omitempty"`
	TopP          float32                   `json:"top_p,omitempty"`
	TopK          int                       `json:"top_k,omitempty"`
	Stream        bool                      `json:"stream,omitempty"`
	StopSequences []string                  `json:"stop_sequences,omitempty"`
	MaxTokens     int                       `json:"max_tokens,omitempty"`

	// These are not accepted from the client an instead are only used to talk to the upstream LLM
	// APIs directly (these do NOT need to be set when talking to Cody Gateway)
	System           string `json:"system,omitempty"`
	AnthropicVersion string `json:"anthropic_version"`
}

type bedrockAnthropicMessage struct {
	Role    string                           `json:"role"` // "user", "assistant", or "system" (only allowed for the first message)
	Content []bedrockAnthropicMessageContent `json:"content"`
}

type bedrockAnthropicMessageContent struct {
	Type string `json:"type"` // "text" or "image" (not yet supported)
	Text string `json:"text"`
}

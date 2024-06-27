package google

type anthropicUsage struct {
	PromptTokenCount int `json:"promptTokenCount"`
	// Use the same name we use elsewhere (completion instead of candidates)
	CompletionTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}
type anthropicRequest struct {
	AnthropicVersion string             `json:"anthropic_version"`
	Messages         []anthropicMessage `json:"messages"`
	MaxTokens        int                `json:"max_tokens"`
	Stream           bool               `json:"stream"`
	System           string             `json:"system"`
}
type anthropicMessage struct {
	Role    string                 `json:"role"`
	Content []anthropicMessagePart `json:"content"`
}

type anthropicMessagePart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type anthropicContentMessage struct {
	Role  string                        `json:"role"`
	Parts []anthropicContentMessagePart `json:"parts"`
}

type anthropicContentMessagePart struct {
	Text string `json:"text"`
}

type anthropicResponse struct {
	Candidates     []anthropicCandidate      `json:"candidates"`
	UsageMetadata  anthropicUsage            `json:"usageMetadata"`
	SafetySettings []anthropicSafetySettings `json:"safetySettings,omitempty"`
	SafetyRatings  []anthropicSafetyRating   `json:"safetyRatings,omitempty"`
}

type anthropicCandidate struct {
	Content    anthropicContentMessage `json:"content,omitempty"`
	StopReason string                  `json:"finishReason,omitempty"`
}

type anthropicSafetyRating struct {
	Category         string  `json:"category"`
	Probability      string  `json:"probability"`
	ProbabilityScore float64 `json:"probabilityScore"`
	Severity         string  `json:"severity"`
	SeverityScore    float64 `json:"severityScore"`
}

// Safety setting, affecting the safety-blocking behavior.
// Ref: https://ai.google.dev/gemini-api/docs/safety-settings
type anthropicSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type anthropicStreamingResponseMessage struct {
	Usage *anthropicMessagesResponseUsage `json:"usage"`
}

type anthropicMessagesResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicStreamingResponseTextBucket struct {
	Text       string `json:"text"`        // for event `content_block_delta`
	StopReason string `json:"stop_reason"` // for event `message_delta`
}

type anthropicStreamingResponseDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicMessagesStreamingResponse captures all relevant-to-us fields from each relevant SSE event from https://docs.anthropic.com/claude/reference/messages_post.
type anthropicStreamingResponse struct {
	Type         string                                `json:"type"`
	Delta        *anthropicStreamingResponseDelta      `json:"delta"`
	ContentBlock *anthropicStreamingResponseTextBucket `json:"content_block"`
	Usage        *anthropicMessagesResponseUsage       `json:"usage"`
	Message      *anthropicStreamingResponseMessage    `json:"message"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicNonStreamingResponse struct {
	ID           string                         `json:"id"`
	Type         string                         `json:"type"`
	Role         string                         `json:"role"`
	Model        string                         `json:"model"`
	Content      []anthropicContent             `json:"content"`
	StopReason   string                         `json:"stop_reason"`
	StopSequence *string                        `json:"stop_sequence"`
	Usage        anthropicMessagesResponseUsage `json:"usage"`
}

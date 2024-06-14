package google

import "github.com/sourcegraph/sourcegraph/internal/httpcli"

type googleCompletionStreamClient struct {
	cli         httpcli.Doer
	accessToken string
	endpoint    string
}

// The request body for the completion stream endpoint.
// Ref: https://ai.google.dev/api/rest/v1beta/models/generateContent
// Ref: https://ai.google.dev/api/rest/v1beta/models/streamGenerateContent
type googleRequest struct {
	Contents []googleContentMessage `json:"contents"`
}

type googleContentMessage struct {
	Role  string                     `json:"role"`
	Parts []googleContentMessagePart `json:"parts"`
}

type googleContentMessagePart struct {
	Text string `json:"text"`
}

// Configuration options for model generation and outputs.
// Ref: https://ai.google.dev/api/rest/v1/GenerationConfig
type googleGenerationConfig struct {
	Temperature     float32  `json:"temperature,omitempty"`     // request.Temperature
	TopP            float32  `json:"topP,omitempty"`            // request.TopP
	TopK            int      `json:"topK,omitempty"`            // request.TopK
	StopSequences   []string `json:"stopSequences,omitempty"`   // request.StopSequences
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"` // request.MaxTokensToSample
	CandidateCount  int      `json:"candidateCount,omitempty"`  // request.CandidateCount
}

type googleResponse struct {
	Candidates []struct {
		Content    googleContentMessage `json:"content,omitempty"`
		StopReason string               `json:"finishReason,omitempty"`
	} `json:"candidates"`

	UsageMetadata  googleUsage            `json:"usageMetadata"`
	SafetySettings []googleSafetySettings `json:"safetySettings,omitempty"`
	SafetyRatings  []googleSafetyRating   `json:"safetyRatings,omitempty"`
}

type googleSafetyRating struct {
	Category         string  `json:"category"`
	Probability      string  `json:"probability"`
	ProbabilityScore float64 `json:"probabilityScore"`
	Severity         string  `json:"severity"`
	SeverityScore    float64 `json:"severityScore"`
}

// Safety setting, affecting the safety-blocking behavior.
// Ref: https://ai.google.dev/gemini-api/docs/safety-settings
type googleSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type googleUsage struct {
	PromptTokenCount int `json:"promptTokenCount"`
	// Use the same name we use elsewhere (completion instead of candidates)
	CompletionTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

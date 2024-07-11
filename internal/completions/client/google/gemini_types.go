package google

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type APIFamily string

type googleCompletionStreamClient struct {
	httpCli     httpcli.Doer
	gcpCli      *http.Client
	accessToken string
	endpoint    string
	viaGateway  bool
	apiFamily   APIFamily
}

// The request body for the completion stream endpoint.
// Ref: https://ai.google.dev/api/rest/v1beta/models/generateContent
// Ref: https://ai.google.dev/api/rest/v1beta/models/streamGenerateContent
type googleRequest struct {
	Model             string                 `json:"model"`
	Contents          []googleContentMessage `json:"contents"`
	GenerationConfig  googleGenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings    []googleSafetySettings `json:"safetySettings,omitempty"`
	SymtemInstruction string                 `json:"systemInstruction,omitempty"`

	// Stream is used for our internal routing of the Google Request, and is not part
	// of the Google API shape.
	Stream bool `json:"stream,omitempty"`
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
	Candidates    []googleCandidates `json:"candidates,omitempty"`
	UsageMetadata googleUsage        `json:"usageMetadata,omitempty"`
}

type googleCandidates struct {
	Content       googleContentMessage  `json:"content,omitempty"`
	FinishReason  string                `json:"finishReason,omitempty"`
	SafetyRatings []googleSafetyRatings `json:"safetyRatings,omitempty"`
}

type googleUsage struct {
	PromptTokenCount int `json:"promptTokenCount"`
	// Use the same name we use elsewhere (completion instead of candidates)
	CompletionTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// Safety setting, affecting the safety-blocking behavior.
// Ref: https://ai.google.dev/gemini-api/docs/safety-settings
type googleSafetySettings struct {
	Category  string `json:"category,omitempty"`
	Threshold string `json:"threshold,omitempty"`
}
type googleSafetyRatings struct {
	Category    string `json:"category,omitempty"`
	Probability string `json:"probability,omitempty"`
}

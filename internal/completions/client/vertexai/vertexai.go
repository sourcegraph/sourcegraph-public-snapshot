package vertexai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string) types.CompletionsClient {
	return &vertexAIChatCompletionStreamClient{
		cli:         cli,
		accessToken: accessToken,
		endpoint:    endpoint,
	}
}

type vertexAIChatCompletionStreamClient struct {
	cli         httpcli.Doer
	accessToken string
	endpoint    string
}

func (c *vertexAIChatCompletionStreamClient) Complete(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *vertexAIChatCompletionStreamClient) Stream(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	var resp *http.Response
	var err error

	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()
	if feature == types.CompletionsFeatureCode {
		resp, err = c.makeCompletionRequest(ctx, requestParams, true)
	} else {
		resp, err = c.makeRequest(ctx, requestParams, true)
	}
	if err != nil {
		return err
	}

	var response vertexResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}
	if len(response.Predictions) == 0 {
		// Empty response.
		return nil
	}
	var content string
	switch feature {
	case types.CompletionsFeatureCode:
		content += response.Predictions[0].Content
	case types.CompletionsFeatureChat:
		if len(response.Predictions[0].Candidates) > 0 {
			content += response.Predictions[0].Candidates[0].Content
		}
	}

	ev := types.CompletionResponse{
		Completion: content,
		StopReason: "done",
	}
	err = sendEvent(ev)
	if err != nil {
		return err
	}

	return nil
}

// makeRequest formats the request and calls the chat/completions endpoint for code_completion requests
func (c *vertexAIChatCompletionStreamClient) makeRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	payload := chatRequest{
		Parameters: parameters{
			CandidateCount:  1,
			MaxOutputTokens: requestParams.MaxTokensToSample,
			Temperature:     requestParams.Temperature,
			StopSequences:   requestParams.StopSequences,
		},
		Instances: []chatInstance{{Messages: []message{}}},
	}
	for _, m := range requestParams.Messages {
		if len(m.Text) == 0 {
			continue
		}
		var role string
		switch m.Speaker {
		case types.HUMAN_MESSAGE_SPEAKER:
			role = "user"
		case types.ASISSTANT_MESSAGE_SPEAKER:
			role = "bot"
			//
		default:
			role = strings.ToLower(role)
		}
		payload.Instances[0].Messages = append(payload.Instances[0].Messages, message{
			Author:  role,
			Content: m.Text,
		})
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(reqBody))

	url, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configured endpoint")
	}

	url.Path += fmt.Sprintf("/%s:predict", requestParams.Model)

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("VertexAI", resp)
	}

	return resp, nil
}

// makeCompletionRequest formats the request and calls the completions endpoint for code_completion requests
func (c *vertexAIChatCompletionStreamClient) makeCompletionRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*http.Response, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return nil, err
	}

	payload := vertexCompletionsRequest{
		Parameters: parameters{
			CandidateCount:  1,
			MaxOutputTokens: requestParams.MaxTokensToSample,
			StopSequences:   requestParams.StopSequences,
			Temperature:     requestParams.Temperature,
		},
		Instances: []completionsInstances{{Prefix: prompt}},
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(reqBody))
	url, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configured endpoint")
	}

	url.Path += fmt.Sprintf("/%s:predict", requestParams.Model)

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("VertexAI", resp)
	}

	return resp, nil
}

type chatRequest struct {
	Instances  []chatInstance `json:"instances"`
	Parameters parameters     `json:"parameters"`
}

type chatInstance struct {
	Messages []message `json:"messages"`
}

type message struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type vertexCompletionsRequest struct {
	Instances  []completionsInstances `json:"instances"`
	Parameters parameters             `json:"parameters"`
}

type completionsInstances struct {
	Prefix string `json:"prefix"`
}

type parameters struct {
	CandidateCount  int      `json:"candidateCount"`
	MaxOutputTokens int      `json:"maxOutputTokens"`
	StopSequences   []string `json:"stopSequences,omitempty"`
	Temperature     float32  `json:"temperature"`
}

type safetyAttribute struct {
	Categories []string `json:"categories"`
	Blocked    bool     `json:"blocked"`
	Scores     []int    `json:"scores"`
}

type candidate struct {
	Content string `json:"content"`
	Author  string `json:"author"`
}

type groundingMetadata struct {
	// fields
}

type citationMetadata struct {
	Citations []string `json:"citations"`
}

type prediction struct {
	Content string `json:"content"` // completion text
	//SafetyAttributes []safetyAttribute `json:"safetyAttributes"`
	Candidates []candidate `json:"candidates"` // chat completions
	//GroundingMetadata []groundingMetadata `json:"groundingMetadata"`
	//CitationMetadata  []citationMetadata  `json:"citationMetadata"`
	Score float32 `json:"score"` // prediction confidence score (code completions only)
}

type tokenMetadata struct {
	InputTokenCount struct {
		TotalTokens             int `json:"totalTokens"`
		TotalBillableCharacters int `json:"totalBillableCharacters"`
	} `json:"inputTokenCount"`
	OutputTokenCount struct {
		TotalTokens             int `json:"totalTokens"`
		TotalBillableCharacters int `json:"totalBillableCharacters"`
	} `json:"outputTokenCount"`
}

type metadata struct {
	TokenMetadata tokenMetadata `json:"tokenMetadata"`
}

type vertexResponse struct {
	Predictions []prediction `json:"predictions"`
	Metadata    metadata     `json:"metadata"`
}

func getPrompt(messages []types.Message) (string, error) {
	if l := len(messages); l != 1 {
		return "", errors.Errorf("expected to receive exactly one message with the prompt (got %d)", l)
	}

	return messages[0].Text, nil
}

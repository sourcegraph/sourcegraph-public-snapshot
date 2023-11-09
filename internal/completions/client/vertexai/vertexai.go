package vertexai

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	ai "cloud.google.com/go/aiplatform/apiv1"
	aipb "cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

// TODO: remove httpcli.Doer
func NewClient(cli httpcli.Doer, endpoint, accessToken string) types.CompletionsClient {
	// TODO: pass these options down to this function
	ctx := context.TODO()
	credentialsFile := "chris-warwick-dev-24660c44aa98.json" // TODO: do not hard-code
	credentialsJSON := ""

	// Client options
	var options []option.ClientOption
	if credentialsFile != "" {
		options = append(options, option.WithCredentialsFile(credentialsFile))
	}
	if credentialsJSON != "" {
		options = append(options, option.WithCredentialsJSON([]byte(credentialsJSON)))
	}

	cli2, err := ai.NewPredictionClient(ctx, options...)
	if err != nil {
		// TODO: bubble this error up
		panic(err)
	}

	_, err = url.Parse(endpoint)
	if err != nil {
		// TODO: bubble this error up
		panic(errors.Wrap(err, "failed to parse configured endpoint"))
	}

	return &vertexAIChatCompletionStreamClient{
		cli2:        cli2,
		cli:         cli, // TODO: remove
		accessToken: accessToken,
		endpoint:    endpoint,
	}
}

type vertexAIChatCompletionStreamClient struct {
	cli2        *ai.PredictionClient // TODO: should call .Close() when done with this
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
	if feature == types.CompletionsFeatureCode {
		resp, err := c.makeCompletionRequest(ctx, requestParams, true)
		if err != nil {
			return err
		}
		if len(resp.Predictions) == 0 {
			// Empty response
			return nil
		}
		// TODO: rewrite this hunk to use unmarshalValue
		content := resp.Predictions[0].GetStructValue().GetFields()["Content"].GetStringValue()

		ev := types.CompletionResponse{
			Completion: content,
			StopReason: "done",
		}
		err = sendEvent(ev)
		if err != nil {
			return err
		}
		return nil
	} else {
		resp, err := c.makeRequest(ctx, requestParams, true)
		if err != nil {
			return err
		}
		if len(resp.Predictions) == 0 {
			// Empty response
			return nil
		}
		var content string
		// TODO: rewrite this hunk to use unmarshalValue
		candidates := resp.Predictions[0].GetStructValue().GetFields()["Candidates"].GetListValue().GetValues()
		if len(candidates) > 0 {
			content += candidates[0].GetStructValue().GetFields()["Content"].GetStringValue()
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
}

// makeRequest formats the request and calls the chat/completions endpoint for code_completion requests
func (c *vertexAIChatCompletionStreamClient) makeRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*aipb.PredictResponse, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	params := parameters{
		CandidateCount:  1,
		MaxOutputTokens: requestParams.MaxTokensToSample,
		Temperature:     requestParams.Temperature,
		StopSequences:   requestParams.StopSequences,
	}
	instances := []chatInstance{{Messages: []message{}}}
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
		instances[0].Messages = append(instances[0].Messages, message{
			Author:  role,
			Content: m.Text,
		})
	}

	paramsPayload, err := marshalValue(params)
	if err != nil {
		return nil, err
	}
	instancesPayload, err := marshalValue(instances)
	if err != nil {
		return nil, err
	}
	return c.cli2.Predict(ctx, &aipb.PredictRequest{
		Parameters: paramsPayload,
		Instances:  instancesPayload.GetListValue().GetValues(),
	})
}

// makeCompletionRequest formats the request and calls the completions endpoint for code_completion requests
func (c *vertexAIChatCompletionStreamClient) makeCompletionRequest(ctx context.Context, requestParams types.CompletionRequestParameters, stream bool) (*aipb.PredictResponse, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	params := parameters{
		CandidateCount:  1,
		MaxOutputTokens: requestParams.MaxTokensToSample,
		StopSequences:   requestParams.StopSequences,
		Temperature:     requestParams.Temperature,
	}
	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return nil, err
	}
	instances := []completionsInstances{{Prefix: prompt}}

	paramsPayload, err := marshalValue(params)
	if err != nil {
		return nil, err
	}
	instancesPayload, err := marshalValue(instances)
	if err != nil {
		return nil, err
	}
	return c.cli2.Predict(ctx, &aipb.PredictRequest{
		Parameters: paramsPayload,
		Instances:  instancesPayload.GetListValue().GetValues(),
	})
}

// json.Marshal, but for structpb values
func marshalValue(x any) (*structpb.Value, error) {
	j, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Wrap(err, "Marshal")
	}
	var v structpb.Value
	if err := json.Unmarshal(j, &v); err != nil {
		return nil, errors.Wrap(err, "Unmarshal")
	}
	return &v, nil
}

// json.Unmarshal, but for structpb values
func unmarshalValue(data *structpb.Value, out interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}
	if err := json.Unmarshal(j, out); err != nil {
		return errors.Wrap(err, "Unmarshal")
	}
	return nil
}

type chatInstance struct {
	Messages []message `json:"messages"`
}

type message struct {
	Author  string `json:"author"`
	Content string `json:"content"`
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

func getPrompt(messages []types.Message) (string, error) {
	if l := len(messages); l != 1 {
		return "", errors.Errorf("expected to receive exactly one message with the prompt (got %d)", l)
	}

	return messages[0].Text, nil
}

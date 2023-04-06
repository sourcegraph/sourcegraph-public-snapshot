package resolvers

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.CompletionsResolver = &completionsResolver{}

type completionsResolver struct {
}

func NewCompletionsResolver() graphqlbackend.CompletionsResolver {
	return &completionsResolver{}
}

func (c *completionsResolver) Completions(ctx context.Context, args graphqlbackend.CompletionsArgs) (string, error) {

	completionsConfig := conf.Get().Completions
	if completionsConfig == nil || !completionsConfig.Enabled {
		return "", errors.New("completions are not configured or disabled")
	}

	client, err := streaming.GetCompletionStreamClient(completionsConfig.Provider, completionsConfig.AccessToken, completionsConfig.Model)
	if err != nil {
		return "", err
	}

	var last types.CompletionEvent

	if err := client.Stream(ctx, convertParams(args), func(event types.CompletionEvent) error {
		fmt.Println(event.Completion)
		last = event
		return nil
	}); err != nil {
		return "", err
	}
	return last.Completion, nil

	// return fmt.Sprintf("%v", args), nil
}

func convertParams(args graphqlbackend.CompletionsArgs) types.CompletionRequestParameters {
	return types.CompletionRequestParameters{
		Messages:          convertMessages(args.Input.Messages),
		Temperature:       float32(args.Input.Temperature),
		MaxTokensToSample: int(args.Input.MaxTokensToSample),
		TopK:              int(args.Input.TopK),
		TopP:              float32(args.Input.TopP),
	}
}

func convertMessages(messages []graphqlbackend.Message) (result []types.Message) {
	for _, message := range messages {
		result = append(result, types.Message{
			Speaker: message.Speaker,
			Text:    message.Text,
		})
	}
	return result
}

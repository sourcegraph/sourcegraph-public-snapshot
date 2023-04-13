package resolvers

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.CompletionsResolver = &completionsResolver{}

// completionsResolver provides chat completions
type completionsResolver struct {
}

func NewCompletionsResolver() graphqlbackend.CompletionsResolver {
	return &completionsResolver{}
}

func (c *completionsResolver) Completions(ctx context.Context, args graphqlbackend.CompletionsArgs) (string, error) {
	if envvar.SourcegraphDotComMode() {
		isEnabled := cody.IsCodyExperimentalFeatureFlagEnabled(ctx)
		if !isEnabled {
			return "", errors.New("cody experimental feature flag is not enabled for current user")
		}
	}

	completionsConfig := conf.Get().Completions
	if completionsConfig == nil || !completionsConfig.Enabled {
		return "", errors.New("completions are not configured or disabled")
	}

	client, err := streaming.GetCompletionClient(completionsConfig.Provider, completionsConfig.AccessToken, completionsConfig.Model)
	if err != nil {
		return "", errors.Wrap(err, "GetCompletionStreamClient")
	}

	var last string
	if err := client.Stream(ctx, convertParams(args), func(event types.ChatCompletionEvent) error {
		// each completion is just a partial of the final result, since we're in a sync request anyway
		// we will just wait for the final completion event
		last = event.Completion
		return nil
	}); err != nil {
		return "", errors.Wrap(err, "client.Stream")
	}
	return last, nil
}

func convertParams(args graphqlbackend.CompletionsArgs) types.ChatCompletionRequestParameters {
	return types.ChatCompletionRequestParameters{
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
			Speaker: strings.ToLower(message.Speaker),
			Text:    message.Text,
		})
	}
	return result
}

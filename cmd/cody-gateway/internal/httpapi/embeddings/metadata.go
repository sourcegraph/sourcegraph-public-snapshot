package embeddings

import (
	"context"

	"github.com/sourcegraph/conc/iter"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/fireworks"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
)

type metadataClient struct {
	logger            log.Logger
	completionsClient types.CompletionsClient
	ctx               context.Context
}

func generateMetadata(ctx context.Context, req codygateway.EmbeddingsRequest, logger log.Logger, completionsClient types.CompletionsClient) ([]string, error) {
	c := metadataClient{
		logger:            logger.Scoped("metadata_gen"),
		completionsClient: completionsClient,
		ctx:               ctx,
	}

	mapper := iter.Mapper[string, string]{MaxGoroutines: 15}
	return mapper.MapErr(req.Input, c.generateMetadataForChunk)
}

func (c *metadataClient) generateMetadataForChunk(input *string) (string, error) {
	promptText := `Here is a section of code.
Please write a paragraph of documentation for each high-level class, struct, function or similar.
Be concise, write no more than a few sentences for each entry.
Return your response in text format. Each entry name should be followed by a newline, then its documentation.
Respond with nothing else, only the entry names and the documentation. Code: ` +
		"```\n" + *input + "\n```"

	compRequest := types.CompletionRequest{
		Feature: types.CompletionsFeatureChat,
		Version: types.CompletionsVersionLegacy,
		Parameters: types.CompletionRequestParameters{
			Messages: []types.Message{{
				Speaker: "user",
				Text:    promptText,
			}},
			MaxTokensToSample: 2000,
			Temperature:       0,
			TopP:              1,
			RequestedModel:    fireworks.Llama38bInstruct,
		},
	}
	resp, err := c.completionsClient.Complete(c.ctx, c.logger, compRequest)

	if err != nil {
		return "", err
	}

	return resp.Completion, nil
}

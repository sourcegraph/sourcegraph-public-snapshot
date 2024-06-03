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
}

func GenerateMetadata(req codygateway.EmbeddingsRequest, logger log.Logger, completionsClient types.CompletionsClient) (codygateway.EmbeddingsRequest, error) {
	c := metadataClient{
		logger:            logger.Scoped("metadata_gen"),
		completionsClient: completionsClient,
	}

	generated, err := iter.MapErr(req.Input, c.generateMetadataForChunk)
	if err != nil {
		logger.Error("failed to generate metadata", log.Error(err))
		return codygateway.EmbeddingsRequest{}, err
	}

	metadataReq := req
	metadataReq.Input = generated
	return metadataReq, nil
}

func (c *metadataClient) generateMetadataForChunk(input *string) (string, error) {
	promptText := "Here is a section of code. " +
		"Please write a paragraph of documentation for each high-level class, struct, function or similar. " +
		"Be concise, write no more than a few sentences for each entry. " +
		"Return your response in text format. Each entry name should be followed by a newline, then its documentation. " +
		"Respond with nothing else, only the entry names and the documentation. Code: ```" +
		*input +
		"```"

	resp, err := c.completionsClient.Complete(context.TODO(), types.CompletionsFeatureChat, types.CompletionsVersionLegacy, types.CompletionRequestParameters{
		Messages: []types.Message{{
			Speaker: "user",
			Text:    promptText,
		}},
		MaxTokensToSample: 2000,
		Temperature:       0,
		TopP:              1,
		Model:             fireworks.Llama370bInstruct,
	}, c.logger)

	if err != nil {
		return "", err
	}

	return resp.Completion, nil
}

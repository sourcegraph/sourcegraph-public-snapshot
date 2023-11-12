package azureopenai

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed/client"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed/client/modeltransformations"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// We want to reuse the client because when using the DefaultAzureCredential
// it will acquire a short lived token and reusing the client
// prevents acquiring a new token on every request.
// The client will refresh the token as needed.
var azureClient *azopenai.Client

func getClient(endpoint, accessToken string) (*azopenai.Client, error) {
	if azureClient != nil {
		return azureClient, nil
	}
	var err error
	if accessToken != "" {
		credential, credErr := azopenai.NewKeyCredential(accessToken)
		if credErr != nil {
			return nil, credErr
		}
		azureClient, err = azopenai.NewClientWithKeyCredential(endpoint, credential, nil)
	} else {
		credential, credErr := azidentity.NewDefaultAzureCredential(nil)
		if credErr != nil {
			return nil, credErr
		}
		azureClient, err = azopenai.NewClient(endpoint, credential, nil)
	}
	return azureClient, err
}

func NewClient(httpClient *http.Client, config *conftypes.EmbeddingsConfig) (*azureOpenaiEmbeddingsClient, error) {
	client, err := getClient(config.Endpoint, config.AccessToken)
	if err != nil {
		return nil, err
	}
	return &azureOpenaiEmbeddingsClient{
		client:      client,
		dimensions:  config.Dimensions,
		accessToken: config.AccessToken,
		model:       config.Model,
		endpoint:    config.Endpoint,
	}, nil
}

type azureOpenaiEmbeddingsClient struct {
	client      *azopenai.Client
	model       string
	dimensions  int
	endpoint    string
	accessToken string
}

func (c *azureOpenaiEmbeddingsClient) GetDimensions() (int, error) {
	if c.dimensions <= 0 {
		return 0, errors.New("invalid config for embeddings.dimensions, must be > 0")
	}
	return c.dimensions, nil
}

func (c *azureOpenaiEmbeddingsClient) GetModelIdentifier() string {
	return fmt.Sprintf("azure-openai/%s", c.model)
}

func (c *azureOpenaiEmbeddingsClient) GetQueryEmbedding(ctx context.Context, query string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, []string{modeltransformations.ApplyToQuery(query, c.GetModelIdentifier())})
}

func (c *azureOpenaiEmbeddingsClient) GetDocumentEmbeddings(ctx context.Context, documents []string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, modeltransformations.ApplyToDocuments(documents, c.GetModelIdentifier()))
}

// getEmbeddings tries to embed the given texts using the external service specified in the config.
func (c *azureOpenaiEmbeddingsClient) getEmbeddings(ctx context.Context, texts []string) (*client.EmbeddingsResults, error) {
	for _, text := range texts {
		if text == "" {
			// The OpenAI API will return an error if any of the strings in texts is an empty string,
			// so fail fast to avoid making tons of retryable requests.
			return nil, errors.New("cannot generate embeddings for an empty string")
		}
	}

	response, err := c.client.GetEmbeddings(ctx, azopenai.EmbeddingsOptions{
		Input:      texts,
		Deployment: c.model,
	}, nil)

	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, nil
	}

	// Ensure embedding responses are sorted in the original order.
	sort.Slice(response.Data, func(i, j int) bool {
		return *response.Data[i].Index < *response.Data[j].Index
	})

	dimensionality := len(response.Data[0].Embedding)
	embeddings := make([]float32, 0, len(response.Data)*dimensionality)
	failed := make([]int, 0)
	for _, embedding := range response.Data {
		if len(embedding.Embedding) != 0 {
			embeddings = append(embeddings, embedding.Embedding...)
		} else {
			// HACK(camdencheek): Nondeterministically, the OpenAI API will
			// occasionally send back a `null` for an embedding in the
			// response. Try it again a few times and hope for the best.
			resp, err := c.requestSingleEmbeddingWithRetryOnNull(ctx, texts[*embedding.Index], 3)
			if err != nil {
				failed = append(failed, int(*embedding.Index))

				// reslice to provide zero value embedding for failed chunk
				embeddings = embeddings[:len(embeddings)+dimensionality]
				continue
			}
			embeddings = append(embeddings, resp.Data[0].Embedding...)
		}
	}

	return &client.EmbeddingsResults{Embeddings: embeddings, Failed: failed, Dimensions: dimensionality}, nil
}

func (c *azureOpenaiEmbeddingsClient) requestSingleEmbeddingWithRetryOnNull(ctx context.Context, input string, retries int) (*azopenai.GetEmbeddingsResponse, error) {
	for i := 0; i < retries; i++ {
		response, err := c.client.GetEmbeddings(ctx, azopenai.EmbeddingsOptions{
			Input:      []string{input},
			Deployment: c.model,
		}, nil)

		if err != nil {
			return nil, err
		}
		if len(response.Data) != 1 || len(response.Data[0].Embedding) == 0 {
			continue
		}
		return &response, nil
	}
	return nil, errors.Newf("null response for embedding after %d retries", retries)
}

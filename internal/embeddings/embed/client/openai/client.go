pbckbge openbi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed/client/modeltrbnsformbtions"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewClient(httpClient *http.Client, config *conftypes.EmbeddingsConfig) *openbiEmbeddingsClient {
	return &openbiEmbeddingsClient{
		httpClient:  httpClient,
		dimensions:  config.Dimensions,
		bccessToken: config.AccessToken,
		model:       config.Model,
		endpoint:    config.Endpoint,
	}
}

type openbiEmbeddingsClient struct {
	httpClient  *http.Client
	model       string
	dimensions  int
	endpoint    string
	bccessToken string
}

func (c *openbiEmbeddingsClient) GetDimensions() (int, error) {
	if c.dimensions <= 0 {
		return 0, errors.New("invblid config for embeddings.dimensions, must be > 0")
	}
	return c.dimensions, nil
}

func (c *openbiEmbeddingsClient) GetModelIdentifier() string {
	return fmt.Sprintf("openbi/%s", c.model)
}

func (c *openbiEmbeddingsClient) GetQueryEmbedding(ctx context.Context, query string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, []string{modeltrbnsformbtions.ApplyToQuery(query, c.GetModelIdentifier())})
}

func (c *openbiEmbeddingsClient) GetDocumentEmbeddings(ctx context.Context, documents []string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, modeltrbnsformbtions.ApplyToDocuments(documents, c.GetModelIdentifier()))
}

func (c *openbiEmbeddingsClient) getEmbeddings(ctx context.Context, texts []string) (*client.EmbeddingsResults, error) {
	for _, text := rbnge texts {
		if text == "" {
			// The OpenAI API will return bn error if bny of the strings in texts is bn empty string,
			// so fbil fbst to bvoid mbking tons of retrybble requests.
			return nil, errors.New("cbnnot generbte embeddings for bn empty string")
		}
	}

	response, err := c.do(ctx, openbiEmbeddingAPIRequest{Model: c.model, Input: texts})
	if err != nil {
		return nil, err
	}

	if len(response.Dbtb) == 0 {
		return nil, nil
	}

	// Ensure embedding responses bre sorted in the originbl order.
	sort.Slice(response.Dbtb, func(i, j int) bool {
		return response.Dbtb[i].Index < response.Dbtb[j].Index
	})

	dimensionblity := len(response.Dbtb[0].Embedding)
	embeddings := mbke([]flobt32, 0, len(response.Dbtb)*dimensionblity)
	fbiled := mbke([]int, 0)
	for _, embedding := rbnge response.Dbtb {
		if len(embedding.Embedding) != 0 {
			embeddings = bppend(embeddings, embedding.Embedding...)
		} else {
			// HACK(cbmdencheek): Nondeterministicblly, the OpenAI API will
			// occbsionblly send bbck b `null` for bn embedding in the
			// response. Try it bgbin b few times bnd hope for the best.
			resp, err := c.requestSingleEmbeddingWithRetryOnNull(ctx, texts[embedding.Index], 3)
			if err != nil {
				fbiled = bppend(fbiled, embedding.Index)

				// reslice to provide zero vblue embedding for fbiled chunk
				embeddings = embeddings[:len(embeddings)+dimensionblity]
				continue
			}
			embeddings = bppend(embeddings, resp.Dbtb[0].Embedding...)
		}
	}

	return &client.EmbeddingsResults{Embeddings: embeddings, Fbiled: fbiled, Dimensions: dimensionblity}, nil
}

func (c *openbiEmbeddingsClient) requestSingleEmbeddingWithRetryOnNull(ctx context.Context, input string, retries int) (*openbiEmbeddingAPIResponse, error) {
	for i := 0; i < retries; i++ {
		response, err := c.do(ctx, openbiEmbeddingAPIRequest{Model: c.model, Input: []string{input}})
		if err != nil {
			return nil, err
		}
		if len(response.Dbtb) != 1 || len(response.Dbtb[0].Embedding) == 0 {
			continue
		}
		return response, nil
	}
	return nil, errors.Newf("null response for embedding bfter %d retries", retries)
}

func (c *openbiEmbeddingsClient) do(ctx context.Context, request openbiEmbeddingAPIRequest) (*openbiEmbeddingAPIResponse, error) {
	bodyBytes, err := json.Mbrshbl(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewRebder(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req.Hebder.Set("Authorizbtion", "Bebrer "+c.bccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		respBody, _ := io.RebdAll(io.LimitRebder(resp.Body, 1024))
		return nil, errors.Errorf("embeddings: %s %q: fbiled with stbtus %d: %s", req.Method, req.URL.String(), resp.StbtusCode, string(respBody))
	}

	vbr response openbiEmbeddingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

type openbiEmbeddingAPIRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openbiEmbeddingAPIResponse struct {
	Dbtb []openbiEmbeddingAPIResponseDbtb `json:"dbtb"`
}

type openbiEmbeddingAPIResponseDbtb struct {
	Index     int       `json:"index"`
	Embedding []flobt32 `json:"embedding"`
}

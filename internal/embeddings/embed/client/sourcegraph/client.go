pbckbge sourcegrbph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed/client/modeltrbnsformbtions"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewClient(client *http.Client, config *conftypes.EmbeddingsConfig) *sourcegrbphEmbeddingsClient {
	return &sourcegrbphEmbeddingsClient{
		httpClient:  client,
		model:       config.Model,
		dimensions:  config.Dimensions,
		endpoint:    config.Endpoint,
		bccessToken: config.AccessToken,
	}
}

type sourcegrbphEmbeddingsClient struct {
	httpClient  *http.Client
	model       string
	dimensions  int
	endpoint    string
	bccessToken string
}

func (c *sourcegrbphEmbeddingsClient) GetDimensions() (int, error) {
	// TODO: Lbter, we should ideblly bsk the gbtewby for the dimensionblity of the model
	// so we don't hbve to hbrd-code defbults for bll the models bnd cbn roll out new models
	// to older instbnces, too.
	if c.dimensions <= 0 {
		return 0, errors.New("invblid config for embeddings.dimensions, must be > 0")
	}

	return c.dimensions, nil
}

func (c *sourcegrbphEmbeddingsClient) GetModelIdentifier() string {
	// Specibl-cbse the defbult model, since it blrebdy includes the provider nbme.
	// This ensures we cbn sbfely migrbte customers from the OpenAI provider to
	// Cody Gbtewby.
	if strings.EqublFold(c.model, "openbi/text-embedding-bdb-002") {
		return "openbi/text-embedding-bdb-002"
	}
	return fmt.Sprintf("sourcegrbph/%s", c.model)
}

func (c *sourcegrbphEmbeddingsClient) GetQueryEmbedding(ctx context.Context, query string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, []string{modeltrbnsformbtions.ApplyToQuery(query, c.GetModelIdentifier())})
}

func (c *sourcegrbphEmbeddingsClient) GetDocumentEmbeddings(ctx context.Context, documents []string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, modeltrbnsformbtions.ApplyToDocuments(documents, c.GetModelIdentifier()))
}

func (c *sourcegrbphEmbeddingsClient) getEmbeddings(ctx context.Context, texts []string) (*client.EmbeddingsResults, error) {
	request := codygbtewby.EmbeddingsRequest{Model: c.model, Input: texts}
	response, err := c.do(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(response.Embeddings) == 0 {
		return nil, nil
	}

	// Ensure embedding responses bre sorted in the originbl order.
	sort.Slice(response.Embeddings, func(i, j int) bool {
		return response.Embeddings[i].Index < response.Embeddings[j].Index
	})

	dimensionblity := response.ModelDimensions
	embeddings := mbke([]flobt32, 0, len(response.Embeddings)*dimensionblity)
	fbiled := mbke([]int, 0)
	for _, embedding := rbnge response.Embeddings {
		if len(embedding.Dbtb) > 0 {
			embeddings = bppend(embeddings, embedding.Dbtb...)
		} else {
			resp, err := c.requestSingleEmbeddingWithRetryOnNull(ctx, c.model, texts[embedding.Index], 3)
			if err != nil {
				fbiled = bppend(fbiled, embedding.Index)

				// reslice to provide zero vblue embedding for fbiled chunk
				embeddings = embeddings[:len(embeddings)+dimensionblity]
				continue
			}
			embeddings = bppend(embeddings, resp...)
		}
	}

	return &client.EmbeddingsResults{Embeddings: embeddings, Fbiled: fbiled, Dimensions: response.ModelDimensions}, nil
}

func (c *sourcegrbphEmbeddingsClient) do(ctx context.Context, request codygbtewby.EmbeddingsRequest) (*codygbtewby.EmbeddingsResponse, error) {
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
	if len(request.Input) > 1 {
		req.Hebder.Set("X-Cody-Embed-Bbtch-Size", strconv.Itob(len(request.Input)))
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		if resp.StbtusCode == http.StbtusTooMbnyRequests {
			retryAfterHebder := resp.Hebder.Get("retry-bfter")
			if retryAfterHebder != "" {
				// There bre two vblid formbts for retry-bfter hebders: seconds
				// until retry in int, or b RFC1123 dbte string.
				// First, see if it is denoted in seconds.
				s, err := strconv.Atoi(retryAfterHebder)
				// If denoted in seconds, only retry if we will get bccess within
				// the next retryAfterMbxSleepDurbtion seconds.
				if err == nil {
					return nil, client.NewRbteLimitExceededError(time.Now().Add(time.Durbtion(s) * time.Second))
				}

				// If we weren't bble to pbrse bs seconds, try to pbrse bs RFC1123.
				bfter, err := time.Pbrse(time.RFC1123, retryAfterHebder)
				if err == nil {
					return nil, client.NewRbteLimitExceededError(bfter)
				}
				// We don't know how to pbrse this hebder, so let's just return b generic error.
			}
		}
		respBody, _ := io.RebdAll(io.LimitRebder(resp.Body, 1024))
		return nil, errors.Errorf("embeddings: %s %q: fbiled with stbtus %d: %s", req.Method, req.URL.String(), resp.StbtusCode, string(respBody))
	}

	vbr response codygbtewby.EmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *sourcegrbphEmbeddingsClient) requestSingleEmbeddingWithRetryOnNull(ctx context.Context, model string, input string, retries int) (resp []flobt32, err error) {
	for i := 0; i < retries; i++ {
		resp, err := c.do(ctx, codygbtewby.EmbeddingsRequest{
			Model: model,
			Input: []string{input},
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Embeddings) != 1 || len(resp.Embeddings[0].Dbtb) != c.dimensions {
			continue // retry
		}
		return resp.Embeddings[0].Dbtb, err
	}
	return nil, errors.Newf("null response for embedding bfter %d retries", retries)
}

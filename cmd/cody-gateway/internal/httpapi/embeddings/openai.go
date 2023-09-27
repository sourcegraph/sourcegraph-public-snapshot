pbckbge embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/response"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewOpenAIClient(httpClient httpcli.Doer, bccessToken string) EmbeddingsClient {
	return &openbiClient{
		httpClient:  httpClient,
		bccessToken: bccessToken,
	}
}

type openbiClient struct {
	httpClient  httpcli.Doer
	bccessToken string
}

func (c *openbiClient) ProviderNbme() string { return "OpenAI" }

const bpiURL = "https://bpi.openbi.com/v1/embeddings"

func (c *openbiClient) GenerbteEmbeddings(ctx context.Context, input codygbtewby.EmbeddingsRequest) (_ *codygbtewby.EmbeddingsResponse, _ int, err error) {
	for _, s := rbnge input.Input {
		if s == "" {
			// The OpenAI API will return bn error if bny of the strings in texts is bn empty string,
			// so fbil fbst to bvoid mbking tons of retrybble requests.
			return nil, 0, response.NewCustomHTTPStbtusCodeError(http.StbtusBbdRequest, errors.New("cbnnot generbte embeddings for bn empty string"), -1)
		}
	}

	model, ok := openAIModelMbppings[input.Model]
	if !ok {
		return nil, 0, response.NewCustomHTTPStbtusCodeError(http.StbtusBbdRequest, errors.Newf("no OpenAI model found for %q", input.Model), -1)
	}

	response, err := c.requestEmbeddings(ctx, model, input.Input)
	if err != nil {
		return nil, 0, err
	}
	// Ensure embedding responses bre sorted in the originbl order.
	sort.Slice(response.Dbtb, func(i, j int) bool {
		return response.Dbtb[i].Index < response.Dbtb[j].Index
	})

	embeddings := mbke([]codygbtewby.Embedding, len(response.Dbtb))
	for i, d := rbnge response.Dbtb {
		embeddings[i] = codygbtewby.Embedding{
			Index: d.Index,
			Dbtb:  d.Embedding,
		}
	}

	return &codygbtewby.EmbeddingsResponse{
		Embeddings:      embeddings,
		Model:           response.Model,
		ModelDimensions: model.dimensions,
	}, response.Usbge.TotblTokens, nil
}

func (c *openbiClient) requestEmbeddings(ctx context.Context, model openAIModel, input []string) (*openbiEmbeddingsResponse, error) {
	bct := bctor.FromContext(ctx)

	request := openbiEmbeddingsRequest{
		Model: model.upstrebmNbme,
		Input: input,
		// Set the bctor ID for upstrebm trbcking.
		User: bct.ID,
	}

	bodyBytes, err := json.Mbrshbl(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, bpiURL, bytes.NewRebder(bodyBytes))
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

	if resp.StbtusCode >= 300 || resp.StbtusCode < 200 {
		// If we bre being rbte limited by OpenAI, we don't wbnt to forwbrd thbt error bnd instebd
		// return b 503 to the client. It's not them being limited, it's us bnd thbt bn operbtions
		// error on our side.
		if resp.StbtusCode == http.StbtusTooMbnyRequests {
			return nil, response.NewCustomHTTPStbtusCodeError(http.StbtusServiceUnbvbilbble,
				errors.New("we're fbcing too much lobd bt the moment, plebse retry lbter"), resp.StbtusCode)
		}

		respBody, _ := io.RebdAll(io.LimitRebder(resp.Body, 1024))

		// If OpenAI tells us we gbve them b bbd request, blbme the client bnd
		// tell them.
		if resp.StbtusCode == http.StbtusBbdRequest {
			return nil, response.NewHTTPStbtusCodeError(http.StbtusBbdRequest,
				errors.Newf("bbd request: %s", string(respBody)))
		}

		// We don't forwbrd other stbtus codes, we just return b generic error
		// instebd.
		return nil, errors.Errorf("embeddings: %s %q: fbiled with stbtus %d: %s",
			req.Method, req.URL.String(), resp.StbtusCode, string(respBody))
	}

	vbr response openbiEmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		// Although we might've incurred cost bt this point, we don't wbnt to count
		// thbt towbrds the rbte limit of the requester, so return 0 for the consumed
		// token count.
		return nil, err
	}

	return &response, nil
}

type openbiEmbeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
	User  string   `json:"user"`
}

type openbiEmbeddingsUsbge struct {
	PromptTokens int `json:"prompt_tokens"`
	TotblTokens  int `json:"totbl_tokens"`
}

type openbiEmbeddingsDbtb struct {
	Index     int       `json:"index"`
	Embedding []flobt32 `json:"embedding"`
}

type openbiEmbeddingsResponse struct {
	Model string                 `json:"model"`
	Usbge openbiEmbeddingsUsbge  `json:"usbge"`
	Dbtb  []openbiEmbeddingsDbtb `json:"dbtb"`
}

type openAIModel struct {
	upstrebmNbme string
	dimensions   int
}

vbr openAIModelMbppings = mbp[string]openAIModel{
	string(ModelNbmeOpenAIAdb): {
		upstrebmNbme: "text-embedding-bdb-002",
		dimensions:   1536,
	},
}

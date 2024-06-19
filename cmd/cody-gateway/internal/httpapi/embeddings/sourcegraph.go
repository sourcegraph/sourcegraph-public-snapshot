package embeddings

import (
	"bytes"
	"context"
	"io"
	"math"
	// nosemgrep: security-semgrep-rules.semgrep-rules.golang.math-random-used
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

const (
	modelDimensions           = 768
	inferenceSecretHeaderName = "sourcegraph-smega-inference-auth"
)

var (
	// When our backend service (Triton) scales down, the following events happen within a few seconds of each other:
	// 1. Triton gets a signal to shut-down
	// 2. Triton waits for all in-flight requests to complete (usually <5s)
	// 3. Triton starts unloading models and returning 404 messages to new requests
	// 4. Triton exits, pod shuts down
	// The issue with that that Triton is not detached from the load-balancer/service between steps 1 and 3, so requests that are forwarded to the Pod after #1 but before #4 might get a 404 response.
	// Retrying a request on the client side will make it go to a random backend, so, we will likely hit a different Pod (that is not shutting down).
	triton404Retries = 3
)

func NewSourcegraphClient(httpClient httpcli.Doer, apiURL string, apiToken string) EmbeddingsClient {
	return &sourcegraphClient{
		httpClient: httpClient,
		json:       jsoniter.ConfigCompatibleWithStandardLibrary,
		apiURL:     apiURL,
		apiToken:   apiToken,
	}
}

type sourcegraphClient struct {
	httpClient httpcli.Doer
	json       jsoniter.API
	apiURL     string
	apiToken   string
}

func (s sourcegraphClient) ProviderName() string {
	return "Sourcegraph"
}

// GenerateEmbeddings uses a KServe-compatible API to generate embedding vectors for items from request.Input
func (s sourcegraphClient) GenerateEmbeddings(ctx context.Context, request codygateway.EmbeddingsRequest) (*codygateway.EmbeddingsResponse, int, error) {
	items := len(request.Input)
	input := kserveInput{
		Name:     "TEXT",
		Shape:    []int{items},
		Datatype: "BYTES",
		Data:     request.Input,
	}
	smegaRequest := kserveRequest{
		ID:     uuid.New().String(),
		Inputs: []kserveInput{input},
	}
	dat, err := s.json.Marshal(smegaRequest)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(dat))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set(inferenceSecretHeaderName, s.apiToken)

	var lastErr error
	for i := 0; i < triton404Retries; i++ {
		if res, lastErr := s.fetch(req, request.Model, len(request.Input)); lastErr == nil {
			return res, 0, nil
		}
		time.Sleep(backoffInterval(i))
	}

	return nil, 0, lastErr
}

func backoffInterval(attempt int) time.Duration {
	// exponential backoff
	dur := time.Duration(math.Pow(5, float64(attempt))) * time.Millisecond
	// jitter
	dur += time.Duration(rand.Int32N(int32(dur))) - dur/2
	return dur
}

func (s sourcegraphClient) fetch(req *http.Request, model string, expectedItems int) (*codygateway.EmbeddingsResponse, error) {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if err != nil {
			return nil, errors.Newf("unexpected status code: %d, failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, errors.Newf("unexpected status code: %d, response: %s", resp.StatusCode, respBody)
	}

	var smegaResponse kserveResponse
	err = s.json.NewDecoder(resp.Body).Decode(&smegaResponse)
	if err != nil {
		return nil, err
	}

	res := codygateway.EmbeddingsResponse{
		Embeddings:      make([]codygateway.Embedding, expectedItems),
		Model:           model,
		ModelDimensions: modelDimensions,
	}
	if len(smegaResponse.Outputs) == 0 {
		return nil, errors.New("no outputs returned")
	}
	for i := 0; i*modelDimensions < len(smegaResponse.Outputs[0].Data); i++ {
		tensor := smegaResponse.Outputs[0].Data[i*modelDimensions : (i+1)*modelDimensions]
		res.Embeddings[i] = codygateway.Embedding{
			Data:  tensor,
			Index: i,
		}
	}
	return &res, nil
}

// Based on https://github.com/kserve/kserve/blob/master/docs/predict-api/v2/required_api.md#inference-request-json-object
type kserveRequest struct {
	ID     string        `json:"id"`
	Inputs []kserveInput `json:"inputs"`
}

// Based on https://github.com/kserve/kserve/blob/master/docs/predict-api/v2/required_api.md#request-input
type kserveInput struct {
	Name     string   `json:"name"`
	Shape    []int    `json:"shape"`
	Datatype string   `json:"datatype"`
	Data     []string `json:"data"`
}

// Based on https://github.com/kserve/kserve/blob/master/docs/predict-api/v2/required_api.md#inference-response-json-object
type kserveResponse struct {
	ID           string         `json:"id"`
	ModelName    string         `json:"model_name"`
	ModelVersion string         `json:"model_version"`
	Outputs      []kserveOutput `json:"outputs"`
}

// Based on https://github.com/kserve/kserve/blob/master/docs/predict-api/v2/required_api.md#response-output
type kserveOutput struct {
	Name     string    `json:"name"`
	Shape    []int     `json:"shape"`
	Datatype string    `json:"datatype"`
	Data     []float32 `json:"data"`
}

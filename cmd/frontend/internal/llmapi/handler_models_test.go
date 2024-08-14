package llmapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	types "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/internal/openapi/goapi"
)

func TestModelsHandler(t *testing.T) {

	mockModels := []types.Model{
		{
			ModelRef:    "anthropic::unknown::claude-3-sonnet-20240229",
			DisplayName: "Claude 3 Sonnet",
		},
		{
			ModelRef:    "openai::unknown::gpt-4",
			DisplayName: "GPT-4",
		},
	}
	c := newTest(t, func() (*types.ModelConfiguration, error) {
		return &types.ModelConfiguration{Models: mockModels}, nil
	})

	req, err := http.NewRequest("GET", "/.api/llm/models", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	c.Handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response goapi.ListModelsResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	autogold.Expect(goapi.ListModelsResponse{Object: "list", Data: []goapi.Model{
		{
			Id:      "anthropic::unknown::claude-3-sonnet-20240229",
			Object:  "model",
			OwnedBy: "anthropic",
		},
		{
			Id:      "openai::unknown::gpt-4",
			Object:  "model",
			OwnedBy: "openai",
		},
	}}).Equal(t, response)

}

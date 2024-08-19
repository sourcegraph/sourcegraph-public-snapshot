package llmapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	types "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/internal/openapi/goapi"
)

func TestModelsModelIDHandler(t *testing.T) {
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

	testCases := []struct {
		name           string
		modelID        string
		expectedStatus int
		expectedModel  *goapi.Model
	}{
		{
			name:           "Existing model",
			modelID:        "anthropic::unknown::claude-3-sonnet-20240229",
			expectedStatus: http.StatusOK,
			expectedModel: &goapi.Model{
				Id:      "anthropic::unknown::claude-3-sonnet-20240229",
				Object:  "model",
				OwnedBy: "anthropic",
			},
		},
		{
			name:           "Non-existent model",
			modelID:        "non-existent-model",
			expectedStatus: http.StatusNotFound,
			expectedModel:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/.api/llm/models/"+tc.modelID, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			c.Handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if rr.Code == http.StatusOK {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			}

			if tc.expectedModel != nil {
				var response goapi.Model
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, response, *tc.expectedModel)
			}
		})
	}
}

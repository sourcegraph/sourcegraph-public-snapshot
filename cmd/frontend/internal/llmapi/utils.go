package llmapi

import (
	"fmt"
	"net/http"

	types "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

type GetModelConfigurationFunc func() (*types.ModelConfiguration, error)

func modelConfigOr500Error(w http.ResponseWriter, getModelConfig GetModelConfigurationFunc) (*types.ModelConfiguration, bool) {
	currentModelConfig, err := getModelConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("modelConfigSvc.Get: %v", err), http.StatusInternalServerError)
		return nil, true
	}
	return currentModelConfig, false
}

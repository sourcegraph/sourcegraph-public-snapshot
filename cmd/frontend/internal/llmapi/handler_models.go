package llmapi

import (
	"fmt"
	"net/http"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/openapi/goapi"
)

type modelsHandler struct {
	logger         sglog.Logger
	GetModelConfig GetModelConfigurationFunc
}

var _ http.Handler = (*modelsHandler)(nil)

func (m *modelsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	currentModelConfig, err := m.GetModelConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("modelConfigSvc.Get: %v", err), http.StatusInternalServerError)
		return
	}

	var data []goapi.Model
	for _, model := range currentModelConfig.Models {
		data = append(data, goapi.Model{
			Object:  "model",
			Id:      string(model.ModelRef),
			OwnedBy: string(model.ModelRef.ProviderID()),
		})
	}

	response := goapi.ListModelsResponse{
		Object: "list",
		Data:   data,
	}

	serveJSON(w, r, m.logger, response)
}

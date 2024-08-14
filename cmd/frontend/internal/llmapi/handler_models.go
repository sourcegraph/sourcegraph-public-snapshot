package llmapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sourcegraph/log"
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
	rawJSON, err := json.MarshalIndent(goapi.ListModelsResponse{
		Object: "list",
		Data:   data,
	}, "", "    ")
	if err != nil {
		m.logger.Error("marshalling ListModelsReponse", log.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(rawJSON); err != nil {
		http.Error(w, "writing response", http.StatusInternalServerError)
		return
	}
}

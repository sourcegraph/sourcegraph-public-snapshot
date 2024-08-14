package llmapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/openapi/goapi"
)

type modelsModelIDHandler struct {
	logger         sglog.Logger
	GetModelConfig GetModelConfigurationFunc
}

var _ http.Handler = (*modelsModelIDHandler)(nil)

func (m *modelsModelIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelId := vars["modelId"]

	if modelId == "" {
		http.Error(w, "modelId is required", http.StatusBadRequest)
		return
	}

	currentModelConfig, err := m.GetModelConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("modelConfigSvc.Get: %v", err), http.StatusInternalServerError)
		return
	}

	for _, model := range currentModelConfig.Models {
		if string(model.ModelRef) == modelId {
			response := goapi.Model{
				Object:  "model",
				Id:      string(model.ModelRef),
				OwnedBy: string(model.ModelRef.ProviderID()),
			}
			serveJSON(w, r, m.logger, response)
			return
		}
	}

	http.Error(w, "Model not found", http.StatusNotFound)
}

package llmapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	sglog "github.com/sourcegraph/log"

	models "github.com/sourcegraph/sourcegraph/internal/openapi/go"
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
			rawJSON, err := json.MarshalIndent(models.Model{
				Object:  "model",
				Id:      string(model.ModelRef),
				OwnedBy: string(model.ModelRef.ProviderID()),
			}, "", "    ")
			if err != nil {
				m.logger.Error("marshalling Model", log.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(rawJSON)
			if err != nil {
				http.Error(w, "writing response", http.StatusInternalServerError)
			}
			return
		}
	}

	http.Error(w, "Model not found", http.StatusNotFound)
}

package llmapi

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"
	sglog "github.com/sourcegraph/log"

	models "github.com/sourcegraph/sourcegraph/internal/openapi/go"
)

type modelsHandler struct {
	logger         sglog.Logger
	GetModelConfig GetModelConfigurationFunc
}

var _ http.Handler = (*modelsHandler)(nil)

func (m *modelsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	currentModelConfig, shouldReturn := modelConfigOr500Error(w, m.GetModelConfig)
	if shouldReturn {
		return
	}

	var data []models.Model
	for _, model := range currentModelConfig.Models {
		data = append(data, models.Model{
			Object:  "model",
			Id:      string(model.ModelRef),
			OwnedBy: string(model.ModelRef.ProviderID()),
		})
	}
	rawJSON, err := json.MarshalIndent(models.ListModelsResponse{
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

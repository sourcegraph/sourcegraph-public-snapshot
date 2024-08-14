package llmapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	sglog "github.com/sourcegraph/log"

	types "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

type GetModelConfigurationFunc func() (*types.ModelConfiguration, error)

func RegisterHandlers(m *mux.Router, apiHandler http.Handler, getModelConfigFunc GetModelConfigurationFunc) {
	logger := sglog.Scoped("llmapi")

	m.Path("/chat/completions").Methods("POST").Handler(&chatCompletionsHandler{
		logger:     logger,
		apiHandler: apiHandler,
	})
	m.Path("/models").Methods("GET").Handler(&modelsHandler{
		logger:         logger,
		GetModelConfig: getModelConfigFunc,
	})
	m.Path("/models/{modelId}").Methods("GET").Handler(&modelsModelIDHandler{
		logger:         logger,
		GetModelConfig: getModelConfigFunc,
	})
}

func serveJSON(w http.ResponseWriter, r *http.Request, logger sglog.Logger, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Error(fmt.Sprintf("writing %s JSON response body", r.URL.Path), log.Error(err))
		http.Error(w, "writing response", http.StatusInternalServerError)
	}
}

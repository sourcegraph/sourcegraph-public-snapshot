package llmapi

import (
	"net/http"

	"github.com/gorilla/mux"
	sglog "github.com/sourcegraph/log"

	types "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

type GetModelConfigurationFunc func() (*types.ModelConfiguration, error)

func RegisterHandlers(m *mux.Router, apiHandler http.Handler, getModelConfigFunc GetModelConfigurationFunc) {
	logger := sglog.Scoped("llmapi")

	m.Path("/chat/completions").Methods("POST").Handler(&chatCompletionsHandler{
		logger:         logger,
		apiHandler:     apiHandler,
		GetModelConfig: getModelConfigFunc,
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

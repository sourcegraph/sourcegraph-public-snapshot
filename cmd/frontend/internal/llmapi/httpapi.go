package llmapi

import (
	"net/http"

	"github.com/gorilla/mux"
	sglog "github.com/sourcegraph/log"
)

func RegisterHandlers(m *mux.Router, apiHandler http.Handler) {
	logger := sglog.Scoped("llmapi")

	m.Path("/chat/completions").Methods("POST").Handler(&chatCompletionsHandler{logger: logger, apiHandler: apiHandler})
}

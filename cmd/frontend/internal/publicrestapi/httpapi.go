package publicrestapi

import (
	"net/http"

	"github.com/gorilla/mux"
	sglog "github.com/sourcegraph/log"
)

func NewHandler(apiHandler http.Handler) http.Handler {
	logger := sglog.Scoped("publicrestapi")

	m := mux.NewRouter()

	m.Path("/api/v1/chat/completions").Methods("POST").Handler(&chatCompletionsHandler{logger: logger, apiHandler: apiHandler})

	return m
}

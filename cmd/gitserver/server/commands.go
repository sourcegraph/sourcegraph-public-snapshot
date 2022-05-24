package server

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

func handleGetObject(getObject gitdomain.GetObjectFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req protocol.GetObjectRequest
		logger := log.Scoped("handleGetObject", "handles get object")

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "decoding body", http.StatusBadRequest)
			logger.Error("handleGetObject: decoding body", log.Error(err))
			return
		}

		obj, err := getObject(r.Context(), req.Repo, req.ObjectName)
		if err != nil {
			http.Error(w, "getting object", http.StatusInternalServerError)
			logger.Error("handleGetObject: getting object", log.Error(err))
			return
		}

		resp := protocol.GetObjectResponse{
			Object: *obj,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("handleGetObject: sending response", log.Error(err))
		}
	}
}

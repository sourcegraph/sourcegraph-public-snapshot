package server

import (
	"encoding/json"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func handleGetObject(getObject gitdomain.GetObjectFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req protocol.GetObjectRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "decoding body", http.StatusBadRequest)
			log15.Error("handleGetObject: decoding body", "error", err)
			return
		}

		obj, err := getObject(r.Context(), req.Repo, req.ObjectName)
		if err != nil {
			http.Error(w, "getting object", http.StatusInternalServerError)
			log15.Error("handleGetObject: getting object", "error", err)
			return
		}

		resp := protocol.GetObjectResponse{
			Object: *obj,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log15.Error("handleGetObject: sending response", "error", err)
		}
	}
}

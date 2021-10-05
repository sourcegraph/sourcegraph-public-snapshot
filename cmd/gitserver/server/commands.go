package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/domain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func (s *Server) handleGetObject(svc domain.GetObjectService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req protocol.GetObjectRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "reading body", http.StatusBadRequest)
			log15.Error("handleGetObject: reading body", "error", err)
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "decoding body", http.StatusBadRequest)
			log15.Error("handleGetObject: decoding body", "error", err)
			return
		}

		obj, err := svc.GetObject(r.Context(), req.Repo, req.ObjectName)
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

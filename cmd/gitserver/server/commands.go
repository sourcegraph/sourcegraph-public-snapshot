package server

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/internal/accesslog"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func handleGetObject(logger log.Logger, getObject gitdomain.GetObjectFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req protocol.GetObjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "decoding body", http.StatusBadRequest)
			logger.Error("decoding body", log.Error(err))
			return
		}

		// Log which actor is accessing the repo.
		accesslog.Record(r.Context(), string(req.Repo), map[string]string{
			"objectname": req.ObjectName,
		})

		obj, err := getObject(r.Context(), req.Repo, req.ObjectName)
		if err != nil {
			http.Error(w, "getting object", http.StatusInternalServerError)
			logger.Error("getting object", log.Error(err))
			return
		}

		resp := protocol.GetObjectResponse{
			Object: *obj,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("sending response", log.Error(err))
		}
	}
}

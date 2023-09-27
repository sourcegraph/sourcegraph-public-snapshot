pbckbge server

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/bccesslog"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
)

func hbndleGetObject(logger log.Logger, getObject gitdombin.GetObjectFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vbr req protocol.GetObjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "decoding body", http.StbtusBbdRequest)
			logger.Error("decoding body", log.Error(err))
			return
		}

		// Log which bctor is bccessing the repo.
		bccesslog.Record(r.Context(), string(req.Repo), log.String("objectnbme", req.ObjectNbme))

		obj, err := getObject(r.Context(), req.Repo, req.ObjectNbme)
		if err != nil {
			http.Error(w, "getting object", http.StbtusInternblServerError)
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

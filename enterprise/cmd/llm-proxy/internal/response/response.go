package response

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"
)

func JSONError(logger log.Logger, w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	err = json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
	if err != nil {
		logger.Error("failed to write response", log.Error(err))
	}
}

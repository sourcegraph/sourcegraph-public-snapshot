package response

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"
)

// JSONError writes an error response in JSON format. If the status code is 5xx,
// the error is logged as well, since it's not fun for clients to receive 5xx and
// we should record it for investigation.
//
// The logger should have trace and actor information attached where relevant.
func JSONError(logger log.Logger, w http.ResponseWriter, code int, err error) {
	if code >= 500 {
		logger.Error(http.StatusText(code), log.Error(err))
	}
	w.WriteHeader(code)
	if encodeErr := json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	}); encodeErr != nil {
		logger.Error("failed to write response", log.Error(encodeErr))
	}
}

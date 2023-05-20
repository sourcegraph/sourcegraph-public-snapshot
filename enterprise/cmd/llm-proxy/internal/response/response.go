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

type StatusHeaderRecorder struct {
	StatusCode int
	http.ResponseWriter
}

func NewStatusHeaderRecorder(w http.ResponseWriter) *StatusHeaderRecorder {
	return &StatusHeaderRecorder{ResponseWriter: w}
}

// Write writes the data to the connection as part of an HTTP reply.
//
// If WriteHeader has not yet been called, Write calls
// WriteHeader(http.StatusOK) before writing the data.
func (r *StatusHeaderRecorder) Write(b []byte) (int, error) {
	if r.StatusCode == 0 {
		r.StatusCode = http.StatusOK // implicit behaviour of http.ResponseWriter
	}
	return r.ResponseWriter.Write(b)
}

// WriteHeader sends an HTTP response header with the provided status code and
// records the status code for later inspection.
func (r *StatusHeaderRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

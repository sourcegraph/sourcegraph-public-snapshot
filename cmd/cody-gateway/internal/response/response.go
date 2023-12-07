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
	} else if code >= 400 {
		// Generate logs for 4xx errors for debugging purposes
		logger.Debug(http.StatusText(code), log.Error(err))
	}

	w.WriteHeader(code)
	if encodeErr := json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	}); encodeErr != nil {
		logger.Error("failed to write response", log.Error(encodeErr))
	}
}

type StatusHeaderRecorder struct {
	StatusCode int
	http.ResponseWriter
	Logger log.Logger
}

var _ http.Flusher = &StatusHeaderRecorder{}

func NewStatusHeaderRecorder(w http.ResponseWriter, logger log.Logger) *StatusHeaderRecorder {
	return &StatusHeaderRecorder{ResponseWriter: w, Logger: logger}
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

func (r *StatusHeaderRecorder) Flush() {
	rc := http.NewResponseController(r.ResponseWriter)
	// we're implementing a stdlib interface that doesn't return the error, so we log
	if err := rc.Flush(); err != nil {
		r.Logger.Warn("flushing response failed", log.Error(err))
	}
}

// NewHTTPStatusCodeError records a status code error returned from a request.
func NewHTTPStatusCodeError(statusCode int, innerErr error) error {
	return HTTPStatusCodeError{
		status: statusCode,
		inner:  innerErr,
	}
}

// NewCustomHTTPStatusCodeError is an error that denotes a custom status code
// error. It is different from NewHTTPStatusCodeError as it indicates this isn't
// really an error from a request, but from something like custom validation.
func NewCustomHTTPStatusCodeError(statusCode int, innerErr error, originalCode int) error {
	return HTTPStatusCodeError{
		status:         statusCode,
		originalStatus: originalCode,
		inner:          innerErr,
		custom:         true,
	}
}

type HTTPStatusCodeError struct {
	status         int
	originalStatus int
	inner          error
	custom         bool
}

func (e HTTPStatusCodeError) Error() string { return e.inner.Error() }

func (e HTTPStatusCodeError) HTTPStatusCode() int { return e.status }

func (e HTTPStatusCodeError) IsCustom() (originalCode int, isCustom bool) {
	return e.originalStatus, e.custom
}

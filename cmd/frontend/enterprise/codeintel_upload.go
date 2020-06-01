package enterprise

import (
	"net/http"
)

type CodeIntelUploadHandlerFactory func(internal bool) http.Handler

var defaultHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("codeintel upload is only available in enterprise"))
})

// NewCodeIntelUploadHandler is re-assigned by the enterprise frontend
var NewCodeIntelUploadHandler CodeIntelUploadHandlerFactory = func(_ bool) http.Handler {
	return defaultHandler
}

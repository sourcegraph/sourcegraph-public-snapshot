package httpapi

import (
	"net/http"
)

// NewCodeIntelUploadHandler is set by the enterprise frontend
var NewCodeIntelUploadHandler func() http.Handler

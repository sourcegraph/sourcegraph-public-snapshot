package httpapi

import (
	"net/http"
)

type LSIFServerProxy struct {
	UploadHandler    http.Handler
	AllRoutesHandler http.Handler
}

// Set by enterprise frontend
var NewLSIFServerProxy func() (*LSIFServerProxy, error)

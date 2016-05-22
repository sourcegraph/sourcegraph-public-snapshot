package httpapi

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"
)

func serveBlackHole(w http.ResponseWriter, r *http.Request) error {
	// Status 410 Gone
	log15.Debug("BlackHole", "url", r.URL)
	w.WriteHeader(410)
	return nil
}

package webapp

import (
	_ "embed"
	"net/http"
)

func (h *Handler) serveRoot(w http.ResponseWriter, r *http.Request) {
	r2 := *r
	r2.URL.Path = "/"
	h.staticFiles.ServeHTTP(w, &r2)
}

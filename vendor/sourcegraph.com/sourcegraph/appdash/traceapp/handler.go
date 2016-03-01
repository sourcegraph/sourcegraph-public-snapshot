package traceapp

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

type handlerFunc func(http.ResponseWriter, *http.Request) error

// ServeHTTP implements http.Handler.
func (h handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var rb responseBuffer

	defer func() {
		if rv := recover(); rv != nil {
			handleError(w, r, fmt.Errorf("handler panic\n\n%s\n\n%s", rv, debug.Stack()))
		}
	}()

	err := h(&rb, r)
	if err != nil {
		handleError(w, r, err)
		return
	}
	rb.WriteTo(w)
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("%s %s: error: %s", r.Method, r.URL.RequestURI(), err.Error())

	// Never cache error responses.
	w.Header().Set("cache-control", "no-cache, max-age=0")

	http.Error(w, err.Error(), http.StatusInternalServerError)
}

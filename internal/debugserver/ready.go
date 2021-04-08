package debugserver

import "net/http"

// healthzHandler is the http.HandlerFunc that responds to /healthz
// requests on the debugserver port. This always returns a 200 OK
// while the binary can be reached.
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// readyHandler returns an http.HandlerFunc that responds to the /ready
// requests on the debugserver port. This will return a 200 OK once the
// given channel is closed, and a 503 Service Unavailable otherwise.
func readyHandler(ready <-chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ready:
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}

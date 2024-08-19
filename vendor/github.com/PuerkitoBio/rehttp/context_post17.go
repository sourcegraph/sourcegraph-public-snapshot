// +build go1.7

package rehttp

import "net/http"

func contextForRequest(req *http.Request) <-chan struct{} {
	// req.Context always returns non-nil.
	return req.Context().Done()
}

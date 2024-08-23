// +build !go1.7

package rehttp

import "net/http"

func contextForRequest(req *http.Request) <-chan struct{} {
	return nil
}

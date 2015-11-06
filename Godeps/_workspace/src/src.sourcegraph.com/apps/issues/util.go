package issues

import "net/http"

type passThrough struct {
	http.Handler
}

func (pt passThrough) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Sourcegraph-Verbatim", "true")
	pt.Handler.ServeHTTP(w, req)
}

package issues

import "net/http"

type passThrough struct {
	Handler  http.Handler
	Verbatim func(w http.ResponseWriter) // Verbatim shouldn't be nil.
}

func (pt passThrough) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pt.Verbatim(w)
	pt.Handler.ServeHTTP(w, req)
}

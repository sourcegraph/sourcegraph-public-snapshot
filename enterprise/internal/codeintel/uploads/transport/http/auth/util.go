package auth

import "net/http"

func hasQuery(r *http.Request, name string) bool {
	return r.URL.Query().Get(name) != ""
}

func getQuery(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

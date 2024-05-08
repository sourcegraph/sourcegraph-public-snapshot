package auth

import "net/http"

func getQuery(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

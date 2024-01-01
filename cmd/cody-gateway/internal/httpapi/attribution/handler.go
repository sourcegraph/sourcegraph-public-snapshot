package attribution

import "net/http"

func NewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok\n"))
		w.WriteHeader(http.StatusOK)
	})
}

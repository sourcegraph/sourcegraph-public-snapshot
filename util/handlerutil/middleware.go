package handlerutil

import "net/http"

type Middleware func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)

func WithMiddleware(h http.Handler, mw ...Middleware) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(mw) >= 2 {
			mw[0](w, r, WithMiddleware(h, mw[1:]...).ServeHTTP)
		} else if len(mw) == 1 {
			mw[0](w, r, h.ServeHTTP)
		} else if len(mw) == 0 {
			h.ServeHTTP(w, r)
		}
	})
}

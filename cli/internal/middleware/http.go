package middleware

import (
	"net"
	"net/http"
)

// RealIP sets req.RemoteAddr from the X-Real-Ip header if it exists.
func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s := r.Header.Get("X-Real-Ip"); s != "" && stripPort(r.RemoteAddr) == "127.0.0.1" {
			r.RemoteAddr = s
		}
		next.ServeHTTP(w, r)
	})
}

// stripPort removes the port specification from an address.
func stripPort(s string) string {
	if h, _, err := net.SplitHostPort(s); err == nil {
		s = h
	}
	return s
}

func StrictTransportSecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("strict-transport-security", "max-age=8640000")
		next.ServeHTTP(w, r)
	})
}

func SecureHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("x-xss-protection", "1; mode=block")
		w.Header().Set("x-frame-options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func NoCacheByDefault(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", "no-cache, max-age=0")
		next.ServeHTTP(w, r)
	})
}

package middleware

import (
	"net"
	"net/http"
)

// RealIP sets req.RemoteAddr from the X-Real-Ip header if it exists.
func RealIP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if s := r.Header.Get("X-Real-Ip"); s != "" && stripPort(r.RemoteAddr) == "127.0.0.1" {
		r.RemoteAddr = s
	}
	next(w, r)
}

// stripPort removes the port specification from an address.
func stripPort(s string) string {
	if h, _, err := net.SplitHostPort(s); err == nil {
		s = h
	}
	return s
}

// SetTLS causes downstream handlers to treat this HTTP
// request as having come via TLS. It is necessary because connection
// muxing (which enables a single port to serve both Web and gRPC)
// does not set the http.Request TLS field (since TLS occurs before
// muxing).
func SetTLS(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Header.Get("x-forwarded-proto") == "" {
		r.Header.Set("x-forwarded-proto", "https")
	}
	next(w, r)
}

func RedirectToHTTPS(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	isHTTPS := r.TLS != nil || r.Header.Get("x-forwarded-proto") == "https"
	if !isHTTPS {
		url := *r.URL
		url.Scheme = "https"
		url.Host = r.Host
		http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
		return
	}
	next(w, r)
}

func StrictTransportSecurity(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("strict-transport-security", "max-age=8640000")
	next(w, r)
}

func SecureHeader(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("x-content-type-options", "nosniff")
	w.Header().Set("x-xss-protection", "1; mode=block")
	w.Header().Set("x-frame-options", "DENY")
	next(w, r)
}

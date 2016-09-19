package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
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

// SetTLS causes downstream handlers to treat this HTTP
// request as having come via TLS. It is necessary because connection
// muxing (which enables a single port to serve both Web and gRPC)
// does not set the http.Request TLS field (since TLS occurs before
// muxing).
func SetTLS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-forwarded-proto") == "" {
			r.Header.Set("x-forwarded-proto", "https")
		}
		next.ServeHTTP(w, r)
	})
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

// EnsureHostname ensures that the URL hostname is whatever is in SG_URL.
func EnsureHostname(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantHost := conf.AppURL.Host
		if strings.Split(wantHost, ":")[0] == "localhost" {
			// if localhost, don't enforce redirect, so the site is easier to share with others
			next.ServeHTTP(w, r)
			return
		}

		if r.Host == wantHost || r.Host == "" || r.URL.Path == statusEndpoint {
			next.ServeHTTP(w, r)
			return
		}

		// redirect to desired host
		newURL := *r.URL
		newURL.User = nil
		newURL.Host = wantHost
		newURL.Scheme = conf.AppURL.Scheme
		log.Printf("ensureHostnameHandler: Permanently redirecting from requested host %q to %q.", r.Host, newURL.String())
		http.Redirect(w, r, newURL.String(), http.StatusMovedPermanently)
	})
}

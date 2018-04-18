package cli

import (
	"net/http"
	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

func secureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// headers for security
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "DENY")
		if v, _ := strconv.ParseBool(enableHSTS); v {
			w.Header().Set("Strict-Transport-Security", "max-age=8640000")
		}
		// no cache by default
		w.Header().Set("Cache-Control", "no-cache, max-age=0")

		// CORS
		// If the headerOrigin is the development or production Chrome Extension explictly set the Allow-Control-Allow-Origin
		// to the incoming header URL. Otherwise use the configured CORS origin.
		headerOrigin := r.Header.Get("Origin")
		isExtensionRequest := (headerOrigin == devExtension || headerOrigin == prodExtension) && !disableBrowserExtension
		if corsOrigin := conf.Get().CorsOrigin; corsOrigin != "" || isExtensionRequest {
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			allowOrigin := corsOrigin
			if isExtensionRequest || isAllowedOrigin(headerOrigin, strings.Fields(corsOrigin)) {
				allowOrigin = headerOrigin
			}

			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, X-Sourcegraph-Client, Content-Type")
				w.WriteHeader(http.StatusOK)
				return // do not invoke next handler
			}
		}

		next.ServeHTTP(w, r)
	})
}

package middleware

import (
	"log" //nolint:logging // TODO move all logging to sourcegraph/log
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var httpTrace, _ = strconv.ParseBool(env.Get("HTTP_TRACE", "false", "dump HTTP requests (including body) to stderr"))

// Trace is an HTTP middleware that dumps the HTTP request body (to stderr) if the env var
// `HTTP_TRACE=1`.
func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if httpTrace {
			data, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Println("HTTP_TRACE: unable to print request:", err)
			}
			log.Println("====================================================================== HTTP_TRACE: HTTP request")
			log.Println(string(data))
			log.Println("===============================================================================================")
		}
		next.ServeHTTP(w, r)
	})
}

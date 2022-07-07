package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"

	sglog "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var httpTrace, _ = strconv.ParseBool(env.Get("HTTP_TRACE", "false", "dump HTTP requests (including body) to stderr"))

// Trace is an HTTP middleware that dumps the HTTP request body (to stderr) if the env var
// `HTTP_TRACE=1`.
func Trace(next http.Handler) http.Handler {
	// logger := sglog.Scoped("Trace", "")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// logger.Warn("trace is here")
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

type sentryHeader struct {
	DSN string `json:"dsn"`
}

func SentryTunnel(next http.Handler) http.Handler {
	logger := sglog.Scoped("sentry tunnel", "")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// logger.Warn("here", sglog.String("path", r.URL.Path))
		// Check if the request is aimed at our sentry tunnel.
		if !(r.URL.Path == "/_sentry_tunnel" && r.Method == http.MethodGet) {
			// logger.Warn("skipping")
			next.ServeHTTP(w, r)
			return
		}
		logger.Warn("serving")

		// Read the envelope.
		b, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("couldn't decode")
			return
		}
		defer r.Body.Close()

		logger.Warn("body", sglog.String("data", string(b)))
		// Extract the DSN and ProjectID
		n := bytes.IndexByte(b, '\n')
		if n < 0 {
			return
		}
		h := sentryHeader{}
		err = json.Unmarshal(b[0:n], &h)
		if err != nil {
			return
		}

		// post stuff to sentry
		// TODO

		logger.Warn("posting to sentry", sglog.String("dsn", h.DSN))
		w.WriteHeader(http.StatusOK)

		return
	})
}

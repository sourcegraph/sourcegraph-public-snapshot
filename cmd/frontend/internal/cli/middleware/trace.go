package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

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

// TODO
var sentryHost = "o19358.ingest.sentry.io"

func SentryTunnel(next http.Handler) http.Handler {
	logger := sglog.Scoped("sentry tunnel", "")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// logger.Warn("here", sglog.String("path", r.URL.Path))
		// Check if the request is aimed at our sentry tunnel.
		if !(r.URL.Path == "/_sentry_tunnel" && r.Method == http.MethodPost) {
			next.ServeHTTP(w, r)
			return
		}
		// Read the envelope.
		b, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Warn("failed to read request body")
			return
		}
		defer r.Body.Close()
		// Extract the DSN and ProjectID
		n := bytes.IndexByte(b, '\n')
		if n < 0 {
			return
		}
		h := sentryHeader{}
		err = json.Unmarshal(b[0:n], &h)
		if err != nil {
			logger.Warn("failed to parse request body")
			return
		}
		u, err := url.Parse(h.DSN)
		if err != nil {
			return
		}
		projectID := u.Path
		if !(projectID == "/1334031" || projectID == "/1391511") {
			// not our projects, just discard the request.
			return
		}

		client := http.Client{
			// We want to keep this short, the default client settings are not strict enough.
			Timeout: 3 * time.Second,
		}
		url := fmt.Sprintf("https://%s/api%s/envelope/", sentryHost, projectID)
		go client.Post(url, "text/plain;charset=UTF-8", bytes.NewReader(b))

		w.WriteHeader(http.StatusOK)
		return
	})
}

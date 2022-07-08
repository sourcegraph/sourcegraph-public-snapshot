package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
)

type sentryHeader struct {
	DSN string `json:"dsn"`
}

func SentryTunnel(next http.Handler) http.Handler {
	logger := log.Scoped("sentryTunnel", "A Sentry.io specific HTTP route that allows to forward client-side reports, https://docs.sentry.io/platforms/javascript/troubleshooting/#dealing-with-ad-blockers")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is targeting our sentry tunnel.
		if r.URL.Path != "/_sentry_tunnel" {
			next.ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Read the envelope.
		b, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Warn("failed to read request body", log.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Extract the DSN and ProjectID
		n := bytes.IndexByte(b, '\n')
		if n < 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		h := sentryHeader{}
		err = json.Unmarshal(b[0:n], &h)
		if err != nil {
			logger.Warn("failed to parse request body", log.Error(err))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		u, err := url.Parse(h.DSN)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		projectID := u.Path
		// TODO
		if !(projectID == "/1334031" || projectID == "/1391511") {
			// not our projects, just discard the request.
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		client := http.Client{
			// We want to keep this short, the default client settings are not strict enough.
			Timeout: 3 * time.Second,
		}
		url := fmt.Sprintf("%s/api%s/envelope/", envvar.SentryTunnelEndpoint, projectID)
		go client.Post(url, "text/plain;charset=UTF-8", bytes.NewReader(b))

		w.WriteHeader(http.StatusOK)
		return
	})
}

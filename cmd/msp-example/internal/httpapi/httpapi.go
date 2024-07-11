package httpapi

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

// TODO: Demonstrate a connectrpc+gRPC example here instead.

type Config struct {
	Variable string
}

func Register(h *http.ServeMux, contract runtime.Contract, config Config) error {
	requestCounter, err := getRequestCounter()
	if err != nil {
		return err
	}

	h.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCounter.Add(r.Context(), 1)
		_, _ = w.Write([]byte(fmt.Sprintf("Variable: %s", config.Variable)))
	}))
	// Test endpoint for making CURL requests to arbitrary targets from this
	// service, for testing networking. Requires diagnostic auth.
	h.Handle("/proxy", contract.Diagnostics.DiagnosticsAuthMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.URL.Query().Get("host")
			if host == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("query parameter 'host' is required"))
				return
			}
			hostURL, err := url.Parse(host)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			path := r.URL.Query().Get("path")
			if path == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("query parameter 'path' is required"))
				return
			}

			insecure, _ := strconv.ParseBool(r.URL.Query().Get("insecure"))

			// Copy the request body and build the request
			defer r.Body.Close()
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			proxiedRequest, err := http.NewRequest(r.Method, "/"+path, bytes.NewReader(body))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			// Copy relevant request headers after stripping their prefixes
			for k, vs := range r.Header {
				if strings.HasPrefix(k, "X-Proxy-") {
					for _, v := range vs {
						proxiedRequest.Header.Add(strings.TrimPrefix(k, "X-Proxy-"), v)
					}
				}
			}

			// Send to target
			proxy := httputil.NewSingleHostReverseProxy(hostURL)
			if insecure {
				customTransport := http.DefaultTransport.(*http.Transport).Clone()
				customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
				proxy.Transport = customTransport
			}
			proxy.ServeHTTP(w, proxiedRequest)
		}),
	))

	return nil
}

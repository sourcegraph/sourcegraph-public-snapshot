package example

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Config struct {
	StatelessMode bool
	Variable      string
}

func (c *Config) Load(env *runtime.Env) {
	c.StatelessMode = env.GetBool("STATELESS_MODE", "false", "if true, disable dependencies")
	c.Variable = env.Get("VARIABLE", "13", "variable value")
}

type Service struct{}

var _ runtime.Service[Config] = Service{}

func (s Service) Name() string    { return "example" }
func (s Service) Version() string { return version.Version() }
func (s Service) Initialize(
	ctx context.Context,
	logger log.Logger,
	contract runtime.Contract,
	config Config,
) (background.Routine, error) {
	logger.Info("starting service")

	if !config.StatelessMode {
		if err := initPostgreSQL(ctx, contract); err != nil {
			return nil, errors.Wrap(err, "initPostgreSQL")
		}
		logger.Info("postgresql connection success")

		if err := writeBigQueryEvent(ctx, contract, "service.initialized"); err != nil {
			return nil, errors.Wrap(err, "writeBigQueryEvent")
		}
		logger.Info("bigquery connection success")

		if err := testRedisConnection(ctx, contract); err != nil {
			return nil, errors.Wrap(err, "newRedisConnection")
		}
		logger.Info("redis connection success")
	}

	requestCounter, err := getRequestCounter()
	if err != nil {
		return nil, err
	}

	h := http.NewServeMux()
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
	contract.Diagnostics.RegisterDiagnosticsHandlers(h, serviceState{
		statelessMode: config.StatelessMode,
		contract:      contract,
	})

	return background.CombinedRoutine{
		&httpRoutine{
			log: logger,
			Server: &http.Server{
				Addr: fmt.Sprintf(":%d", contract.Port),
				Handler: otelhttp.NewHandler(h, "http",
					otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
						// If incoming, just include the path since our own host is not
						// very interesting. If outgoing, include the host as well.
						target := r.URL.Path
						if r.RemoteAddr == "" { // no RemoteAddr indicates this is an outgoing request
							target = r.Host + target
						}
						if operation != "" {
							return fmt.Sprintf("%s.%s %s", operation, r.Method, target)
						}
						return fmt.Sprintf("%s %s", r.Method, target)
					})),
			},
		},
	}, nil
}

type httpRoutine struct {
	log log.Logger
	*http.Server
}

func (s *httpRoutine) Name() string { return "http" }

func (s *httpRoutine) Start() {
	if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Error("error stopping server", log.Error(err))
	}
}

func (s *httpRoutine) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "shutdown")
	}

	s.log.Info("server stopped")
	return nil
}

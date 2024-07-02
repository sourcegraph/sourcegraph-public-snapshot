package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sourcegraph/sourcegraph/dev/linearhooks/internal/handlers"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Config struct {
	ConfigFilePath             string
	LinearPersonalAPIKey       string
	LinearWebhookSigningSecret string
}

func (c *Config) Load(env *runtime.Env) {
	c.ConfigFilePath = env.Get("CONFIG_FILE_PATH", "/etc/linear-issue-mover/config.yaml", "path to config file")
	c.LinearPersonalAPIKey = env.Get("LINEAR_PERSONAL_API_KEY", "", "personal API key for Linear")
	c.LinearWebhookSigningSecret = env.Get("LINEAR_WEBHOOK_SIGNING_SECRET", "", "webhook signing secret for Linear")
}

type Service struct{}

var _ runtime.Service[Config] = Service{}

func (s Service) Name() string    { return "linear-issue-mover" }
func (s Service) Version() string { return version.Version() }

func (s Service) Initialize(
	ctx context.Context,
	logger log.Logger,
	contract runtime.ServiceContract,
	config Config,
) (background.Routine, error) {
	logger.Info("starting service")

	b, err := os.ReadFile(config.ConfigFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "read config file %s", config.ConfigFilePath)
	}
	handlers, err := handlers.New(ctx, logger, b, config.LinearPersonalAPIKey, config.LinearWebhookSigningSecret)
	if err != nil {
		return nil, errors.Wrapf(err, "parse config file %s", config.ConfigFilePath)
	}

	h := http.NewServeMux()
	h.Handle("/issue-mover", http.HandlerFunc(handlers.HandleIssueMover))
	// implement MSP contentional healthz endpoint
	h.Handle("/-/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

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

func (s *httpRoutine) Name() string {
	return "http"
}

func (s *httpRoutine) Start() {
	if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Error("error stopping server", log.Error(err))
	}
}

func (s *httpRoutine) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		s.log.Error("error shutting down server", log.Error(err))
	} else {
		s.log.Info("server stopped")
	}
	return nil
}

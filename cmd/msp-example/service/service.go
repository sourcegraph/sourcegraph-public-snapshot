package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/bigquery"
	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/postgresql"
	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/redis"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Service struct{}

var _ runtime.Service[Config] = Service{}

func (s Service) Name() string    { return "msp-example" }
func (s Service) Version() string { return version.Version() }

func (s Service) Initialize(
	ctx context.Context,
	logger log.Logger,
	contract runtime.ServiceContract,
	config Config,
) (background.Routine, error) {
	logger.Info("starting service")

	var (
		bq *bigquery.Client
		rd *redis.Client
		pg *postgresql.Client
	)
	if !config.StatelessMode {
		var err error

		if bq, err = bigquery.NewClient(ctx, contract.Contract); err != nil {
			return nil, errors.Wrap(err, "bigquery")
		}
		if err := bq.Write(ctx, "service.initialized"); err != nil {
			return nil, errors.Wrap(err, "bigquery.Write")
		}
		logger.Info("bigquery connection success")

		if rd, err = redis.NewClient(ctx, contract.Contract); err != nil {
			return nil, errors.Wrap(err, "redis")
		}
		if err := rd.Ping(ctx); err != nil {
			return nil, errors.Wrap(err, "redis.Ping")
		}
		logger.Info("redis connection success")

		if pg, err = postgresql.NewClient(ctx, contract.Contract); err != nil {
			return nil, errors.Wrap(err, "postgresl")
		}
		if err := pg.Ping(ctx); err != nil {
			return nil, errors.Wrap(err, "postgresql.Ping")
		}
		logger.Info("postgresql connection success")
	}

	h := http.NewServeMux()
	if err := httpapi.Register(h, contract.Contract, config.HTTPAPI); err != nil {
		return nil, errors.Wrap(err, "httpapi.Register")
	}

	contract.Diagnostics.RegisterDiagnosticsHandlers(h, serviceState{
		statelessMode: config.StatelessMode,
		bq:            bq,
		rd:            rd,
		pg:            pg,
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

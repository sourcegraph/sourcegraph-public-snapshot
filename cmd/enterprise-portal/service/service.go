package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/codyaccessservice"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/subscriptionsservice"

	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

// Service is the implementation of the Enterprise Portal service.
type Service struct{}

var _ runtime.Service[Config] = (*Service)(nil)

func (Service) Name() string    { return "enterprise-portal" }
func (Service) Version() string { return version.Version() }

func (Service) Initialize(ctx context.Context, logger log.Logger, contract runtime.Contract, config Config) (background.Routine, error) {
	// We use Sourcegraph tracing code, so explicitly configure a trace policy
	policy.SetTracePolicy(policy.TraceAll)

	dotcomDB, err := newDotComDBConn(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "newDotComDBConn")
	}
	defer dotcomDB.Close(context.Background())

	// Validate connection on startup
	if err := dotcomDB.Ping(context.Background()); err != nil {
		return nil, errors.Wrap(err, "dotcomDB.Ping")
	}

	httpServer := http.NewServeMux()

	// Register MSP endpoints
	contract.Diagnostics.RegisterDiagnosticsHandlers(httpServer, serviceState{})

	// Register connect endpoints
	codyaccessservice.RegisterV1(logger, httpServer, dotcomDB)
	subscriptionsservice.RegisterV1(logger, httpServer)

	// Initialize server
	listenAddr := fmt.Sprintf(":%d", contract.Port)
	server := httpserver.NewFromAddr(
		listenAddr,
		&http.Server{
			Addr: listenAddr,
			Handler: h2c.NewHandler(
				otelhttp.NewHandler(
					// Cloud Run only supports HTTP/2 if the service accepts HTTP/2 cleartext (h2c),
					// see https://cloud.google.com/run/docs/configuring/http2
					httpServer,
					"handler",
					// Don't trust incoming spans, start our own.
					otelhttp.WithPublicEndpoint(),
					// Generate custom span names from the request, the default is very vague.
					otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
						// Prefix with 'handle' because outgoing HTTP requests can have similar-looking
						// spans.
						return fmt.Sprintf("handle.%s %s", r.Method, r.URL.Path)
					}),
				),
				&http2.Server{},
			),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: time.Minute,
		},
	)
	return server, nil
}

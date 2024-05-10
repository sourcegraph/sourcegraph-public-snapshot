package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

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

	// Simple test for now, move elsewhere later
	if err := dotcomDB.Ping(context.Background()); err != nil {
		return nil, errors.Wrap(err, "dotcomDB.Ping")
	}

	httpServer := http.NewServeMux()
	contract.Diagnostics.RegisterDiagnosticsHandlers(httpServer, serviceState{})

	listenAddr := fmt.Sprintf(":%d", contract.Port)
	server := httpserver.NewFromAddr(
		listenAddr,
		&http.Server{
			ReadTimeout:  2 * time.Minute,
			WriteTimeout: 2 * time.Minute,
			Handler:      httpServer,
		},
	)
	return server, nil
}

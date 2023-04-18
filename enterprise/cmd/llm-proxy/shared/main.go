package shared

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/trace"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/server"
	proto "github.com/sourcegraph/sourcegraph/internal/llmproxy/v1"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	svc := server.Server{}

	grpcServer := defaults.NewServer(obctx.Logger.Scoped("grpc", "llm-proxy grpc server"))
	proto.RegisterLLMProxyServiceServer(grpcServer, &svc)

	var handler http.Handler
	if config.GRPCWebUI {
		grpcWebServer := mux.NewRouter()
		e := debugserver.NewGRPCWebUIEndpoint("llm-proxy", config.Address)
		grpcWebServer.PathPrefix(e.Path).Handler(e.Handler)

		obctx.Logger.Info("grpc web UI enabled", log.String("path", e.Path))
		handler = grpc.MultiplexHandlers(grpcServer, grpcWebServer)
	} else {
		handler = grpcServer
	}
	handler = trace.HTTPMiddleware(obctx.Logger, handler, conf.DefaultClient())
	handler = instrumentation.HTTPMiddleware("", handler)
	handler = actor.HTTPMiddleware(obctx.Logger, handler)

	server := httpserver.NewFromAddr(config.Address, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      handler,
	})

	// Mark health server as ready and go!
	ready()
	obctx.Logger.Info("service ready", log.String("address", config.Address))

	goroutine.MonitorBackgroundRoutines(ctx, server)

	return nil
}

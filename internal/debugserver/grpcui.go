package debugserver

import (
	"fmt"
	"net/http"

	"github.com/fullstorydev/grpcui/standalone"
	"github.com/sourcegraph/log"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/env"

	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GRPCWebUIEnabled is an additional environment variable that must be true to
// enable the gRPC Web UI.
var GRPCWebUIEnabled = env.MustGetBool("GRPC_WEB_UI_ENABLED", false, "Enable the gRPC Web UI to debug and explore gRPC services")

const gRPCWebUIPath = "/debug/grpcui"

// NewGRPCWebUIEndpoint returns a new Endpoint that serves a gRPC Web UI instance
// that targets the gRPC server specified by target.
//
// serviceName is the name of the gRPC service that will be displayed on the debug page.
func NewGRPCWebUIEndpoint(serviceName, target string) Endpoint {
	logger := log.Scoped("gRPCWebUI")

	var handler http.Handler = &grpcHandler{
		target:   target,
		dialOpts: defaults.DialOptions(logger),
	}

	// gRPC Web UI expects to serve all of its resources
	// under "/". We can't do that, so we need to rewrite
	// the requests to strip the "/debug/grpcui" prefix before
	// passing it to the gRPC Web UI handler.
	handler = http.StripPrefix(gRPCWebUIPath, handler)

	return Endpoint{
		Name: fmt.Sprintf("gRPC Web UI (%s)", serviceName),

		Path: fmt.Sprintf("%s/", gRPCWebUIPath),
		// gRPC Web UI serves multiple assets, so we need to forward _all_ requests under this path
		// to the handler.
		IsPrefix: true,

		Handler: handler,
	}
}

type grpcHandler struct {
	target   string
	dialOpts []grpc.DialOption
}

func (g *grpcHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !GRPCWebUIEnabled {
		http.Error(w, "gRPC Web UI is disabled", http.StatusNotFound)
		return
	}

	ctx := r.Context()

	//lint:ignore SA1019 DialContext will be supported throughout 1.x
	cc, err := grpc.DialContext(ctx, g.target, g.dialOpts...)
	if err != nil {
		err = errors.Wrap(err, "dialing GRPC server")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer cc.Close()

	handler, err := standalone.HandlerViaReflection(ctx, cc, g.target)
	if err != nil {
		err = errors.Wrap(err, "initializing standalone GRPCUI handler")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handler.ServeHTTP(w, r)
}

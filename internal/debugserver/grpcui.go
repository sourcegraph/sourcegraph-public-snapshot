package debugserver

import (
	"fmt"
	"net/http"

	"github.com/fullstorydev/grpcui/standalone"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const gRPCWebUIPath = "/debug/grpcui"

// NewGRPCWebUIEndpoint returns a new Endpoint that serves a gRPC Web UI instance
// that targets the gRPC server specified by target.
//
// serviceName is the name of the gRPC service that will be displayed on the debug page.
func NewGRPCWebUIEndpoint(serviceName, target string) Endpoint {
	var handler http.Handler = &grpcHandler{
		target:   target,
		dialOpts: defaults.DialOptions(),
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
	ctx := r.Context()

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

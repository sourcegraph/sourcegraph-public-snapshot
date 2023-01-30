package api

import (
	"net/http"
	"sync"

	"github.com/fullstorydev/grpcui/standalone"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type GRPCUIHandler struct {
	ctx context.Context

	dialOpts []grpc.DialOption
	target   string

	mu            sync.RWMutex
	grpcUIHandler http.Handler
}

// NewGRPCUIHandler returns a http.Handler that, after the start() callback has been invoked, serves a GRPCUI
// instance that points at the GRPC server that's specified by the given target and dialOpts. The grpc server
// must have reflection enabled.
//
// Until start() has been invoked, the handler will return a 503 Service Unavailable response.
func NewGRPCUIHandler(ctx context.Context, target string, dialOpts ...grpc.DialOption) (handler http.Handler, start func() error) {

	// TODO@ggilmore: Would it be a deeper interface to go ahead and provide the default GRPC client options here?

	h := &GRPCUIHandler{
		ctx:      ctx,
		dialOpts: dialOpts,
		target:   target,
	}

	return h, h.start
}

func (g *GRPCUIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.grpcUIHandler == nil {
		g.serveNotStarted(w, r)
		return
	}

	g.grpcUIHandler.ServeHTTP(w, r)
}

// start dials the GRPC server that's specified by the provided settings, starts the GRPCUI server,
// and swaps the handler implementation to point to the GRPCUI server instead.
func (g *GRPCUIHandler) start() error {
	cc, err := grpc.DialContext(g.ctx, g.target, g.dialOpts...)
	if err != nil {
		return errors.Wrap(err, "dialing GRPC server")
	}

	handler, err := standalone.HandlerViaReflection(g.ctx, cc, g.target)
	if err != nil {
		return errors.Wrap(err, "initializing standalone GRPCUI handler")

	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.grpcUIHandler = handler

	return nil
}

// serveNotStarted is a default http.Handler implementation that is
// used when the grpcUI server hasn't been started yet.
func (g *GRPCUIHandler) serveNotStarted(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("grpcui not started"))
}

var _ http.Handler = &GRPCUIHandler{}

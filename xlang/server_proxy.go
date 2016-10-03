package xlang

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	"github.com/neelance/parallel"
	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx"
)

// serverID identifies a lang/build server by the minimal state
// necessary to reinitialize it. At most one lang/build server per
// serverID will be used; if two clients issue requests that route to
// the same serverID, their requests will be sent to the same
// lang/build server.
type serverID struct {
	contextID
	pathPrefix string // path to subdirectory, if lang/build server should run in subdirectory (otherwise empty)
}

func (id serverID) String() string {
	if id.pathPrefix == "" {
		return fmt.Sprintf("server(%s)", id.contextID)
	}
	return fmt.Sprintf("server(%s prefix=%q)", id.contextID, id.pathPrefix)
}

type serverProxyConn struct {
	proxy *Proxy         // the proxy that opened this conn
	conn  *jsonrpc2.Conn // the LSP JSON-RPC 2.0 connection to the server

	id serverID

	mu     sync.Mutex
	rootFS ctxvfs.FileSystem // the workspace's file system
	last   time.Time         // max(last request sent, last response received), used to disconnect from unused servers

	shutdown chan struct{} // a channel that is closed when the server is shut down by us
}

var (
	serverConnsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "open_lsp_server_connections",
		Help:      "Number of open connections to the LSP servers.",
	})
	serverConnsCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "cumu_lsp_server_connections",
		Help:      "Cumulative number of connections to the LSP servers (total of open + previously closed since process startup).",
	})
)

func init() {
	prometheus.MustRegister(serverConnsGauge)
	prometheus.MustRegister(serverConnsCounter)
}

// ShutDownIdleServers shuts down servers whose last communication
// with the proxy (either a request or a response) was longer than
// maxIdle ago. The Proxy runs ShutDownIdleServers periodically based
// on p.MaxServerIdle.
func (p *Proxy) ShutDownIdleServers(ctx context.Context, maxIdle time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoff := time.Now().Add(-1 * maxIdle)
	par := parallel.NewRun(runtime.GOMAXPROCS(0))
	for s := range p.servers {
		par.Acquire()
		go func(s *serverProxyConn) {
			defer par.Release()
			s.mu.Lock()
			idle := s.last.Before(cutoff)
			s.mu.Unlock()
			if idle {
				shutdownOK := true
				if err := s.shutdownAndExit(ctx); err != nil {
					par.Error(err)
					shutdownOK = false
				}
				if err := s.conn.Close(); err != nil && shutdownOK {
					par.Error(err)
				}
			}
		}(s)
	}
	return par.Wait()
}

// getServerConn returns an existing connection to the specified
// server or creates one if none exists.
func (p *Proxy) getServerConn(ctx context.Context, id serverID) (*serverProxyConn, error) {
	var c *serverProxyConn

	{
		p.mu.Lock()

		// Check for an already established connection.
		for sc := range p.servers {
			if sc.id == id {
				p.mu.Unlock()
				return sc, nil
			}
		}

		// Acquire per-serverID mu so we don't initiate multiple
		// unnecessary connections, and so that nobody else tries to
		// send on this connection before we've initialized it.
		mu, ok := p.serverNewConnMus[id]
		if !ok {
			mu = new(sync.Mutex)
			p.serverNewConnMus[id] = mu
		}
		p.mu.Unlock()

		mu.Lock()
		defer mu.Unlock()
	}

	// No connection found, so we need to open one.
	if c == nil {
		rwc, err := connectToServer(ctx, id.mode)
		if err != nil {
			return nil, err
		}

		// Create connection.
		var connOpt []jsonrpc2.ConnOpt
		if p.Trace {
			connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(os.Stderr, "", 0)))
		}

		c = &serverProxyConn{
			proxy:    p,
			last:     time.Now(),
			shutdown: make(chan struct{}),
		}
		c.conn = jsonrpc2.NewConn(ctx, rwc, jsonrpc2.HandlerWithError(c.handle), connOpt...)
		c.id = id

		// SECURITY NOTE: We assume that the caller to the LSP client
		// proxy has already checked the user's permissions to read
		// this repo, so we don't need to check permissions again
		// here.
		c.rootFS, err = NewRemoteRepoVFS(id.rootPath.CloneURL(), id.rootPath.Rev())
		if err != nil {
			return nil, err
		}

		if err := c.lspInitialize(ctx); err != nil {
			if err2 := rwc.Close(); err2 != nil {
				return nil, fmt.Errorf("cleaning up after failed server proxy initialize: %s (orig error: %s)", err2, err)
			}
			return nil, err
		}

		// Save connection.
		p.mu.Lock()
		if p.servers == nil {
			p.servers = make(map[*serverProxyConn]struct{}, 1)
		}
		p.servers[c] = struct{}{}
		serverConnsGauge.Set(float64(len(p.servers)))
		serverConnsCounter.Inc()
		p.mu.Unlock()
		go func() {
			select {
			case <-c.conn.DisconnectNotify():
			case <-c.shutdown:
			}
			p.mu.Lock()
			delete(p.servers, c)
			delete(p.serverNewConnMus, c.id)
			serverConnsGauge.Set(float64(len(p.servers)))
			p.mu.Unlock()
			serverConnsGauge.Dec()
		}()
	}

	return c, nil
}

func (c *serverProxyConn) lspInitialize(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "LSP server proxy: initialize",
		opentracing.Tags{"mode": c.id.mode, "rootPath": c.id.rootPath.String()},
	)
	defer span.Finish()
	return c.conn.Call(ctx, "initialize", lspx.InitializeParams{
		InitializeParams: lsp.InitializeParams{RootPath: "file:///"},
		OriginalRootPath: c.id.rootPath.String(),
	}, nil, addTraceMeta(ctx))
}

// callServer sends an LSP request to the specified server
// (establishing the connection first if necessary).
func (p *Proxy) callServer(ctx context.Context, id serverID, method string, params, result interface{}) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "LSP server proxy: "+method,
		opentracing.Tags{"mode": id.mode, "rootPath": id.rootPath.String(), "method": method, "params": params},
	)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()

	c, err := p.getServerConn(ctx, id)
	if err != nil {
		return err
	}
	c.updateLastTime()

	return c.conn.Call(ctx, method, params, result, addTraceMeta(ctx))
}

// traceFSRequests is whether to trace the LSP proxy's incoming
// requests for fs/readFile, fs/readDir, fs/stat, fs/lstat, etc. It is
// off by default because there are a lot of these and the traces can
// get quite noisy if it's enabled, but it is configurable because
// it's useful when you are debugging certain perf issues.
var traceFSRequests, _ = strconv.ParseBool(os.Getenv("LSP_PROXY_TRACE_FS_REQUESTS"))

func (c *serverProxyConn) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	c.updateLastTime()
	defer c.updateLastTime()

	// Trace the handling of this request. Only create child spans for
	// significant operations, not when we're just receiving traces or
	// performing simple VFS ops.
	var span opentracing.Span
	if shouldCreateChildSpan := req.Method != "telemetry/event" && (traceFSRequests || !strings.HasPrefix(req.Method, "fs/")); shouldCreateChildSpan {
		op := "LSP server proxy: handle " + req.Method

		// Try to get our parent span context from this JSON-RPC request's metadata.
		if req.Meta != nil {
			var carrier opentracing.TextMapCarrier
			if err := json.Unmarshal(*req.Meta, &carrier); err != nil {
				return nil, err
			}
			if clientCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier); err == nil {
				span = opentracing.StartSpan(op, ext.RPCServerOption(clientCtx))
				ctx = opentracing.ContextWithSpan(ctx, span)
			} else if err != opentracing.ErrSpanContextNotFound {
				return nil, err
			}
		}

		// Otherwise derive the span from our own context.
		if span == nil {
			span, ctx = opentracing.StartSpanFromContext(ctx, op)
		}

		span.SetTag("method", req.Method)
		defer func() {
			if err != nil {
				ext.Error.Set(span, true)
				span.LogEvent(fmt.Sprintf("error: %v", err))
			}
			span.Finish()
		}()
	}

	switch req.Method {
	case "telemetry/event":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var rawSpan basictracer.RawSpan
		if err := json.Unmarshal(*req.Params, &rawSpan); err != nil {
			return nil, err
		}
		// Recording the raw span as-is requires the lower-level impl
		// types.
		if o, ok := opentracing.GlobalTracer().(basictracer.Tracer); ok {
			if r, ok := o.Options().Recorder.(*lightstep.Recorder); ok {
				r.RecordSpan(rawSpan)
			}
		}
		return nil, nil

	case "fs/readFile", "fs/readDir", "fs/stat", "fs/lstat":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var path string
		if err := json.Unmarshal(*req.Params, &path); err != nil {
			return nil, err
		}
		if span != nil {
			span.SetTag("path", path)
		}
		return c.handleFS(ctx, req.Method, path)

	case "textDocument/publishDiagnostics":
		// Forward to all clients.
		//
		// TODO(sqs): some clients will have already received these
		c.proxy.mu.Lock()
		defer c.proxy.mu.Unlock()
		for cc := range c.proxy.clients {
			// TODO(sqs): equality match omits pathPrefix
			if cc.context == c.id.contextID {
				if _, err := cc.handleFromServer(ctx, cc.conn, req); err != nil {
					return nil, err
				}
			}
		}
		return nil, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("server proxy handler: method not found: %q", req.Method)}
}

func (c *serverProxyConn) updateLastTime() {
	c.mu.Lock()
	c.last = time.Now()
	c.mu.Unlock()
}

func (c *serverProxyConn) shutdownAndExit(ctx context.Context) error {
	var errs errorList
	done := make(chan struct{})
	go func() {
		if err := c.conn.Call(ctx, "shutdown", nil, nil); err != nil {
			errs.add(err)
		}
		// Even if "shutdown" failed, still call "exit" to (hopefully)
		// tell the server to REALLY exit.
		if err := c.conn.Notify(ctx, "exit", nil); err != nil {
			errs.add(err)
		}
		close(done)
	}()

	// Respect the ctx deadline so we don't block for too long on an
	// unresponsive server or work.
	select {
	case <-done:
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			errs.add(err)
		}
	}

	// Close file system to free up resources (e.g., if the VFS is
	// backed by a file on disk, this will close the file).
	if fs, ok := c.rootFS.(io.Closer); ok && fs != nil {
		if err := fs.Close(); err != nil {
			errs.add(err)
		}
	}

	close(c.shutdown)
	return errs.error()
}

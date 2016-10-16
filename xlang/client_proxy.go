package xlang

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

func (p *Proxy) newClientProxyConn(ctx context.Context, rwc io.ReadWriteCloser) {
	var connOpt []jsonrpc2.ConnOpt
	if p.Trace {
		connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(os.Stderr, "", 0)))
	}

	c := &clientProxyConn{
		proxy: p,
		last:  time.Now(),
	}
	c.conn = jsonrpc2.NewConn(ctx, rwc, jsonrpc2.HandlerWithError(c.handle), connOpt...)

	p.mu.Lock()
	p.clients[c] = struct{}{}
	clientConnsGauge.Set(float64(len(p.clients)))
	clientConnsCounter.Inc()
	p.mu.Unlock()
	go func() {
		select {
		case <-c.conn.DisconnectNotify():
		}
		p.removeClientConn(c)
	}()
}

var (
	clientConnsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "open_client_proxy_connections",
		Help:      "Number of open connections to the xlang client proxy.",
	})
	clientConnsCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "xlang",
		Name:      "cumu_client_proxy_connections",
		Help:      "Cumulative number of connections to the xlang client proxy (total of open + previously closed since process startup).",
	})
)

func init() {
	prometheus.MustRegister(clientConnsGauge)
	prometheus.MustRegister(clientConnsCounter)
}

func (p *Proxy) removeClientConn(c *clientProxyConn) {
	p.mu.Lock()
	delete(p.clients, c)
	clientConnsGauge.Set(float64(len(p.clients)))
	p.mu.Unlock()
}

// DisconnectIdleClients shuts down clients whose last communication
// with the proxy (either a request or response) was longer than
// maxIdle ago. The Proxy runs DisconnectIdleClients periodically
// based on p.MaxClientIdle.
func (p *Proxy) DisconnectIdleClients(maxIdle time.Duration) error {
	cutoff := time.Now().Add(-1 * maxIdle)
	errs := &errorList{}
	var wg sync.WaitGroup
	p.mu.Lock()
	for c := range p.clients {
		c.mu.Lock()
		idle := c.last.Before(cutoff)
		c.mu.Unlock()
		if idle {
			wg.Add(1)
			go func(c *clientProxyConn) {
				defer wg.Done()
				p.removeClientConn(c)
				if err := c.conn.Close(); err != nil {
					errs.add(err)
				}
			}(c)
		}
	}
	// Only hold lock during fast loop iter, not while waiting to
	// close each idle connection (otherwise we could block p.mu for a
	// long time if closing blocks).
	p.mu.Unlock()

	wg.Wait()
	return errs.error()
}

// contextID identifies a client's session by the minimal information
// necessary to reinitialize it. Two client connections can have
// identical contextInfo, in which case they will share lang/build
// servers. This happens frequently, e.g. in the case when two
// anonymous clients are accessing the same repository at the same
// commit.
type contextID struct {
	rootPath uri.URI // the rootPath in the initialize request (typically the repo clone URL + "?REV")
	mode     string  // the mode (i.e., "go" or "typescript")
}

func (id contextID) String() string {
	return fmt.Sprintf("context(%s mode=%s)", id.rootPath.String(), id.mode)
}

type clientProxyConn struct {
	proxy *Proxy         // the proxy that accepted this conn
	conn  *jsonrpc2.Conn // the LSP JSON-RPC 2.0 connection to the client

	mu       sync.Mutex
	context  contextID
	init     *ClientProxyInitializeParams
	last     time.Time // max(last request received, last response sent), used to evict idle clients
	shutdown bool      // whether this connection has received an LSP "shutdown"
}

// ClientProxyInitializeParams are sent by the client to the proxy in
// the "initialize" request. It has a non-standard field "mode", which
// is the name of the language (using vscode terminology); "go" or
// "typescript", for example.
type ClientProxyInitializeParams struct {
	lsp.InitializeParams
	Mode string `json:"mode"`
}

// LogTrackedErrors, if true, causes errors to be logged if they are
// related to language analysis.
var LogTrackedErrors = true

type trackedError struct {
	RootPath string
	Mode     string
	Method   string
	Params   interface{}
	Error    string
}

// handleFromClient receives requests from the client, modifies them,
// sends them to the appropriate lang/build server(s), modifies the
// responses, and returns them to the client.
//
// It modifies the request to rewrite paths (such as initialize's
// rootPath and textDocument/definition's textDocument.uri fields) to
// point to file system paths, checking out the repo to that file
// system path if necessary.
//
// Certain operations (such as workspace/symbols) must be called on
// all build/lang servers, in which case the results are merged
// transparently to the client.
func (c *clientProxyConn) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	c.updateLastTime()
	defer c.updateLastTime()

	// Try to get our parent span context from the JSON-RPC request
	// from the LSP client.
	opName := "LSP client proxy: " + req.Method
	var span opentracing.Span
	var carrier opentracing.TextMapCarrier
	if req.Meta != nil {
		if err := json.Unmarshal(*req.Meta, &carrier); err != nil {
			return nil, err
		}
	}
	if clientCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier); err == nil {
		span = opentracing.StartSpan(opName, ext.RPCServerOption(clientCtx))
		ctx = opentracing.ContextWithSpan(ctx, span)
	} else if err != opentracing.ErrSpanContextNotFound {
		return nil, err
	} else {
		span, ctx = opentracing.StartSpanFromContext(ctx, opName)
	}
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()

	c.mu.Lock()
	shutdown := c.shutdown
	c.mu.Unlock()
	if shutdown && req.Method != "exit" {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidRequest,
			Message: fmt.Sprintf("invalid LSP request %q received while client proxy is shutting down (only \"exit\" is allowed)", req.Method),
		}
	}

	// ensureInitialized should be used below methods that require the
	// client to have already sent an "initialize" request.
	ensureInitialized := func() error {
		c.mu.Lock()
		initialized := c.init != nil
		c.mu.Unlock()
		if !initialized {
			return &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidRequest,
				Message: fmt.Sprintf("LSP client must send \"initialize\" request before sending %q", req.Method),
			}
		}
		return nil
	}

	switch req.Method {
	case "initialize":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params ClientProxyInitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		rootPathURI, err := uri.Parse(params.RootPath)
		if err != nil {
			return nil, fmt.Errorf("invalid rootPath: %s", err)
		}
		if params.Mode == "" {
			return nil, fmt.Errorf(`client must send a "mode" in the initialize request to specify the language`)
		}
		if rootPathURI.Rev() == "" {
			return nil, fmt.Errorf("invalid empty Git revision in rootPath %q", rootPathURI)
		}

		c.mu.Lock()
		defer c.mu.Unlock()
		if c.init != nil {
			// This would only happen if the client is misbehaving (if
			// it sends 2 "initialize" requests).
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: fmt.Sprintf("client proxy handler is already initialized")}
		}
		c.init = &params
		c.context.rootPath = *rootPathURI
		c.context.mode = c.init.Mode
		return lsp.ServerCapabilities{
			ReferencesProvider: true,
			DefinitionProvider: true,
			HoverProvider:      true,
		}, nil

	case "textDocument/definition", "textDocument/hover", "textDocument/references", "workspace/symbol":
		if err := ensureInitialized(); err != nil {
			return nil, err
		}
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var respObj interface{}
		if err := c.callServer(ctx, req.Method, req.Params, &respObj); err != nil {
			// Machine parseable to assist us finding most common errors
			msg, _ := json.Marshal(trackedError{
				RootPath: c.context.rootPath.String(),
				Mode:     c.context.mode,
				Method:   req.Method,
				Params:   req.Params,
				Error:    err.Error(),
			})
			if LogTrackedErrors {
				log.Printf("tracked error: %s", string(msg))
			}
			return nil, err
		}
		return respObj, nil

	case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose", "textDocument/didSave":
		if err := ensureInitialized(); err != nil {
			return nil, err
		}

		// Specifically forbid these methods so we don't accidentally
		// allow them through. If we did, then any user of a shared
		// workspace could modify the files used for analysis for all
		// users.
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: fmt.Sprintf("client proxy handler: text document modifications not allowed by client (%s)", req.Method)}

	case "shutdown":
		c.mu.Lock()
		c.shutdown = true
		c.mu.Unlock()
		return nil, nil

	case "exit":
		c.mu.Lock()
		c.shutdown = true
		c.mu.Unlock()
		c.proxy.removeClientConn(c)
		if err := c.conn.Close(); err != jsonrpc2.ErrClosed { // ignore if already closed
			return nil, err
		}
		return nil, nil

	default:
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("client proxy handler: method not found: %q", req.Method)}
	}
}

// handleFromServer is called by associated server proxy connections
// when they receive requests that should be propagated to the client.
func (c *clientProxyConn) handleFromServer(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	c.updateLastTime()
	defer c.updateLastTime()

	c.mu.Lock()
	shutdown := c.shutdown
	c.mu.Unlock()
	if shutdown {
		return nil, nil
	}

	switch req.Method {
	case "textDocument/publishDiagnostics":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var paramsObj interface{}
		if err := json.Unmarshal(*req.Params, &paramsObj); err != nil {
			return nil, err
		}

		// Rewrite paths from server->client and send rewritten
		// notification to client.
		var walkErr error
		lspx.WalkURIFields(paramsObj, nil, func(uriStr string) string {
			newURI, err := c.rewritePathFromServer(uriStr)
			if err != nil {
				walkErr = err
			}
			return newURI
		})
		if walkErr != nil {
			return nil, walkErr
		}
		if err := conn.Notify(ctx, req.Method, paramsObj); err != nil {
			if err == jsonrpc2.ErrClosed || strings.Contains(err.Error(), "use of closed network connection") {
				err = nil // suppress worthless "notification handling error" log messages when the client has hung up
			}
			return nil, err
		}
		return nil, nil

	default:
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("client handler for propagating server messages: method not found: %q", req.Method)}
	}
}

// rewritePathFromClient updates URIs in client messages that refer to
// repositories (e.g.,
// "git://github.com/facebook/react.git?master#dir/file.txt" ->
// "file:///dir/file.txt") to point to paths on the virtual file
// system that will contain the original path's contents. It checks
// out the remote repository to the file system if necessary.
func (c *clientProxyConn) rewritePathFromClient(uriStr string) (string, error) {
	uri, err := uri.Parse(uriStr)
	if err != nil {
		return "", err
	}
	if p := path.Clean(uri.FilePath()); strings.HasPrefix(p, "/") || strings.HasPrefix(p, "..") {
		return "", fmt.Errorf("invalid file path in URI %q in LSP proxy client request (must not begin with '/', '..', or contain '.' or '..' components)", uriStr)
	} else if uri.FilePath() != "" && p != uri.FilePath() {
		return "", fmt.Errorf("invalid file path in URI %q (raw file path %q != cleaned file path %q)", uriStr, uri.FilePath(), p)
	}
	if *uri.WithFilePath("") != *c.context.rootPath.WithFilePath("") {
		// SECURITY NOTE: This is a safety check against the user
		// trying to specify one repository in the initialize request
		// and refer to another repository's files in the another
		// request. This is important, because we only perform the
		// access check for the initialize request.
		return "", fmt.Errorf("file path %q in LSP proxy client request must be underneath root path %q", uriStr, &c.context.rootPath)
	}
	return "file:///" + uri.FilePath(), nil
}

// rewritePathFromServer is the reverse of rewritePathFromClient. It
// updates URIs in server messages that refer to the local workspace's
// virtual file system to point back to files in the original
// repository (e.g., "file:///dir/file.txt" ->
// "git://github.com/facebook/react.git?master#dir/file.txt").
func (c *clientProxyConn) rewritePathFromServer(uriStr string) (string, error) {
	uri, err := uri.Parse(uriStr)
	if err != nil {
		return "", err
	}
	if uri.Scheme == "file" {
		return c.context.rootPath.WithFilePath(c.context.rootPath.ResolveFilePath(uri.Path)).String(), nil
	}
	return uriStr, nil
	// Another possibility is a "git://" URI that the build/lang
	// server knew enough to produce on its own (e.g., to refer to
	// git://github.com/golang/go for a Go stdlib definition). No need
	// to rewrite those.
}

// callServer sends the LSP request to the server chosen based on the
// client's context and the file URI specified (e.g., for a ".go"
// file, it will choose a Go lang/build server). It rewrites any file
// URIs to refer to file paths in the virtual workspace, not the
// repository clone URL.
func (c *clientProxyConn) callServer(ctx context.Context, method string, params, result interface{}) error {
	pb, err := json.Marshal(params)
	if err != nil {
		return err
	}
	params = nil
	if err := json.Unmarshal(pb, &params); err != nil {
		return err
	}
	var uris []string
	lspx.WalkURIFields(params, func(uri string) {
		uris = append(uris, uri)
	}, nil)
	if len(uris) != 1 && method != "workspace/symbol" {
		return fmt.Errorf("expected exactly 1 document URI (got %d) in LSP params object %s", len(uris), pb)
	}

	// Now that we know the prefix of the workspace, rewrite the paths
	// in the LSP params object.
	var walkErr error
	lspx.WalkURIFields(params, nil, func(uriStr string) string {
		newURI, err := c.rewritePathFromClient(uriStr)
		if err != nil {
			walkErr = err
		}
		return newURI
	})
	if walkErr != nil {
		return walkErr
	}

	id := serverID{contextID: c.context, pathPrefix: ""} // which kind of lang/build server to communicate with
	if err := c.proxy.callServer(ctx, id, method, params, result); err != nil {
		return err
	}

	// Convert the URIs back.
	result2, err := json.Marshal(result)
	if err != nil {
		return err
	}
	var resultObj interface{}
	if err := json.Unmarshal(result2, &resultObj); err != nil {
		return err
	}
	lspx.WalkURIFields(resultObj, nil, func(uriStr string) string {
		newURI, err := c.rewritePathFromServer(uriStr)
		if err != nil {
			walkErr = err
		}
		return newURI
	})
	if walkErr != nil {
		return walkErr
	}
	result2, err = json.Marshal(resultObj)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(result2, result); err != nil {
		return err
	}
	return nil
}

func (c *clientProxyConn) updateLastTime() {
	c.mu.Lock()
	c.last = time.Now()
	c.mu.Unlock()
}

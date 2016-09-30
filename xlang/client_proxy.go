package xlang

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/neelance/parallel"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"

	"golang.org/x/tools/godoc/vfs"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspx"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

func (p *Proxy) newClientProxyConn(ctx context.Context, rwc io.ReadWriteCloser) {
	var connOpt []jsonrpc2.ConnOpt
	if p.Trace {
		connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(os.Stderr, "", 0)))
	}

	c := &clientProxyConn{
		proxy:            p,
		last:             time.Now(),
		disconnectedByUs: make(chan struct{}),
	}
	c.conn = jsonrpc2.NewConn(ctx, rwc, jsonrpc2.HandlerWithError(c.handle), connOpt...)

	p.mu.Lock()
	if p.clients == nil {
		p.clients = make(map[*clientProxyConn]struct{}, 1)
	}
	p.clients[c] = struct{}{}
	clientConnsGauge.Set(float64(len(p.clients)))
	clientConnsCounter.Inc()
	p.mu.Unlock()
	go func() {
		select {
		case <-c.conn.DisconnectNotify():
		case <-c.disconnectedByUs:
		}
		p.mu.Lock()
		delete(p.clients, c)
		clientConnsGauge.Set(float64(len(p.clients)))
		p.mu.Unlock()
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

// DisconnectIdleClients shuts down clients whose last communication
// with the proxy (either a request or response) was longer than
// maxIdle ago. The Proxy runs DisconnectIdleClients periodically
// based on p.MaxClientIdle.
func (p *Proxy) DisconnectIdleClients(maxIdle time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	cutoff := time.Now().Add(-1 * maxIdle)
	par := parallel.NewRun(runtime.GOMAXPROCS(0))
	for c := range p.clients {
		par.Acquire()
		go func(c *clientProxyConn) {
			defer par.Release()
			c.mu.Lock()
			idle := c.last.Before(cutoff)
			c.mu.Unlock()
			if idle {
				if err := c.close(); err != nil {
					par.Error(err)
				}
			}
		}(c)
	}
	return par.Wait()
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
	rootFS   vfs.FileSystem
	last     time.Time // max(last request received, last response sent), used to evict idle clients
	shutdown bool      // whether this connection has received an LSP "shutdown"

	disconnectedByUs chan struct{} // a channel that is closed when we disconnect the client (c.f. conn.DisconnectNotify() tells us when the client disconnects from us)
}

// ClientProxyInitializeParams are sent by the client to the proxy in
// the "initialize" request. It has a non-standard field "mode", which
// is the name of the language (using vscode terminology); "go" or
// "typescript", for example.
type ClientProxyInitializeParams struct {
	lsp.InitializeParams
	Mode string `json:"mode"`
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

	c.mu.Lock()
	if c.shutdown {
		c.mu.Unlock()
		if req.Method == "exit" {
			// Ignore error returned by close, since we want to exit
			// anyways.
			_ = c.close()
			return nil, nil
		}
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: "client proxy handler is shutting down"}
	}
	if c.init == nil && req.Method != "initialize" {
		c.mu.Unlock()
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: "client proxy handler must be initialized"}
	}
	c.mu.Unlock()

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

		rootFS, err := c.prepareRootFileSystem(params.RootPath)
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		if c.init != nil {
			c.mu.Unlock()
			// This would only happen if the client is misbehaving (if
			// it sends 2 "initialize" requests).
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: fmt.Sprintf("client proxy handler is already initialized")}
		}
		c.rootFS = rootFS
		c.init = &params
		c.context.rootPath = *rootPathURI
		c.context.mode = c.init.Mode
		c.mu.Unlock()

		return lsp.ServerCapabilities{
			ReferencesProvider: true,
			DefinitionProvider: true,
			HoverProvider:      true,
		}, nil

	case "textDocument/definition", "textDocument/hover", "textDocument/references", "workspace/symbol":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var respObj interface{}
		if err := c.callServer(ctx, req.Method, req.Params, &respObj); err != nil {
			// Machine parseable to assist us finding most common errors
			msg, _ := json.Marshal(map[string]interface{}{
				"RootPath": c.context.rootPath.String(),
				"Mode":     c.context.mode,
				"Method":   req.Method,
				"Params":   req.Params,
				"Error":    err.Error(),
			})
			log.Printf("tracked error: %s", string(msg))
			return nil, err
		}
		return respObj, nil

	case "shutdown":
		c.mu.Lock()
		c.shutdown = true
		c.mu.Unlock()
		return nil, nil

	case "exit":
		return nil, c.close()

	case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose", "textDocument/didSave":
		// Specifically forbid these methods so we don't accidentally
		// allow them through. If we did, then any user of a shared
		// workspace could modify the files used for analysis for all
		// users.
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: fmt.Sprintf("client proxy handler: text document modifications not allowed by client (%s)", req.Method)}

	default:
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("client proxy handler: method not found: %q", req.Method)}
	}
}

// handleFromServer is called by associated server proxy connections
// when they receive requests that should be propagated to the client.
func (c *clientProxyConn) handleFromServer(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	c.updateLastTime()
	defer c.updateLastTime()

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
			return nil, err
		}
		return nil, nil

	default:
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("client handler for propagating server messages: method not found: %q", req.Method)}
	}
}

func (c *clientProxyConn) prepareRootFileSystem(rootPath string) (vfs.FileSystem, error) {
	root, err := uri.Parse(rootPath)
	if err != nil {
		return nil, err
	}
	create, ok := VFSCreatorsByScheme[root.Scheme]
	if !ok {
		return nil, fmt.Errorf("no VFS creator for scheme %q", root.Scheme)
	}
	return create(root)
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
		return err
	}

	id := serverID{contextID: c.context, pathPrefix: ""} // which kind of lang/build server to communicate with
	if err := c.proxy.callServer(ctx, id, c.rootFS, method, params, result); err != nil {
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
		return err
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

// close closes this connection.
func (c *clientProxyConn) close() error {
	close(c.disconnectedByUs)
	return c.conn.Close()
}

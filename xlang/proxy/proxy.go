package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/neelance/parallel"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

const (
	// CodeModeNotFound is the JSON-RPC 2.0 error code that indicates that the client
	// requested to initialize a session for a mode (language ID, such as "go") that no
	// servers are registered to handle.
	CodeModeNotFound = -32000

	// CodePlatformNotSupported is the JSON-RPC 2.0 error code that indicates that the extension
	// manifest specified an execution platform that is not supported.
	CodePlatformNotSupported = -32001
)

// IsModeNotFound returns whether err (or err's underlying cause) is a JSON-RPC 2.0
// (*jsonrpc2.Error) with error code CodeModeNotFound.
func IsModeNotFound(err error) bool {
	e, ok := errors.Cause(err).(*jsonrpc2.Error)
	return ok && e != nil && e.Code == CodeModeNotFound
}

// New creates a new LSP proxy.
func New() *Proxy {
	return &Proxy{
		MaxClientIdle:   30 * time.Minute,
		MaxServerIdle:   300 * time.Second,
		MaxServerUnused: 30 * time.Second,

		closed: make(chan struct{}),

		clients: map[*clientProxyConn]struct{}{},
		servers: map[*serverProxyConn]struct{}{},
	}
}

// Proxy proxies LSP JSON-RPC 2.0 connections, sitting between the
// client (typically a user's browser, via our HTTP API) and
// lang/build servers.
type Proxy struct {
	MaxClientIdle   time.Duration // disconnect idle clients after this duration
	MaxServerIdle   time.Duration // shut down idle servers after this duration
	MaxServerUnused time.Duration // shut down unused servers after this duration

	Trace bool // print traces of all requests/responses between proxy and client

	closed chan struct{} // a channel that is closed when (*Proxy).Close is called

	mu      sync.Mutex
	clients map[*clientProxyConn]struct{} // open connections from clients
	servers map[*serverProxyConn]struct{} // open connections to lang/build servers
}

// Serve accepts incoming client connections on the listener l.
//
// The client should send an LSP initialize request immediately after
// connecting.
//
// Serve always returns a non-nil error.
func (p *Proxy) Serve(ctx context.Context, lis net.Listener) error {
	// Run background goroutines to disconnect idle clients and
	// terminate idle servers.
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return // stop when the listener is closed
			case <-time.After(p.MaxClientIdle / 2):
				if err := p.DisconnectIdleClients(p.MaxClientIdle); err != nil {
					log15.Error("LSP proxy: disconnecting idle clients", "error", err)
				}
			}
		}
	}()
	go func() {
		for {
			d := p.MaxServerIdle
			if p.MaxServerUnused < d {
				d = p.MaxServerUnused
			}
			select {
			case <-done:
				return // stop when the listener is closed
			case <-time.After(d / 2):
				ctx, cancel := context.WithTimeout(context.Background(), d)
				idleCutoff := time.Now().Add(-1 * p.MaxServerIdle)
				unusedCutoff := time.Now().Add(-1 * p.MaxServerUnused)
				filter := func(s *serverProxyConn) bool {
					s.mu.Lock()
					last := s.stats.Last
					unused := s.stats.TotalCount == 0
					// If the only request has been for workspace/xreference,
					// expire now. If workspace/xreferences is present it is
					// usually the only request done to a server.
					isShortLived := s.stats.TotalCount == 1 && s.stats.TotalFinishedCount == 1 && s.stats.Counts["workspace/xreferences"] == 1
					s.mu.Unlock()
					return last.Before(idleCutoff) || isShortLived || (unused && last.Before(unusedCutoff))
				}
				if err := p.shutdownServers(ctx, filter); err != nil {
					log15.Error("LSP proxy: shutting down idle servers", "error", err)
				}
				cancel()
			}
		}
	}()

	// Watch for language server conf changes and restart if anything changes
	var lsConfMu sync.Mutex
	lsConf, err := json.Marshal(conf.Get().Langservers)
	if err != nil {
		return err
	}
	conf.Watch(func() {
		newLSConf, err := json.Marshal(conf.Get().Langservers)
		if err != nil {
			log15.Error("Error marshaling new langserver config", "error", err)
			return
		}

		lsConfMu.Lock()
		defer lsConfMu.Unlock()

		if bytes.Equal(lsConf, newLSConf) {
			return
		}
		lsConf = newLSConf
		log15.Info("Shutting down all language servers due to config change")
		p.shutdownServers(ctx, func(*serverProxyConn) bool { return true })
	})

	for {
		nc, err := lis.Accept()
		if err != nil {
			return err
		}
		err = p.newClientProxyConn(ctx, nc)
		if err != nil {
			return err
		}
	}
}

// Close shuts down all build/language servers and closes all client
// and server connections. It does NOT stop any listeners passed to
// Serve; those must be closed prior to calling Close.
//
// TODO(sqs): consider returning from Serve or printing a log message
// if this Close is called but there are still active listeners.
func (p *Proxy) Close(ctx context.Context) error {
	par := parallel.NewRun(runtime.GOMAXPROCS(0))
	p.mu.Lock()
	for c := range p.clients {
		par.Acquire()
		go func(c *clientProxyConn) {
			defer par.Release()
			if err := c.conn.Close(); err != nil {
				par.Error(err)
			}
		}(c)
	}
	for s := range p.servers {
		s.Close()
	}

	// Set to nil so that calls to DisconnectIdleClients and
	// shutdownIdleServers that are blocked on p.mu (which we hold) do
	// not attempt to double-close any client/server conns (thereby
	// causing a panic).
	p.clients = nil
	p.servers = nil

	// Only hold lock during fast loop iter; no need to wait for the
	// shutdowns/disconnects to complete.
	p.mu.Unlock()

	return par.Wait()
}

func (p *Proxy) getSavedDiagnostics(id serverID, documentURI lsp.DocumentURI) []lsp.Diagnostic {
	if !id.share {
		return nil
	}

	var c *serverProxyConn
	p.mu.Lock()
	for cc := range p.servers {
		if cc.id.share && cc.id == id {
			c = cc
			break
		}
	}
	p.mu.Unlock()
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.diagnostics[diagnosticsKey{serverID: c.id, documentURI: documentURI}]
}

// getSavedMessages returns the saved messages for the specified
// server proxy. The slice returned should not be mutated.
func (p *Proxy) getSavedMessages(id serverID) []json.RawMessage /* lsp.{Log,Show}MessageParams */ {
	if !id.share {
		return nil
	}

	var c *serverProxyConn
	p.mu.Lock()
	for cc := range p.servers {
		if cc.id.share && cc.id == id {
			c = cc
			break
		}
	}
	p.mu.Unlock()
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// return a shallow copy as the slice may be mutated by other goroutines
	return c.messages[:]
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/jsonrpc2"
)

var (
	proxyAddr = flag.String("proxyAddress", "127.0.0.1:8080", "proxy server listen address (tcp)")
	lspAddr   = flag.String("lspAddress", "", "language server listen address (tcp)")
)

type cloneProxy struct {
	client *jsonrpc2.Conn // connection to the browser
	server *jsonrpc2.Conn // connection to the language server
	ready  chan struct{}  // barrier to block handling requests until proxy is fully initalized
	id     uuid.UUID
	ctx    context.Context
}

func (p *cloneProxy) start() {
	close(p.ready)
}

type jsonrpc2HandlerFunc func(context.Context, *jsonrpc2.Conn, *jsonrpc2.Request)

func (h jsonrpc2HandlerFunc) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	h(ctx, conn, req)
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *proxyAddr)
	if err != nil {
		err = errors.Wrap(err, "setting up proxy listener failed")
		panic(err)
	}

	defer lis.Close()
	log.Println(fmt.Sprintf("CloneProxy: accepting connections at %s", *proxyAddr))

	for {
		clientNetConn, err := lis.Accept()
		if err != nil {
			log.Println("error when accepting client connection", err.Error())
			continue
		}

		go func(clientNetConn net.Conn) {
			ctx := context.Background()
			lsNetConn, err := dialLanguageServer(ctx, *lspAddr)
			if err != nil {
				log.Println("dialing language server failed", err.Error())
				return
			}

			proxy := &cloneProxy{
				ready: make(chan struct{}),
				ctx:   ctx,
			}

			proxy.client = jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(clientNetConn, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.AsyncHandler(jsonrpc2HandlerFunc(proxy.handleClientRequest)))
			proxy.server = jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(lsNetConn, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.AsyncHandler(jsonrpc2HandlerFunc(proxy.handleServerRequest)))
			proxy.id = uuid.New()

			proxy.start()

			// When one side of the connection disconnects, close the other side.
			select {
			case <-proxy.client.DisconnectNotify():
				proxy.server.Close()
			case <-proxy.server.DisconnectNotify():
				proxy.client.Close()
			}
		}(clientNetConn)
	}
}

func (p *cloneProxy) handleServerRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	<-p.ready
	roundTrip(ctx, p.server, p.client, req)
}

func (p *cloneProxy) handleClientRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	<-p.ready
	roundTrip(ctx, p.client, p.server, req)
}

// roundtrip passes requests from one side of the connection to the other.
func roundTrip(ctx context.Context, from *jsonrpc2.Conn, to jsonrpc2.JSONRPC2, req *jsonrpc2.Request) {
	if req.Notif {
		if err := to.Notify(ctx, req.Method, req.Params); err != nil {
			log.Println("CloneProxy: error when sending notification", err.Error())
		}
		return
	}

	callOpts := []jsonrpc2.CallOption{
		// Proxy the ID used. Otherwise we assign our own ID, breaking
		// calls that depend on controlling the ID such as
		// $/cancelRequest and $/partialResult.
		jsonrpc2.PickID(req.ID),
	}

	var result json.RawMessage
	err := to.Call(ctx, req.Method, req.Params, &result, callOpts...)

	if err != nil {
		var respErr *jsonrpc2.Error
		if e, ok := err.(*jsonrpc2.Error); ok {
			respErr = e
		} else {
			respErr = &jsonrpc2.Error{Message: err.Error()}
		}

		var multiErr error = respErr
		if err = from.ReplyWithError(ctx, req.ID, respErr); err != nil {
			multiErr = multierror.Append(multiErr, errors.Wrap(err, "when sending error reply back"))
		}
		log.Println("CloneProxy: error when calling method", multiErr.Error())
		return
	}

	if err = from.Reply(ctx, req.ID, &result); err != nil {
		log.Println("CloneProxy: error when sending reply back", err.Error())
	}
}

// dialLanguageServer creates a connection to the language server specified at addr.
func dialLanguageServer(ctx context.Context, addr string) (net.Conn, error) {
	if addr == "" {
		return nil, errors.New(fmt.Sprintf("language server not found at addr: %s", addr))
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
}

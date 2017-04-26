package backend

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/sourcegraph/zap"
)

// TODO(john): this file is copypasta from zap, there's much more here than is strictly necessary.

var (
	ZapServerURL = os.ExpandEnv(env.Get("ZAP_SERVER", "ws://zap:8433", "zap server URL (ws:///abspath or ws://host:port)"))
)

func parseListenDialURL(urlStr string) (*url.URL, error) {
	if strings.HasPrefix(urlStr, "unix://") && !strings.HasPrefix(urlStr, "unix:///") {
		// Relative path to socket.
		return &url.URL{Scheme: "unix", Path: strings.TrimPrefix(urlStr, "unix://")}, nil
	}
	return url.Parse(urlStr)
}

func dial(ctx context.Context, urlStr string) (jsonrpc2.ObjectStream, error) {
	if urlStr == "" {
		panic("empty dial URL")
	}
	u, err := parseListenDialURL(urlStr)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "unix":
		conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: u.Path, Net: "unix"})
		if err != nil {
			return nil, err
		}
		return jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), nil

	case "tcp":
		conn, err := net.Dial("tcp", u.Host)
		if err != nil {
			return nil, err
		}
		return jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), nil

	case "ws", "wss", "http", "https":
		if u.Scheme == "http" {
			u.Scheme = "ws"
		}
		if u.Scheme == "https" {
			u.Scheme = "wss"
		}
		dialer := *websocket.DefaultDialer
		dialer.NetDial = func(network, addr string) (net.Conn, error) {
			if u.Host == "" {
				network = "unix"
				addr = u.Path
			}
			return net.Dial(network, addr)
		}

		// 🚨 SECURITY: Pass through the actor by overwriting the X-Actor HTTP 🚨
		// header.
		//
		// DO NOT remove this or allow the user to specify an X-Actor header in any
		// way past this point.
		headers := make(http.Header)
		actor.SetTrustedHeader(ctx, headers)
		conn, _, err := dialer.Dial(u.String(), headers)
		if err != nil {
			return nil, err
		}
		return websocketjsonrpc2.NewObjectStream(conn), nil

	default:
		return nil, fmt.Errorf("bad dial URL %q (%s)", urlStr, `supported URL formats: "unix://PATH", "tcp://HOST:PORT", and http/https/ws/wss`)
	}
}

// NewZapClient returns a Zap jsonrpc client. The returned client is solely
// for the actor in the context.
//
// SECURITY: Do NOT cache or otherwise share the returned client across
// different users (actor in ctx).
func NewZapClient(ctx context.Context) (*zap.Client, error) {
	var connOpt []jsonrpc2.ConnOpt
	stream, err := dial(ctx, ZapServerURL)
	if err != nil {
		return nil, err
	}
	cl := zap.NewClient(ctx, stream, connOpt...)
	if _, err := cl.Initialize(ctx, zap.InitializeParams{
		ID:           os.Getenv("USER"),
		Capabilities: zap.ClientCapabilities{},
		Trace:        "off",
	}); err != nil {
		return nil, err
	}
	if err != nil {
		if _, ok := err.(net.Error); ok {
			return nil, fmt.Errorf("%s\n\nIs the local Zap server running and listening at %s? You must have \"zap server\" running", err, "ws:///Users/rothfels/.sourcegraph/zap")
		}
		return nil, err
	}
	return cl, nil
}

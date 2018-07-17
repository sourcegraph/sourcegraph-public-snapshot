package cxpmain

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	mode     = env.Get("CX_MODE", "", "communication mode (stdio|tcp|websocket)")
	addr     = env.Get("CX_ADDR", "", "TCP listen address (if CX_MODE=tcp or CX_MODE=websocket)")
	trace, _ = strconv.ParseBool(env.Get("CX_TRACE", "", "log all messages"))
)

// Main runs a program that exposes the JSON-RPC2 handler according to the configuration specified
// in environment variables.
func Main(name string, handler func() jsonrpc2.Handler) {
	env.Lock()
	log.SetFlags(0)
	env.HandleHelpFlag()

	tracer.Init(tracer.ServiceName(name))

	// Filter log output by level.
	lvl, err := log15.LvlFromString(env.LogLevel)
	if err == nil {
		log15.Root().SetHandler(log15.LvlFilterHandler(lvl, log15.StderrHandler))
	}

	go debugserver.Start()

	var opts []jsonrpc2.ConnOpt
	if trace {
		opts = append(opts, jsonrpc2.LogMessages(log.New(os.Stderr, "", 0)))
	}

	switch mode {
	case "stdio":
		log15.Info("Starting CXP connection on stdio.")
		<-jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(stdio{}, jsonrpc2.VSCodeObjectCodec{}), handler(), opts...).DisconnectNotify()

	case "tcp":
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log15.Error("Error listening for TCP connection.", "error", err)
			return
		}
		defer lis.Close()
		log15.Info("Listening for CXP connections on TCP.", "addr", addr)
		for {
			conn, err := lis.Accept()
			if err != nil {
				log15.Error("Terminating TCP listener.", "error", err)
				return
			}
			jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), handler(), opts...)
		}

	case "websocket":
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span, ctx := opentracing.StartSpanFromContext(r.Context(), "CXP session")
			defer func() {
				if err != nil {
					ext.Error.Set(span, true)
					span.SetTag("err", err.Error())
				}
				span.Finish()
			}()

			conn, err := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
			if err != nil {
				log15.Error("Error upgrading HTTP request to WebSocket CXP session.", "error", err)
				return
			}
			<-jsonrpc2.NewConn(ctx, websocketjsonrpc2.NewObjectStream(conn), handler(), opts...).DisconnectNotify()
		})

		server := &http.Server{
			Addr: addr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// For Kubernetes liveness and readiness probes.
				if r.URL.Path == "/healthz" {
					w.WriteHeader(200)
					w.Write([]byte("ok"))
					return
				}
				handler.ServeHTTP(w, r)
			}),
		}

		// Shutdown on SIGINT.
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			<-c
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := server.Shutdown(ctx)
			if err != nil {
				log15.Crit("Graceful CXP HTTP server shutdown failed, will exit.", "error", err)
				os.Exit(1)
			}
		}()

		log15.Info("Listening for CXP connections over WebSockets.", "addr", addr)
		err = server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}

	default:
		fmt.Fprintln(os.Stderr, "Invalid CX_MODE.")
		fmt.Fprintln(os.Stderr)
		env.PrintHelp()
		os.Exit(1)
	}
}

type stdio struct{}

func (v stdio) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (v stdio) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdio) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}

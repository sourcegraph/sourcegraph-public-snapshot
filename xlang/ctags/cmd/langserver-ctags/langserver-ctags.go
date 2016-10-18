package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/ctags"
)

var (
	mode     = flag.String("mode", "stdio", "communication mode (stdio|tcp)")
	addr     = flag.String("addr", ":2088", "server listen address (tcp)")
	profbind = flag.String("prof-http", ":6060", "net/http/pprof http bind address")
	logfile  = flag.String("log", "/tmp/langserver-ctags.log", "write log output to this file (and stderr)")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	logger := log.New(os.Stderr, "", 0)

	if *logfile != "" {
		f, err := os.Create(*logfile)
		if err != nil {
			return err
		}
		defer f.Close()
		logger.SetOutput(f)
		log.SetOutput(f)
	}

	if *profbind != "" {
		go debugserver.Start(*profbind)
	}

	if t := os.Getenv("LIGHTSTEP_ACCESS_TOKEN"); t != "" {
		opentracing.InitGlobalTracer(lightstep.NewTracer(lightstep.Options{
			AccessToken: t,
		}))
	}

	newHandler := func() jsonrpc2.Handler {
		return jsonrpc2.HandlerWithError((&ctags.Handler{}).Handle)
	}

	switch *mode {
	case "tcp":
		lis, err := net.Listen("tcp", *addr)
		if err != nil {
			return err
		}
		defer lis.Close()
		log.Println("listening on", *addr)
		// We do not use jsonrpc2.Serve since we want to store state per connection
		for {
			conn, err := lis.Accept()
			if err != nil {
				return err
			}
			jsonrpc2.NewConn(context.Background(), conn, newHandler())
		}

	case "stdio":
		log.Println("reading on stdin, writing on stdout")
		<-jsonrpc2.NewConn(context.Background(), stdrwc{}, newHandler(), jsonrpc2.LogMessages(logger)).DisconnectNotify()
		log.Println("connection closed")
		return nil

	default:
		return fmt.Errorf("invalid mode %q", *mode)
	}
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/lang/golang"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
)

var (
	mode     = flag.String("mode", "stdio", "communication mode (stdio|tcp)")
	addr     = flag.String("addr", ":2088", "server listen address (tcp)")
	profbind = flag.String("prof-http", "", "net/http/pprof http bind address")
	logfile  = flag.String("log", "/tmp/langserver-golang.log", "write log output to this file (and stderr)")
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
	if *logfile != "" {
		f, err := os.Create(*logfile)
		if err != nil {
			return err
		}
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, f))
	}

	if *profbind != "" {
		go debugserver.Start(*profbind)
	}

	h := &jsonrpc2.LoggingHandler{&golang.Handler{}}

	switch *mode {
	case "tcp":
		lis, err := net.Listen("tcp", *addr)
		if err != nil {
			return err
		}
		defer lis.Close()
		log.Println("listening on", *addr)
		return jsonrpc2.Serve(lis, h)

	case "stdio":
		log.Println("reading on stdin, writing on stdout")
		jsonrpc2.NewServerConn(os.Stdin, os.Stdout, h)
		select {}

	default:
		return fmt.Errorf("invalid mode %q", *mode)
	}
}

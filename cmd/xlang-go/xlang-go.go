package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/keegancsmith/tmpfriend"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/gobuildserver"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

func init() {
	// If CACHE_DIR is specified, use that
	cacheDir := env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
	vfsutil.ArchiveCacheDir = filepath.Join(cacheDir, "xlang-archive-cache")
}

var (
	mode     = flag.String("mode", "stdio", "communication mode (stdio|tcp)")
	addr     = flag.String("addr", ":4389", "server listen address (tcp)")
	profbind = flag.String("prof-http", ":6060", "net/http/pprof http bind address")

	openGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "golangserver",
		Subsystem: "build",
		Name:      "open_connections",
		Help:      "Number of open connections to the langserver.",
	})
)

func init() {
	prometheus.MustRegister(openGauge)
}

func main() {
	env.Lock()
	flag.Parse()
	log.SetFlags(0)

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	tracer.Init("xlang-go")

	if *profbind != "" {
		go debugserver.Start(*profbind)
	}

	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()

	gobuildserver.Debug = true

	// If xlang-go crashes, all the archives it has cached are not
	// evicted. Over time this leads to us filling up the disk. This is a
	// simple fix were we do a best-effort purge of the cache.
	// https://github.com/sourcegraph/sourcegraph/issues/6090
	_ = os.RemoveAll(vfsutil.ArchiveCacheDir)

	// PERF: Hide latency of fetching golang/go from the first typecheck
	go gobuildserver.FetchCommonDeps()

	switch *mode {
	case "tcp":
		lis, err := net.Listen("tcp", *addr)
		if err != nil {
			return err
		}
		defer lis.Close()

		log.Println("xlang-go: listening on", *addr)
		for {
			conn, err := lis.Accept()
			if err != nil {
				return err
			}
			openGauge.Inc()
			c := jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.AsyncHandler(gobuildserver.NewHandler()))
			go func() {
				<-c.DisconnectNotify()
				openGauge.Dec()
			}()
		}

	case "stdio":
		log.Println("xlang-go: reading on stdin, writing on stdout")
		<-jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.AsyncHandler(gobuildserver.NewHandler())).DisconnectNotify()
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

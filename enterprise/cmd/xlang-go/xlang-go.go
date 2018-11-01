package main

// We require git to clone dependencies not on github or gitserver.
//docker:install git@edge

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
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/xlang-go/internal/server"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
	"github.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

func init() {
	// If CACHE_DIR is specified, use that
	cacheDir := env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
	vfsutil.ArchiveCacheDir = filepath.Join(cacheDir, "xlang-archive-cache")
}

var (
	mode = flag.String("mode", "tcp", "communication mode (stdio|tcp)")
	addr = flag.String("addr", ":4389", "server listen address (tcp)")

	openGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "golangserver",
		Subsystem: "build",
		Name:      "open_connections",
		Help:      "Number of open connections to the language server.",
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
	tracer.Init()

	go debugserver.Start()

	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()

	server.Debug = true

	// If xlang-go crashes, all the archives it has cached are not
	// evicted. Over time this leads to us filling up the disk. This is a
	// simple fix were we do a best-effort purge of the cache.
	// https://github.com/sourcegraph/sourcegraph/issues/6090
	_ = os.RemoveAll(vfsutil.ArchiveCacheDir)

	// PERF: Hide latency of fetching golang/go from the first typecheck
	go server.FetchCommonDeps()

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
			c := jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), jsonrpc2.AsyncHandler(server.NewHandler()))
			go func() {
				<-c.DisconnectNotify()
				openGauge.Dec()
			}()
		}

	default:
		return fmt.Errorf("invalid mode %q", *mode)
	}
}

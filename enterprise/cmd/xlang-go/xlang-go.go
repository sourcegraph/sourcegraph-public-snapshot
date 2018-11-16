package main

// We require git to clone dependencies not on github or gitserver.
//docker:install git@edge

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/keegancsmith/tmpfriend"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/jsonrpc2"
	wsjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/xlang-go/internal/server"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
	"github.com/sourcegraph/sourcegraph/pkg/vfsutil"
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

	listen := func(addr string) (*net.Listener, error) {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Could not bind to address %s: %v", addr, err)
			return nil, err
		}

		if os.Getenv("TLS_CERT") != "" && os.Getenv("TLS_KEY") != "" {
			cert, err := tls.X509KeyPair([]byte(os.Getenv("TLS_CERT")), []byte(os.Getenv("TLS_KEY")))
			if err != nil {
				return nil, err
			}

			listener = tls.NewListener(listener, &tls.Config{
				Certificates: []tls.Certificate{cert},
			})
		}

		return &listener, nil
	}

	switch *mode {
	case "tcp":
		lis, err := listen(*addr)
		if err != nil {
			return err
		}
		defer (*lis).Close()

		log.Println("xlang-go: listening on", *addr)
		for {
			conn, err := (*lis).Accept()
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

	case "websocket":
		mux := http.NewServeMux()
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

		connectionCount := 0

		mux.HandleFunc("/", func(w http.ResponseWriter, request *http.Request) {
			connection, err := upgrader.Upgrade(w, request, nil)
			if err != nil {
				log.Println("error upgrading HTTP to WebSocket:", err)
				http.Error(w, errors.Wrap(err, "could not upgrade to WebSocket").Error(), http.StatusBadRequest)
				return
			}
			defer connection.Close()
			connectionCount = connectionCount + 1
			connectionID := connectionCount
			log.Printf("langserver-go: received incoming connection #%d\n", connectionID)
			<-jsonrpc2.NewConn(context.Background(), wsjsonrpc2.NewObjectStream(connection), server.NewHandler()).DisconnectNotify()
			log.Printf("langserver-go: connection #%d closed\n", connectionID)
		})

		l, err := listen(*addr)
		if err != nil {
			log.Println(err)
			return err
		}

		server := &http.Server{
			Handler:      mux,
			ReadTimeout:  75 * time.Second,
			WriteTimeout: 60 * time.Second,
		}
		log.Println("langserver-go: listening for WebSocket connections on", *addr)
		err = server.Serve(*l)
		log.Println(errors.Wrap(err, "HTTP server"))
		return err

	default:
		return fmt.Errorf("invalid mode %q", *mode)
	}
}

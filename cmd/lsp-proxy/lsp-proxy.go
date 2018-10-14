package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/keegancsmith/tmpfriend"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
	"github.com/sourcegraph/sourcegraph/xlang/proxy"
	"github.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

func init() {
	// If CACHE_DIR is specified, use that
	cacheDir := env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
	vfsutil.ArchiveCacheDir = filepath.Join(cacheDir, "xlang-archive-cache")
}

var (
	addr  = flag.String("addr", ":4388", "proxy server TCP listen address")
	trace = flag.Bool("trace", false, "print traces of JSON-RPC 2.0 requests/responses")
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
	tracer.Init()

	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()

	proxy.RegisterServers()

	if env.InsecureDev && strings.HasPrefix(*addr, ":") {
		*addr = net.JoinHostPort("127.0.0.1", (*addr)[1:])
	}
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	log15.Info("lsp-proxy: listening", "addr", lis.Addr())
	p := proxy.New()
	p.Trace = *trace

	go debugserver.Start(debugserver.Endpoint{
		Name:    "LSP-Proxy Connections",
		Path:    "/lsp-conns",
		Handler: &proxy.DebugHandler{Proxy: p},
	})

	return p.Serve(context.Background(), lis)
}

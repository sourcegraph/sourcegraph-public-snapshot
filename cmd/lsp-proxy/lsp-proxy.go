package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/keegancsmith/tmpfriend"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/proxy"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
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

	// Enable colors by default
	color.NoColor = env.Get("COLOR", "true", "Whether to output colors") == "false"

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	tracer.Init("lsp-proxy")

	// Filter log output by level.
	if lvl, err := log15.LvlFromString(env.LogLevel); err == nil {
		log15.Root().SetHandler(log15.LvlFilterHandler(lvl, log15.StderrHandler))
	}

	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()

	if err := proxy.RegisterServers(); err != nil {
		return err
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

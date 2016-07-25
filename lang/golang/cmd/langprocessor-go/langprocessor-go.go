package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lputil"
)

var (
	httpAddr = flag.String("http", ":4141", "HTTP address to listen on")
	lspAddr  = flag.String("lsp", ":2088", "LSP server address")
	profbind = flag.String("prof-http", ":6060", "net/http/pprof http bind address")
)

func main() {
	flag.Parse()

	if *profbind != "" {
		go debugserver.Start(*profbind)
	}

	gopath := os.Getenv("GOPATH")
	rootPath := func(repo, commit string) string {
		return filepath.Join(gopath, "src", repo)
	}

	log.Println("Translating HTTP", *httpAddr, "to LSP", *lspAddr)
	http.Handle("/", &lputil.Translator{
		Addr:     *lspAddr,
		RootPath: rootPath,
	})
	http.ListenAndServe(*httpAddr, nil)
}

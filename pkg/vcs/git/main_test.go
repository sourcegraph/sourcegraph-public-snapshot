package git_test

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
)

func TestMain(m *testing.M) {
	flag.Parse()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("listen failed: %s", err)
	}

	srv := &http.Server{Handler: (&server.Server{}).Handler()}
	go srv.Serve(l)

	gitserver.DefaultClient = &gitserver.Client{Addrs: []string{l.Addr().String()}}

	os.Exit(m.Run())
}

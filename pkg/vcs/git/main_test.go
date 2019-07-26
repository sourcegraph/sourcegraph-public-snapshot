package git_test

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"sourcegraph.com/cmd/gitserver/server"
	"sourcegraph.com/pkg/gitserver"
)

func TestMain(m *testing.M) {
	flag.Parse()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("listen failed: %s", err)
	}

	srv := &http.Server{Handler: (&server.Server{}).Handler()}
	go srv.Serve(l)

	gitserver.DefaultClient.Addrs = func(ctx context.Context) []string {
		return []string{l.Addr().String()}
	}

	os.Exit(m.Run())
}

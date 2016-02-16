package vcs_test

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"src.sourcegraph.com/sourcegraph/pkg/gitserver"
)

func TestMain(m *testing.M) {
	flag.Parse()

	gitserver.RegisterHandler()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("listen failed: %s", err)
	}
	go http.Serve(l, nil)

	if err := gitserver.Dial(l.Addr().String()); err != nil {
		log.Fatalf("dial failed: %s", err)
	}

	os.Exit(m.Run())
}

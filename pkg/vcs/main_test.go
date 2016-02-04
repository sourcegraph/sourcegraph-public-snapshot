package vcs

import (
	"flag"
	"os"
	"testing"

	"src.sourcegraph.com/sourcegraph/pkg/gitserver"
)

func TestMain(m *testing.M) {
	flag.Parse()
	go gitserver.ListenAndServe()
	os.Exit(m.Run())
}

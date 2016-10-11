package golang

import (
	"context"
	"io"
	"os"

	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

func init() {
	// Let Go lang servers specified in env vars take precedence over
	// this in-memory server.
	if os.Getenv("LANGSERVER_GO") != "" {
		return
	}
	xlang.ServersByMode["go"] = func() (io.ReadWriteCloser, error) {
		// Run in-process for easy development (no recompiles, etc.).
		a, b := xlang.InMemoryPeerConns()
		jsonrpc2.NewConn(context.Background(), a, NewBuildHandler())
		return b, nil
	}
}

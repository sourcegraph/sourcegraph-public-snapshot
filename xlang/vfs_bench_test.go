package xlang

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/neelance/parallel"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
)

// BenchmarkVFSOverJSONRPC2 measures the performance of sending a lot
// of files over JSON-RPC. The motivation of creating this benchmark
// was to see if the performance was competitive with Git cloning or
// creating/downloading a Git archive and unpacking it to disk.
//
// Run it with:
//
//   go test ./xlang -bench VFS -benchmem -run '^$'
//
// Over the local Unix socket, on my (@sqs) workstation, I get about
// 17 MB/s. TCP is about the same. Changing the parameters (file byte
// size, batch size, number of batches, and parallelism) doesn't
// affect it much.
func BenchmarkVFSOverJSONRPC2(b *testing.B) {
	ctx := context.Background()

	// Start server.
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		b.Fatal("Listen:", err)
	}
	sh := &vfsBenchServerHandler{rootPath: "/"}
	go func() {
		if err := jsonrpc2.Serve(ctx, l, jsonrpc2.HandlerWithError(sh.handle)); err != nil {
			b.Fatal("jsonrpc2.Serve:", err)
		}
	}()

	// Start the test client.
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		b.Fatal("Dial:", err)
	}
	c := jsonrpc2.NewConn(ctx, conn, jsonrpc2.HandlerWithError(vfsBenchClientHandler{}.handle))

	contents := strings.Repeat("x", 1024*30)
	batches := make([][]vfsBenchFile, 10)
	filesPerBatch := 10
	for b := range batches {
		batches[b] = make([]vfsBenchFile, filesPerBatch)
		for f := range batches[b] {
			batches[b][f] = vfsBenchFile{
				Path:     fmt.Sprintf("/batch/%d/file/%d", b, f),
				Contents: contents,
			}
		}
	}

	npar := 4
	b.Logf("sending ~%.1f MB total (%d batches, each with %d files of %d bytes each) with parallelism %d", float64(len(contents)*len(batches)*filesPerBatch)/1024/1024, len(batches), filesPerBatch, len(contents), npar)
	b.SetBytes(int64(len(contents) * len(batches) * filesPerBatch))

	par := parallel.NewRun(npar)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, batch := range batches {
			par.Acquire()
			go func(files []vfsBenchFile) {
				defer par.Release()
				if err := c.Call(ctx, "create", files, nil); err != nil {
					b.Error(err)
				}
			}(batch)
		}
		par.Wait()
	}
}

type vfsBenchFile struct {
	Path     string
	Contents string
}

type vfsBenchClientHandler struct{}

func (vfsBenchClientHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	panic("handler unexpectedly received a request: " + req.Method)
}

type vfsBenchServerHandler struct {
	rootPath string
}

func (h vfsBenchServerHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	switch req.Method {
	case "create":
		var files []vfsBenchFile
		if err := json.Unmarshal(*req.Params, &files); err != nil {
			return nil, err
		}
		return nil, nil

	default:
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("server proxy handler: method not found: %q", req.Method)}
	}
}

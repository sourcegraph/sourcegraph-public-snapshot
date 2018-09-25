package proxy

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// simulateFSLatency simulates latency to test likely performance when
// this is deployed, if the LSP_PROXY_SIMULATED_LATENCY env var is set
// (otherwise it's a no-op). The lsp-proxy and lang/build server pods
// typically have a 3-6ms of effective network latency, which
// (multiplied by many VFS requests) is significant.
func simulateFSLatency() {
	if s := os.Getenv("LSP_PROXY_SIMULATED_LATENCY"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			panic(err)
		}
		time.Sleep(d)
	}
}

// handleFS handles file system-related requests from the build/lang
// server to the server proxy. It provides a VFS to the build/lang
// server.
func (c *serverProxyConn) handleFS(ctx context.Context, method, path string) (result interface{}, err error) {
	simulateFSLatency()

	switch method {
	case "fs/readFile":
		contents, err := ctxvfs.ReadFile(ctx, c.rootFS, path)
		if err != nil {
			return nil, err
		}
		return contents, nil

	case "fs/readDirFiles":
		dir, _ := filepath.Split(path)
		ls, err := c.rootFS.ReadDir(ctx, dir)
		if err != nil {
			return nil, err
		}
		contents := make(map[string][]byte)
		for _, f := range ls {
			if f.IsDir() {
				continue
			}
			p := filepath.Join(dir, f.Name())
			contents[p], err = ctxvfs.ReadFile(ctx, c.rootFS, p)
			if err != nil {
				return nil, err
			}
		}
		return contents, nil

	case "fs/readDir":
		fis, err := c.rootFS.ReadDir(ctx, path)
		if err != nil {
			return nil, err
		}
		fis2 := make([]lspext.FileInfo, len(fis))
		for i, fi := range fis {
			fis2[i] = lspext.FileInfo{Name_: fi.Name(), Size_: fi.Size(), Dir_: fi.Mode().IsDir()}
		}
		return fis2, nil

	case "fs/stat", "fs/lstat":
		var stat func(context.Context, string) (os.FileInfo, error)
		if method == "fs/stat" {
			stat = c.rootFS.Stat
		} else {
			stat = c.rootFS.Lstat
		}
		fi, err := stat(ctx, path)
		if err != nil {
			return nil, err
		}
		return lspext.FileInfo{Name_: fi.Name(), Size_: fi.Size(), Dir_: fi.Mode().IsDir()}, nil

	default:
		panic("unreachable")
	}
}

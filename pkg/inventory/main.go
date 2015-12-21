// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"src.sourcegraph.com/sourcegraph/pkg/inventory"
)

var (
	dir     = flag.String("dir", ".", "directory to inventory")
	timeout = flag.Duration("timeout", time.Second, "maximum allowed time")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(*timeout))
		defer cancel()
	}

	inv, err := inventory.Scan(ctx, walkableFileSystem{vfs.OS(*dir)})
	if err != nil {
		if err == context.Canceled {
			fmt.Fprintln(os.Stderr, "warning: timeout reached, inventory is incomplete")
		} else {
			fmt.Fprintf(os.Stderr, "error: listing inventory of %s: %s\n", *dir, err)
			os.Exit(1)
		}
	}

	data, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

type walkableFileSystem struct{ vfs.FileSystem }

func (walkableFileSystem) Join(path ...string) string { return filepath.Join(path...) }

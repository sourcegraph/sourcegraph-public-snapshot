// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"context"

	"github.com/sourcegraph/sourcegraph/pkg/inventory"
	"github.com/sourcegraph/sourcegraph/pkg/vfsutil"
	"golang.org/x/tools/godoc/vfs"
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
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	inv, err := inventory.Scan(ctx, vfsutil.Walkable(vfs.OS(*dir), filepath.Join))
	if err != nil {
		if err == context.DeadlineExceeded {
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

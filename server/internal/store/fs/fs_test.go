package fs

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"src.sourcegraph.com/sourcegraph/conf"

	"golang.org/x/net/context"
)

// testContext constructs a new context that owns a temp dir. Call
// done() when done using it to remove the temp dir.
func testContext() (ctx context.Context, done func()) {
	tmpDir, err := ioutil.TempDir("", "fs-store")
	if err != nil {
		panic(err)
	}

	ctx = context.Background()
	ctx = WithReposVFS(ctx, filepath.Join(tmpDir, "repos"))
	ctx = conf.WithURL(ctx, &url.URL{Scheme: "http", Host: "example.com"})
	return ctx, func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Fatalf("Warning: failed to remove temp dir %q: %s.", tmpDir, err)
		}
	}
}

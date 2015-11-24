package fs

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/rwvfs"
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

	appdata := filepath.Join(tmpDir, "appdata")
	err = os.MkdirAll(appdata, 0777)
	if err != nil {
		panic(err)
	}

	ctx = context.Background()
	ctx = WithReposVFS(ctx, filepath.Join(tmpDir, "repos"))
	ctx = WithBuildStoreVFS(ctx, rwvfs.Walkable(rwvfs.OS(filepath.Join(tmpDir, "builds"))))
	ctx = WithDBVFS(ctx, rwvfs.Map(map[string]string{}))
	ctx = WithAppStorageVFS(ctx, rwvfs.Walkable(rwvfs.OS(appdata)))
	ctx = conf.WithAppURL(ctx, &url.URL{Scheme: "http", Host: "example.com"})
	return ctx, func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Fatalf("Warning: failed to remove fs store temp dir %q: %s.", tmpDir, err)
		}
	}
}

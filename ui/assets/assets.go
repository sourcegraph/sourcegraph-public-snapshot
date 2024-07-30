package assets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DefaultAssetPath is the default path where assets should be loaded from. It is primarily used when
//
// * assetPath is empty
// * env var SRC_ASSETS_DIR is empty
//
// This is the default value where assets are created when you run //client/web/dist:copy_bundle
const DefaultAssetPath = "/assets-dist"

// assetsPath is absolute path where assets should be loaded from.
// * During init if it's value is empty the value of the environment variable `SRC_ASSETS_DIR` is used
// * If the environment variable is ALSO empty, value of the constant `DefaultAssetPath` is used
var assetsPath = ""

func Init() {
	// If assetsPath is empty try:
	// * Getting the env var value of `SRC_ASSETS_DIR`
	// * Settle for DefaultAssetPath
	if assetsPath == "" {
		path := os.Getenv("SRC_ASSETS_DIR")
		if path == "" {
			path = DefaultAssetPath
		}
		if !strings.HasPrefix(path, "/") {
			panic(fmt.Sprintf("SRC_ASSETS_DIR %q is not an absolute path", path))
		}
		assetsPath = path
	}
	UseAssetsProviderForPath(assetsPath)
}

// AssetsProvider abstracts accessing assets and the web build manifest.
// One implementation must be explicitly set in the main.go using
// this code. See ui/assets/doc.go
type AssetsProvider interface {
	LoadWebBuildManifest() (*WebBuildManifest, error)
	Assets() http.FileSystem
}

// Provider is a global variable that all assets code will
// reference to access them.
//
// By default, it's assigned the FailingAssetsProvider that
// ensure that not configuring this will result in an explicit
// error message about it.
var Provider AssetsProvider = FailingAssetsProvider{}

// FailingAssetsProvider will panic or return an error if called.
// It's meant to be a safeguard against misconfiguration.
type FailingAssetsProvider struct{}

func (p FailingAssetsProvider) LoadWebBuildManifest() (*WebBuildManifest, error) {
	return nil, errors.New("assets are not configured for this binary, please see ui/assets/doc.go")
}

func (p FailingAssetsProvider) Assets() http.FileSystem {
	panic("assets are not configured for this binary, please see ui/assets/doc.go")
}

// UseDevAssetsProvider installs the development variant of the DirProvider
// which expects assets to be generated on the fly by an external web builder process.
func UseDevAssetsProvider() {
	// When we're using the dev asset provider we expect to be in the monorepo, therefore we use a relative path
	UseAssetsProviderForPath("client/web/dist")
}

// UseAssetsProviderForPath sets the global Provider to a DirProvider using the given path
func UseAssetsProviderForPath(path string) {
	Provider = DirProvider{dir: path, assets: http.Dir(path)}
}

// DirProvider is the development variant of the UseDevAssetsProvider
// which expects assets to be generated on the fly by an external web builder process.
type DirProvider struct {
	dir    string
	assets http.FileSystem
}

func (p DirProvider) LoadWebBuildManifest() (*WebBuildManifest, error) {
	return loadWebBuildManifest(p.dir)
}

func (p DirProvider) Assets() http.FileSystem {
	return p.assets
}

var MockLoadWebBuildManifest func() (*WebBuildManifest, error)

// loadWebBuildManifest uses a web builder manifest to extract hashed bundle names to
// serve to the client. In dev mode, we load this file from disk on demand, so it doesn't
// have to exist at compile time, to avoid a build dependency between frontend
// and client.
func loadWebBuildManifest(rootDir string) (m *WebBuildManifest, err error) {
	if MockLoadWebBuildManifest != nil {
		return MockLoadWebBuildManifest()
	}

	manifestContent, err := os.ReadFile(filepath.Join(rootDir, "web.manifest.json"))
	if err != nil {
		return nil, errors.Wrapf(err, "failed loading web build manifest file from disk at %q", rootDir)
	}
	if err := json.Unmarshal(manifestContent, &m); err != nil {
		return nil, errors.Wrap(err, "parsing manifest json")
	}
	return m, nil
}

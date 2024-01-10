package assets

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

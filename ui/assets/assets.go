package assets

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AssetsProvider interface {
	LoadWebpackManifest() (*WebpackManifest, error)
	Assets() http.FileSystem
}

var Provider AssetsProvider = FailingAssetsProvider{}

func UseDevAssetsProvider() {
	Provider = DevProvider{assets: http.Dir("./ui/assets")}
}

type FailingAssetsProvider struct{}

func (p FailingAssetsProvider) LoadWebpackManifest() (*WebpackManifest, error) {
	return nil, errors.New("assets are not configured for this binary, please see ui/assets")
}

func (p FailingAssetsProvider) Assets() http.FileSystem {
	panic("assets are not configured for this binary, please see ui/assets")
}

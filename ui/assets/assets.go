package assets

import "net/http"

type AssetsProvider interface {
	LoadWebpackManifest() (*WebpackManifest, error)
	Assets() http.FileSystem
}

var Provider AssetsProvider

func UseDevAssetsProvider() {
	Provider = DevProvider{assets: http.Dir("./ui/assets")}
}

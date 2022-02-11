//go:build dist
// +build dist

package assets

import (
	"embed"
	"net/http"
)

//go:embed *
var assetsFS embed.FS

var Assets = http.FS(assetsFS)

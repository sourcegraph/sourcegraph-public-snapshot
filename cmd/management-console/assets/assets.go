// +build !dist

package assets

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/httpfs/filter"
)

// Assets contains the bundled web assets
var Assets http.FileSystem

func init() {
	path := "."
	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		path = filepath.Join(projectRoot, "assets")
	}
	Assets = http.Dir(path)

	// Don't include Go files (which would e.g. include the generated asset file itself).
	Assets = filter.Skip(Assets, filter.FilesWithExtensions(".go"))
}

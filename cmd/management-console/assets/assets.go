//go:generate bash -c "cd ../web && npm install && npm run build"
//go:generate vfsgendev -source="github.com/sourcegraph/sourcegraph/cmd/management-console/assets".Assets

package assets

import (
	"net/http"
	"os"
	"path/filepath"
)

// Assets contains the bundled web assets
var Assets http.FileSystem = http.Dir(".")

func init() {
	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		Assets = http.Dir(filepath.Join(projectRoot, "assets"))
	}
}

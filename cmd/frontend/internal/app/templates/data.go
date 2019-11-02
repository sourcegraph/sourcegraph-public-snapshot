// +build !dist

package templates

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/shurcooL/httpfs/filter"
	"golang.org/x/tools/go/packages"
)

func importPathToDir(importPath string) string {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedFiles}, importPath)
	if err != nil || len(pkgs) == 0 || len(pkgs[0].GoFiles) == 0 {
		log.Fatal("Failed to find templates directory: ", err)
	}
	return filepath.Dir(pkgs[0].GoFiles[0])
}

// Data is a virtual filesystem that contains template data used by Sourcegraph app.
var Data = filter.Skip(
	http.Dir(importPathToDir("github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/templates")),
	filter.FilesWithExtensions(".go"),
)

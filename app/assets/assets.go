// +build !dist

package assets

import (
	"go/build"
	"log"
	"net/http"

	"github.com/shurcooL/httpfs/filter"
)

func importPathToDir(importPath string) string {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		log.Fatalln(err)
	}
	return p.Dir
}

// assets is a virtual filesystem that contains static files served by Sourcegraph app.
var Assets = filter.Skip(
	http.Dir(importPathToDir("sourcegraph.com/sourcegraph/sourcegraph/ui/assets")),
	filter.FilesWithExtensions(".go", ".html"),
)

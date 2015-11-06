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

// Assets is a virtual filesystem that contains template data used by the issues app.
var Assets = filter.New(
	http.Dir(importPathToDir("src.sourcegraph.com/sourcegraph/platform/apps/issues/assets")),
	filter.FilesWithExtensions(".go"),
)

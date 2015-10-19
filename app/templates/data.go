// +build !dist

package templates

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

// Data is a virtual filesystem that contains template data used by Sourcegraph app.
var Data = filter.New(
	http.Dir(importPathToDir("src.sourcegraph.com/sourcegraph/app/templates")),
	filter.FilesWithExtensions(".go"),
)

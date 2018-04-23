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
var Data = filter.Skip(
	http.Dir(importPathToDir("github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/templates")),
	filter.FilesWithExtensions(".go"),
)

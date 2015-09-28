// +build !dist

package templates

import (
	"go/build"
	"log"
	"net/http"
	"os"
	pathpkg "path"

	"github.com/shurcooL/httpfs/filter"
)

func defaultBase(path string) string {
	p, err := build.Import(path, "", build.FindOnly)
	if err != nil {
		log.Fatal(err)
	}
	return p.Dir
}

// Data is a virtual filesystem that contains template data used by Sourcegraph app.
var Data = filter.NewIgnore(
	http.Dir(defaultBase("src.sourcegraph.com/sourcegraph/app/templates")),
	func(fi os.FileInfo, _ string) bool {
		return fi.IsDir() == false && pathpkg.Ext(fi.Name()) == ".go"
	},
)

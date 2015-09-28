// +build !dist

package assets

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

// Assets is a virtual filesystem that contains static files served by Sourcegraph app.
var Assets = filter.NewIgnore(
	http.Dir(defaultBase("sourcegraph.com/sourcegraph/sourcegraph/app/assets")),
	func(fi os.FileInfo, _ string) bool {
		return fi.IsDir() == false && (pathpkg.Ext(fi.Name()) == ".go" || pathpkg.Ext(fi.Name()) == ".html")
	},
)

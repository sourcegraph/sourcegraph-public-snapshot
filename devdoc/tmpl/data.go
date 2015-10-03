// +build !dist

package tmpl

import (
	"go/build"
	"log"
	"net/http"
	"path/filepath"
)

func defaultBase(path string) string {
	p, err := build.Import(path, "", build.FindOnly)
	if err != nil {
		log.Fatal(err)
	}
	return p.Dir
}

// Data is a virtual filesystem that contains template data used by the developer documentation site.
var Data = http.Dir(filepath.Join(defaultBase("src.sourcegraph.com/sourcegraph/devdoc/tmpl"), "data"))

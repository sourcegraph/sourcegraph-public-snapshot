// +build !dist

package assets

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

// Data is a virtual filesystem that contains static files served by Sourcegraph developer documentation site.
var Data = http.Dir(filepath.Join(defaultBase("src.sourcegraph.com/sourcegraph/devdoc/assets"), "data"))

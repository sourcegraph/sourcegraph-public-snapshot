// +build dev

package tmpl

import (
	"go/build"
	"log"
	"path/filepath"
)

// rootDir is the directory containing the html/template template files.
var rootDir = filepath.Join(defaultBase("sourcegraph.com/sourcegraph/appdash/traceapp/tmpl"), "data")

func defaultBase(path string) string {
	p, err := build.Import(path, "", build.FindOnly)
	if err != nil {
		log.Fatal(err)
	}
	return p.Dir
}

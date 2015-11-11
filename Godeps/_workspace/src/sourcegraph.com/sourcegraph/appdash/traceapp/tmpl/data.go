// +build dev

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

// Data is a virtual filesystem that contains template data used by Appdash.
var Data = http.Dir(filepath.Join(defaultBase("sourcegraph.com/sourcegraph/appdash/traceapp/tmpl"), "data"))

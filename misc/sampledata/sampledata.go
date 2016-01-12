// +build !dist

package sampledata

import (
	"go/build"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/shurcooL/httpfs/filter"
)

func importPathToDir(importPath string) string {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		log.Fatalln(err)
	}
	return p.Dir
}

// Data is a virtual filesystem that contains sample data that will be
// added to the server's SGPATH.
var Data = filter.New(
	http.Dir(filepath.Join(importPathToDir("src.sourcegraph.com/sourcegraph/misc/sampledata"), "sgpath")),
	func(path string, fi os.FileInfo) bool {
		if fi.IsDir() {
			return false
		}
		if strings.HasSuffix(path, ".test") {
			return true
		}
		if filepath.Base(path) == "COMMIT_EDITMSG" {
			return true
		}
		return false
	},
)

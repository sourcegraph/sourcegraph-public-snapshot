// +build generate

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/bundle"
)

func main() {
	err := vfsgen.Generate(http.Dir(bundle.BaseDir), vfsgen.Options{
		PackageName:  "bundle",
		BuildTags:    "dist distbundle",
		VariableName: "Data",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

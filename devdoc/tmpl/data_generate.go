// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"src.sourcegraph.com/sourcegraph/devdoc/tmpl"
)

func main() {
	err := vfsgen.Generate(tmpl.Data, vfsgen.Options{
		PackageName:  "tmpl",
		BuildTags:    "dist",
		VariableName: "Data",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

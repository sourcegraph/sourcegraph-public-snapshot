// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"src.sourcegraph.com/sourcegraph/devdoc/assets"
)

func main() {
	err := vfsgen.Generate(assets.Data, vfsgen.Options{
		PackageName:  "assets",
		BuildTags:    "dist",
		VariableName: "Data",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

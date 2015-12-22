// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"src.sourcegraph.com/apps/apidocs/assets"
)

func main() {
	err := vfsgen.Generate(assets.Assets, vfsgen.Options{
		PackageName:  "assets",
		BuildTags:    "dist",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

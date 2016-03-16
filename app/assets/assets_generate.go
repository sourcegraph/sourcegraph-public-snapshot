// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
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

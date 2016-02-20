// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"

	"src.sourcegraph.com/apps/updater"
)

func main() {
	err := vfsgen.Generate(updater.Assets, vfsgen.Options{
		PackageName:  "updater",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

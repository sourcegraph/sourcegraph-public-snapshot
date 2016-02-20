// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"

	"src.sourcegraph.com/apps/tracker"
)

func main() {
	err := vfsgen.Generate(tracker.Assets, vfsgen.Options{
		PackageName:  "tracker",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

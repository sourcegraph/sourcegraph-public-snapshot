// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"

	issues "src.sourcegraph.com/apps/tracker"
)

func main() {
	err := vfsgen.Generate(issues.Assets, vfsgen.Options{
		PackageName:  "issues",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

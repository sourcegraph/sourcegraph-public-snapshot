// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"sourcegraph.com/sourcegraph/sourcegraph/app/templates"
)

func main() {
	err := vfsgen.Generate(templates.Data, vfsgen.Options{
		PackageName:  "templates",
		BuildTags:    "dist",
		VariableName: "Data",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

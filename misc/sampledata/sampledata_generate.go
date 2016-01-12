// +build generate

package main

import (
	"log"

	"src.sourcegraph.com/sourcegraph/misc/sampledata"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(sampledata.Data, vfsgen.Options{
		PackageName:  "sampledata",
		BuildTags:    "dist",
		VariableName: "Data",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

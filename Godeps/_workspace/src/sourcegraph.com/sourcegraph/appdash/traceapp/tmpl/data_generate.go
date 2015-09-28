// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"sourcegraph.com/sourcegraph/appdash/traceapp/tmpl"
)

func main() {
	err := vfsgen.Generate(tmpl.Data, vfsgen.Options{
		PackageName:  "tmpl",
		BuildTags:    "!dev",
		VariableName: "Data",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

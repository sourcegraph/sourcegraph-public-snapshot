// +build generate

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	dir := "../../../doc/"
	err := vfsgen.Generate(http.Dir(dir), vfsgen.Options{
		PackageName:  "docsite",
		BuildTags:    "dist",
		VariableName: "content",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

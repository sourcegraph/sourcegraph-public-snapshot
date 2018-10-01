// +build generate

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	dir := "../../../../../ui/assets/"
	err := vfsgen.Generate(http.Dir(dir), vfsgen.Options{
		PackageName:  "assets",
		BuildTags:    "dist",
		VariableName: "DistAssets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

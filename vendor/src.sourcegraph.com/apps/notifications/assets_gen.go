// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"

	"src.sourcegraph.com/apps/notifications"
)

func main() {
	err := vfsgen.Generate(notifications.Assets, vfsgen.Options{
		PackageName:  "notifications",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

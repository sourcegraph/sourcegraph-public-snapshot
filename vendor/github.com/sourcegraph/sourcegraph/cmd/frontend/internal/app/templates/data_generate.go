// +build generate

package main

import (
	"log"
	"time"

	"github.com/shurcooL/vfsgen"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/templates"
)

func main() {
	err := vfsgen.Generate(templates.Data, vfsgen.Options{
		PackageName:  "templates",
		BuildTags:    "dist",
		VariableName: "Data",
		ModTime:      time.Date(2018, 9, 25, 17, 2, 45, 887370446, time.UTC),
	})
	if err != nil {
		log.Fatalln(err)
	}
}

// +build generate

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(http.Dir("data"), vfsgen.Options{
		PackageName:  "data",
		VariableName: "Data",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

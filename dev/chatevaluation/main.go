package main

import (
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/dev/chatevaluation/controller"
)

func main() {
	repo := flag.String("repo", "", "repository root")
	flag.Parse()

	if *repo == "" {
		fmt.Println("Please specify a repository root")
		return
	}

	if err := controller.Run(*repo); err != nil {
		fmt.Println(err.Error())
		return
	}
}

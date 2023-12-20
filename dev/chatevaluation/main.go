package main

import (
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/dev/chatevaluation/controller"
	"github.com/sourcegraph/sourcegraph/dev/chatevaluation/feature"
)

func main() {
	repo := flag.String("repo", "", "repository root")
	verbose := flag.Bool("v", false, "verbose")
	numFilesPerTest := flag.Int("numFilesPerTest", 10, "number of files per test")
	flag.Parse()
	if *repo == "" {
		fmt.Println("Please specify a repository root")
		return
	}
	if *numFilesPerTest < 1 {
		fmt.Println("Need at least one file to test")
		return
	}
	r := controller.LocalRepo(*repo)
	c := controller.Config{
		NumFilesPerTest: *numFilesPerTest,
		Features: []controller.Feature{
			feature.TypeScriptTypeBreak{},
		},
	}
	if *verbose {
		controller.Diagnosef = func(line string, args ...any) {
			fmt.Printf(line+"\n", args...)
		}
	}
	if err := controller.Run(r, c); err != nil {
		fmt.Println(err.Error())
		return
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

var commands = map[string]func(*graph.DependencyGraph) error{
	"lint":  lint,
	"trace": trace,
}

func mainErr() error {
	if len(os.Args) < 2 {
		fmt.Printf("No command supplied.\n")
		return nil
	}

	graph, err := graph.Load()
	if err != nil {
		return err
	}

	if command, ok := commands[os.Args[1]]; ok {
		return command(graph)
	}

	return fmt.Errorf(fmt.Sprintf("unknown subcommand '%s'", os.Args[1]))
}

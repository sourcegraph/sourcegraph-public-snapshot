package main

//docker:install graphviz

import (
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

func main() {
	env.Lock()
	err := cli.Main()
	if err != nil {
		os.Exit(1)
	}
}

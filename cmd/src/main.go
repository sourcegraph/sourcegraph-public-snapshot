// src is the Sourcegraph server and API client program.
package main // import "sourcegraph.com/sourcegraph/sourcegraph/cmd/src"

//docker:install graphviz

import (
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

func main() {
	env.Lock()
	err := cli.Main()
	if err != nil {
		os.Exit(1)
	}
}

// src is the Sourcegraph server and API client program.
package main // import "sourcegraph.com/sourcegraph/sourcegraph/cmd/src"

//docker:install graphviz

import (
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	// External services
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"

	// Misc.
	_ "sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

func main() {
	env.Lock()
	err := cli.Main()
	if err != nil {
		os.Exit(1)
	}
}

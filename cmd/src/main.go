// src is the Sourcegraph server and API client program.
package main // import "sourcegraph.com/sourcegraph/sourcegraph/cmd/src"

import (
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/cli"

	// App
	_ "sourcegraph.com/sourcegraph/sourcegraph/app/cmd"

	// Server
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/backend/cmd"

	// Events
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/events"
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/events/listeners"

	// External services
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"

	// Misc.
	_ "sourcegraph.com/sourcegraph/sourcegraph/xlang"
	_ "sourcegraph.com/sourcegraph/sourcegraph/xlang/golang"
)

func main() {
	err := cli.Main()
	if err != nil {
		os.Exit(1)
	}
}

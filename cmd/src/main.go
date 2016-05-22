// src is the Sourcegraph server and API client program.
package main

import (
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/services/worker"

	// App
	_ "sourcegraph.com/sourcegraph/sourcegraph/app/cmd"

	// Server
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/backend/cmd"

	// Events
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/events"
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/events/listeners"

	// External services
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	_ "sourcegraph.com/sourcegraph/sourcegraph/services/ext/papertrail"

	// Misc.
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil/cli"
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/wellknown"
)

func main() {
	err := cli.Main()
	worker.CloseLogs()
	if err != nil {
		os.Exit(1)
	}
}

// src is the Sourcegraph server and API client program.
package main

import (
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/sgx"
	"sourcegraph.com/sourcegraph/sourcegraph/worker"

	// App
	_ "sourcegraph.com/sourcegraph/sourcegraph/app/cmd"

	// Server
	_ "sourcegraph.com/sourcegraph/sourcegraph/server/cmd"

	// Events
	_ "sourcegraph.com/sourcegraph/sourcegraph/events"
	_ "sourcegraph.com/sourcegraph/sourcegraph/events/listeners"

	// External services
	_ "sourcegraph.com/sourcegraph/sourcegraph/ext/github"
	_ "sourcegraph.com/sourcegraph/sourcegraph/ext/papertrail"

	// Misc.
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/wellknown"
	_ "sourcegraph.com/sourcegraph/sourcegraph/util/traceutil/cli"

	// Platform applications
	_ "sourcegraph.com/sourcegraph/sourcegraph/platform/apps/docs"
	_ "sourcegraph.com/sourcegraph/sourcegraph/platform/apps/godoc"
)

func main() {
	err := sgx.Main()
	worker.CloseLogs()
	if err != nil {
		os.Exit(1)
	}
}

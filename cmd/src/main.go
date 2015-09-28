// src is the Sourcegraph server and API client program.
package main

import (
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/sgx"

	// App
	_ "sourcegraph.com/sourcegraph/sourcegraph/app/cmd"

	// Stores
	_ "sourcegraph.com/sourcegraph/sourcegraph/server/cmd"

	// External services
	_ "sourcegraph.com/sourcegraph/sourcegraph/ext/aws"
	_ "sourcegraph.com/sourcegraph/sourcegraph/ext/github"
	_ "sourcegraph.com/sourcegraph/sourcegraph/ext/papertrail"

	// Misc.
	_ "sourcegraph.com/sourcegraph/sourcegraph/devdoc"
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/wellknown"
	_ "sourcegraph.com/sourcegraph/sourcegraph/util/traceutil/cli"

	// Platform applications
	_ "sourcegraph.com/sourcegraph/issues"
	_ "sourcegraph.com/sourcegraph/sourcegraph/platform/apps/docs"
)

func main() {
	err := sgx.Main()
	sgx.CloseLogs()
	if err != nil {
		os.Exit(1)
	}
}

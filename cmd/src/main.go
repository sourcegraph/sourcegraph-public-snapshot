// src is the Sourcegraph server and API client program.
package main

import (
	"os"

	"src.sourcegraph.com/sourcegraph/sgx"
	"src.sourcegraph.com/sourcegraph/worker"

	// App
	_ "src.sourcegraph.com/sourcegraph/app/cmd"

	// Stores
	_ "src.sourcegraph.com/sourcegraph/server/cmd"

	// Events
	_ "src.sourcegraph.com/sourcegraph/events"
	_ "src.sourcegraph.com/sourcegraph/events/listeners"

	// External services
	_ "src.sourcegraph.com/sourcegraph/ext/github"
	_ "src.sourcegraph.com/sourcegraph/ext/papertrail"

	// Misc.
	_ "src.sourcegraph.com/sourcegraph/devdoc"
	_ "src.sourcegraph.com/sourcegraph/pkg/wellknown"
	_ "src.sourcegraph.com/sourcegraph/util/traceutil/cli"

	// Platform applications
	_ "src.sourcegraph.com/apps/apidocs"
	_ "src.sourcegraph.com/apps/notifications/sgapp"
	_ "src.sourcegraph.com/apps/tracker/sgapp"
	_ "src.sourcegraph.com/sourcegraph/platform/apps/changesets"
	_ "src.sourcegraph.com/sourcegraph/platform/apps/docs"
	_ "src.sourcegraph.com/sourcegraph/platform/apps/godoc"

	// VCS
	_ "sourcegraph.com/sourcegraph/go-vcs/vcs/git"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
	_ "sourcegraph.com/sourcegraph/go-vcs/vcs/hgcmd"
)

func init() {
	// Getting ModTime for directories with many files is slow, so avoid doing it since we don't need results.
	gitcmd.SetModTime = false
}

func main() {
	err := sgx.Main()
	worker.CloseLogs()
	if err != nil {
		os.Exit(1)
	}
}

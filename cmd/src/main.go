// src is the Sourcegraph server and API client program.
package main

import (
	"log"
	"os"
	"os/signal"

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
	_ "src.sourcegraph.com/sourcegraph/ext/aws"
	_ "src.sourcegraph.com/sourcegraph/ext/github"
	_ "src.sourcegraph.com/sourcegraph/ext/papertrail"

	// Misc.
	_ "src.sourcegraph.com/sourcegraph/devdoc"
	_ "src.sourcegraph.com/sourcegraph/pkg/wellknown"
	_ "src.sourcegraph.com/sourcegraph/util/traceutil/cli"

	// Platform applications
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

func init() {
	// Log OS signals that the process receives. Note that when a
	// signal leads to termination of the process (e.g., SIGINT or
	// SIGKILL), the process may terminate before the signal is
	// printed. This is especially the case if GOMAXPROCS=1.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for sig := range c {
			log.Printf("SIGNAL: Process received signal %q", sig)
		}
	}()
}

func main() {
	err := sgx.Main()
	worker.CloseLogs()
	if err != nil {
		os.Exit(1)
	}
}

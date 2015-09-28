// sgtool prepares and releases the Sourcegraph binary and OS packages.
package main

import (
	"log"
	"os"

	"sourcegraph.com/sourcegraph/go-flags"
)

var globalOpts struct {
	Verbose bool `short:"v" long:"verbose" description:"show verbose output"`
}

var CLI = flags.NewParser(&globalOpts, flags.Default)

func init() {
	CLI.LongDescription = "sgtool performs Sourcegraph release and deployment tasks."
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("")

	if _, err := CLI.Parse(); err != nil {
		os.Exit(1)
	}
}

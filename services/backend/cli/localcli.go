// Package cli exposes command-line flags for parent local package.
package cli

import (
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
)

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		_, err := cli.Serve.AddGroup("Local", "Local service", &Flags)
		if err != nil {
			log.Fatal(err)
		}
	})
}

// Flags defines command-line flags for parent local package.
var Flags struct {
	NumAsyncWorkers int `long:"local.num-async-workers" description:"number of async workers to run" default:"4" env:"SRC_NUM_ASYNC_WORKERS"`
}

// Package cli exposes command-line flags for parent local package.
package cli

import (
	"log"
	"time"

	"src.sourcegraph.com/sourcegraph/sgx/cli"
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
	CommitLogCachePeriod time.Duration `long:"local.clcache" description:"how often to refresh the commit-log cache in seconds; if 0, then no cache is used"`
	CommitLogCacheSize   int32         `long:"local.clcachesize" description:"number of commits to cache on refresh" default:"500"`
}

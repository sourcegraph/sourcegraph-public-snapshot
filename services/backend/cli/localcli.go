// Package cli exposes command-line flags for parent local package.
package cli

import (
	"log"
	"time"

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
	CommitLogCachePeriod time.Duration `long:"local.clcache" description:"how often to refresh the commit-log cache in seconds; if 0, then no cache is used"`
	CommitLogCacheSize   int32         `long:"local.clcachesize" description:"number of commits to cache on refresh" default:"500"`

	// DisableRepoInventory disables repo inventorying, which is slow
	// for very large repos because it walks the entire tree for a
	// commit.
	DisableRepoInventory bool `long:"local.disable-repo-inventory" description:"disable repo inventorying (walks all files to determine langs used, etc.; slow for very large repos)"`

	NumAsyncWorkers int `long:"local.num-async-workers" description:"number of async workers to run" default:"4" env:"SRC_NUM_ASYNC_WORKERS"`
}

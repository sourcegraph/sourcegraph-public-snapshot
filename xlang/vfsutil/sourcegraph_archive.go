package vfsutil

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// NewSourcegraphRepoVFS downloads a zip archive from Sourcegraph's raw API and
// returns a new VFS backed by that zip archive.
func NewSourcegraphRepoVFS(repo, rev string) (*ArchiveFS, error) {
	// TODO(chris) switch to the non-internal raw API once authorization is implemented.
	url := api.InternalClient.URL + "/.internal/" + repo + "@" + rev + "/-/raw/"
	cacheKey := repo + "@" + rev
	rootDirInZip := "/"
	return NewZipVFS(url, cacheKey, rootDirInZip, sgFetch.Inc, sgFetchFailed.Inc)
}

var sgFetch = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "sourcegraph_fetch_total",
	Help:      "Total number of fetches by SourcegraphRepoVFS of zip archives from the Sourcegraph raw API.",
})
var sgFetchFailed = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "sourcegraph_fetch_failed_total",
	Help:      "Total number of fetches by SourcegraphRepoVFS of zip archives from the Sourcegraph raw API that failed.",
})

func init() {
	prometheus.MustRegister(sgFetch)
	prometheus.MustRegister(sgFetchFailed)
}

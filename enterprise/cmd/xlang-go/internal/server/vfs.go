package server

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
	"github.com/sourcegraph/sourcegraph/pkg/vfsutil"
)

// RemoteFS fetches a zip archive from gitserver and returns a virtual file
// system interface for accessing the files in the specified repo at the given
// commit.
//
// - If the originalRootUri is not git://, it is assumed to be a zip URL, so the zip archive is fetched from there.
// - If zipURL is specified in the initializationOptions, the zip archive will be fetched from there.
// - Otherwise, the zip is fetched from gitserver.
//
// SECURITY NOTE: This DOES NOT check that the user or context has permissions
// to read the repo. We assume permission checks happen before a request reaches
// a build server.
var RemoteFS = func(ctx context.Context, initializeParams lspext.InitializeParams) (ctxvfs.FileSystem, error) {
	zipURL := func() string {
		initializationOptions, ok := initializeParams.InitializationOptions.(map[string]interface{})
		if !ok {
			return ""
		}
		url, _ := initializationOptions["zipURL"].(string)
		return url
	}()
	if zipURL != "" {
		return vfsutil.NewZipVFS(zipURL, zipFetch.Inc, zipFetchFailed.Inc, true)
	}

	gitURL, err := gituri.Parse(string(initializeParams.OriginalRootURI))
	if err != nil {
		return nil, errors.Wrap(err, "could not parse workspace URI for remotefs")
	}
	if gitURL.Rev() == "" {
		return nil, errors.Errorf("rev is required in uri: %s", initializeParams.OriginalRootURI)
	}
	archiveFS := vfsutil.NewGitServer(gitURL.Repo(), api.CommitID(gitURL.Rev()))
	archiveFS.EvictOnClose = true
	return archiveFS, nil
}

var zipFetch = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "zip_fetch_total",
	Help:      "Total number of times a zip archive was fetched for the currently-viewed repo.",
})
var zipFetchFailed = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "xlang",
	Subsystem: "vfs",
	Name:      "zip_fetch_failed_total",
	Help:      "Total number of times fetching a zip archive for the currently-viewed repo failed.",
})

func init() {
	prometheus.MustRegister(zipFetch)
	prometheus.MustRegister(zipFetchFailed)
}

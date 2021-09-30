/*
Package images describes the publishing scheme for Sourcegraph images.

It is published as a standalone module to enable tooling in other repositories to more
easily use these definitions.
*/
package images

import (
	"fmt"
)

const (
	// SourcegraphDockerDevRegistry is a private registry for dev images, and requires authentication to pull from.
	SourcegraphDockerDevRegistry = "us.gcr.io/sourcegraph-dev"
	// SourcegraphDockerPublishRegistry is a public registry for final images, and does not require authentication to pull from.
	SourcegraphDockerPublishRegistry = "index.docker.io/sourcegraph"
)

// DevRegistryImage returns the name of the image for the given app and tag on the
// private dev registry.
func DevRegistryImage(app, tag string) string {
	root := fmt.Sprintf("%s/%s", SourcegraphDockerDevRegistry, app)
	return maybeTaggedImage(root, tag)
}

// PublishedRegistryImage returns the name of the image for the given app and tag on the
// publish registry.
func PublishedRegistryImage(app, tag string) string {
	root := fmt.Sprintf("%s/%s", SourcegraphDockerPublishRegistry, app)
	return maybeTaggedImage(root, tag)
}

func maybeTaggedImage(rootImage, tag string) string {
	if tag != "" {
		return fmt.Sprintf("%s:%s", rootImage, tag)
	}
	return rootImage
}

// SourcegraphDockerImages is a list of all images published by Sourcegraph.
//
// In general:
//
// - dev images (candidates - see `candidateImageTag`) are published to `SourcegraphDockerDevRegistry`
// - final images (releases, `insiders`) are published to `SourcegraphDockerPublishRegistry`
// - app must be a legal Docker image name (e.g. no `/`)
//
// The `addDockerImages` pipeline step determines what images are built and published.
var SourcegraphDockerImages = []string{
	// Slow images first for faster CI
	"server",
	"frontend",
	"grafana",
	"prometheus",

	"github-proxy",
	"gitserver",
	"query-runner",
	"repo-updater",
	"worker",
	"searcher",
	"symbols",
	"precise-code-intel-worker",

	// Images under docker-images/
	"alpine-3.12",
	"cadvisor",
	"indexed-searcher",
	"redis-cache",
	"redis_exporter",
	"redis-store",
	"search-indexer",
	"syntax-highlighter",
	"jaeger-agent",
	"jaeger-all-in-one",
	"codeintel-db",
	"codeinsights-db",
	"minio",
	"postgres-12.6-alpine",
	"postgres_exporter",
}

// CandidateImageTag provides the tag for a candidate image built for this Buildkite run.
//
// Note that the availability of this image depends on whether a candidate gets built,
// as determined in `addDockerImages()`.
func CandidateImageTag(commit, buildNumber string) string {
	return fmt.Sprintf("%s_%s_candidate", commit, buildNumber)
}

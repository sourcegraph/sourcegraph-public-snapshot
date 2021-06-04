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
	// SourcegraphDockerPublishRegistry is a public registry for final iamges, and does not require authentication to pull from.
	SourcegraphDockerPublishRegistry = "index.docker.io/sourcegraph"
)

// SourcegraphDockerImages is a list of all images published by Sourcegraph.
//
// In general:
//
// - dev images (candidates - see `candidateImageTag`) are published to `SourcegraphDockerDevRegistry`
// - final images (releases, `insiders`) are published to `SourcegraphDockerPublishRegistry`
//
// The `addDockerImages` pipeline step determines what images are built and published.
var SourcegraphDockerImages = []string{
	// Slow images first for faster CI
	"server",
	"frontend",
	"grafana",
	"prometheus",
	"ignite-ubuntu",

	"github-proxy",
	"gitserver",
	"query-runner",
	"repo-updater",
	"worker",
	"searcher",
	"symbols",
	"precise-code-intel-worker",
	"executor-queue",
	"executor",

	// Images under docker-images/
	"cadvisor",
	"indexed-searcher",
	"postgres-11.4",
	"postgres-12.6",
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
	"postgres-12.6",
	"postgres_exporter",
}

// CandidateImageTag provides the tag for a candidate image built for this Buildkite run.
//
// Note that the availability of this image depends on whether a candidate gets built,
// as determined in `addDockerImages()`.
func CandidateImageTag(commit, buildNumber string) string {
	return fmt.Sprintf("%s_%s_candidate", commit, buildNumber)
}

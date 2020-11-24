package ci

import (
	"fmt"
	"os"
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
	"searcher",
	"symbols",
	"precise-code-intel-worker",
	"executor-queue",
	"executor",

	// Images under docker-images/
	"cadvisor",
	"indexed-searcher",
	"postgres-11.4",
	"redis-cache",
	"redis-store",
	"search-indexer",
	"syntax-highlighter",
	"jaeger-agent",
	"jaeger-all-in-one",
	"codeintel-db",
	"minio",
}

// candidateImageTag provides the tag for a candidate image built for this Buildkite run.
//
// Note that the availability of this image depends on whether a candidate gets built,
// as determined in `addDockerImages()`.
func candidateImageTag(c Config) string {
	buildNumber := os.Getenv("BUILDKITE_BUILD_NUMBER")
	return fmt.Sprintf("%s_%s_candidate", c.commit, buildNumber)
}

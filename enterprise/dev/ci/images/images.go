/*
Package images describes the publishing scheme for Sourcegraph images.

It is published as a standalone module to enable tooling in other repositories to more
easily use these definitions.
*/
package images

import (
	"fmt"
	"time"
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

// SourcegraphDockerImages denotes all Docker images that are published by Sourcegraph.
//
// In general:
//
// - dev images (candidates - see `candidateImageTag`) are published to `SourcegraphDockerDevRegistry`
// - final images (releases, `insiders`) are published to `SourcegraphDockerPublishRegistry`
// - app must be a legal Docker image name (e.g. no `/`)
//
// The `addDockerImages` pipeline step determines what images are built and published.
var SourcegraphDockerImages = append(DeploySourcegraphDockerImages,
	"server", "sg")

// DeploySourcegraphDockerImages denotes all Docker images that are included in a typical
// deploy-sourcegraph installation.
//
// Used to cross check images in the deploy-sourcegraph repo. If you are adding or removing an image to https://github.com/sourcegraph/deploy-sourcegraph
// it must also be added to this list.
var DeploySourcegraphDockerImages = []string{
	"alpine-3.14",
	"cadvisor",
	"codeinsights-db",
	"codeintel-db",
	"frontend",
	"github-proxy",
	"gitserver",
	"grafana",
	"indexed-searcher",
	"jaeger-agent",
	"jaeger-all-in-one",
	"minio",
	"postgres-12-alpine",
	"postgres_exporter",
	"precise-code-intel-worker",
	"prometheus",
	"prometheus-gcp",
	"redis-cache",
	"redis-store",
	"redis_exporter",
	"repo-updater",
	"search-indexer",
	"searcher",
	"symbols",
	"syntax-highlighter",
	"worker",
	"migrator",
	"executor",
	"executor-vm",
	"opentelemetry-collector",
}

// CandidateImageTag provides the tag for a candidate image built for this Buildkite run.
//
// Note that the availability of this image depends on whether a candidate gets built,
// as determined in `addDockerImages()`.
func CandidateImageTag(commit string, buildNumber int) string {
	return fmt.Sprintf("%s_%d_candidate", commit, buildNumber)
}

// MainBranchImageTag provides the tag for all commits built on main, which are used for
// continuous deployment.
func MainBranchImageTag(now time.Time, commit string, buildNumber int) string {
	return fmt.Sprintf("%05d_%10s_%.12s", buildNumber, now.Format("2006-01-02"), commit)
}

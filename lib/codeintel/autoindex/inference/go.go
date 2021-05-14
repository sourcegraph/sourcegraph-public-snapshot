package inference

import (
	"path/filepath"
	"regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func GoPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern("go.mod"),
		segmentPattern("vendor"),
	}
}

func CanIndexGoRepo(gitclient GitClient, paths []string) bool {
	for _, path := range paths {
		if canIndexGoPath(path) {
			return true
		}
	}

	return false
}

const lsifGoImage = "sourcegraph/lsif-go:latest"

func InferGoIndexJobs(gitclient GitClient, paths []string) (indexes []config.IndexJob) {
	for _, path := range paths {
		if !canIndexGoPath(path) {
			continue
		}

		root := dirWithoutDot(path)

		dockerSteps := []config.DockerStep{
			{
				Root:     root,
				Image:    lsifGoImage,
				Commands: []string{"go mod download"},
			},
		}

		indexes = append(indexes, config.IndexJob{
			Steps:       dockerSteps,
			Root:        root,
			Indexer:     lsifGoImage,
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		})
	}

	return indexes
}

var goSegmentBlockList = append([]string{
	"vendor",
}, segmentBlockList...)

func canIndexGoPath(path string) bool {
	// TODO(efritz) - support glide, dep, other historic package managers
	// TODO(efritz) - support projects without go.mod but a vendor dir and go files
	return filepath.Base(path) == "go.mod" && containsNoSegments(path, goSegmentBlockList...)
}

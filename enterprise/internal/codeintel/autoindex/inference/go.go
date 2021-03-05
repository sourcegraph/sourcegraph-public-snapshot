package inference

import (
	"path/filepath"
	"regexp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/config"
)

const lsifGoImage = "sourcegraph/lsif-go:latest"

type lsifGoJobRecognizer struct{}

var _ IndexJobRecognizer = lsifGoJobRecognizer{}

func (r lsifGoJobRecognizer) CanIndex(paths []string, gitserver GitserverClientWrapper) bool {
	for _, path := range paths {
		if r.canIndexPath(path) {
			return true
		}
	}

	return false
}

func (r lsifGoJobRecognizer) InferIndexJobs(paths []string, gitserver GitserverClientWrapper) (indexes []config.IndexJob) {
	for _, path := range paths {
		if !r.canIndexPath(path) {
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

func (lsifGoJobRecognizer) Patterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern("go.mod"),
		segmentPattern("vendor"),
	}
}

func (r lsifGoJobRecognizer) canIndexPath(path string) bool {
	// TODO(efritz) - support glide, dep, other historic package managers
	// TODO(efritz) - support projects without go.mod but a vendor dir and go files
	return filepath.Base(path) == "go.mod" && containsNoSegments(path, goSegmentBlockList...)
}

var goSegmentBlockList = append([]string{
	"vendor",
}, segmentBlockList...)

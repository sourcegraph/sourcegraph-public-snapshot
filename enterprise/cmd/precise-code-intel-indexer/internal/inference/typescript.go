package inference

import (
	"path/filepath"
	"regexp"
)

type lsifTscJobRecognizer struct{}

var _ IndexJobRecognizer = lsifTscJobRecognizer{}

func (lsifTscJobRecognizer) CanIndex(paths []string) bool {
	for _, path := range paths {
		if filepath.Base(path) == "tsconfig.json" && !containsSegment(path, "node_modules") {
			return true
		}

		// TODO(efritz) - check for javascript files
	}

	return false
}

func (lsifTscJobRecognizer) InferIndexJobs(paths []string) (indexes []IndexJob) {
	for _, path := range paths {
		if filepath.Base(path) == "tsconfig.json" && !containsSegment(path, "node_modules") {
			root := filepath.Dir(path)
			if root == "." {
				root = ""
			}

			indexes = append(indexes, IndexJob{
				DockerSteps: nil, // TODO(efritz) - yarn or npm
				Root:        root,
				Indexer:     "sourcegraph/lsif-node:latest",
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			})
		}
	}

	return indexes
}

func (lsifTscJobRecognizer) Patterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern("tsconfig.json"),
	}
}

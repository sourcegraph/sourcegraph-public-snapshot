package inference

import (
	"path/filepath"
	"regexp"
)

type lsifGoJobRecognizer struct{}

var _ IndexJobRecognizer = lsifGoJobRecognizer{}

func (lsifGoJobRecognizer) CanIndex(paths []string) bool {
	for _, path := range paths {
		if filepath.Base(path) == "go.mod" {
			return true
		}
	}

	return false
}

func (lsifGoJobRecognizer) InferIndexJobs(paths []string) (indexes []IndexJob) {
	for _, path := range paths {
		if filepath.Base(path) == "go.mod" {
			root := filepath.Dir(path)
			if root == "." {
				root = ""
			}

			indexes = append(indexes, IndexJob{
				DockerSteps: nil,
				Root:        root,
				Indexer:     "sourcegraph/lsif-go:latest",
				IndexerArgs: []string{"lsif-go", "--no-animation"},
				Outfile:     "",
			})
		}
	}

	return indexes
}

func (lsifGoJobRecognizer) Patterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern("go.mod"),
	}
}

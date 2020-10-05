package inference

import "path/filepath"

type goRecognizer struct{}

func (goRecognizer) CanIndex(paths []string) bool {
	for _, path := range paths {
		if filepath.Base(path) == "go.mod" {
			return true
		}
	}

	return false
}

func (goRecognizer) InferIndexJobs(paths []string) (indexes []IndexJob) {
	for _, path := range paths {
		if filepath.Base(path) == "go.mod" {
			root := filepath.Dir(path)
			if root == "." {
				root = ""
			}

			indexes = append(indexes, IndexJob{
				DockerSteps: []DockerStep{},
				Root:        root,
				Indexer:     "sourcegraph/lsif-go:latest",
				IndexerArgs: []string{"lsif-go", "--no-animation"},
				Outfile:     "",
			})
		}
	}

	return indexes
}

func (goRecognizer) Patterns() []string {
	return []string{
		`go\.mod$`,
	}
}

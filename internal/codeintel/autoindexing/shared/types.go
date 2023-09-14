package shared

import "github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"

// IndexConfiguration stores the index configuration for a repository.
type IndexConfiguration struct {
	ID           int
	RepositoryID int
	Data         []byte
}

type InferenceResult struct {
	IndexJobs       []config.IndexJob
	InferenceOutput string
}

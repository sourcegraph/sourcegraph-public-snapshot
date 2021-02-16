package inference

import (
	"regexp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/config"
)

// IndexJobRecognizer infers index jobs from repository structure.
type IndexJobRecognizer interface {
	// CanIndex returns true if the given list of file paths describe a
	// repository for which InferIndexJobs is likely to recognize.
	CanIndex(paths []string, gitserver GitserverClientWrapper) bool

	// InferIndexJobs returns a set of index jobs which are likely to be
	// correct given the list of file paths that describe a repository.
	// The given file paths should be all of the file path matches in the
	// repository that matches any pattern returned from Patterns.
	InferIndexJobs(paths []string, gitserver GitserverClientWrapper) []config.IndexJob

	// Patterns returns a set of regular expressions that match file paths
	// which are of interest to InferIndexJobs.
	Patterns() []*regexp.Regexp
}

// Recognizers is a list of registered index job recognizers.
var Recognizers = map[string]IndexJobRecognizer{
	"go":  lsifGoJobRecognizer{},
	"tsc": lsifTscJobRecognizer{},
}

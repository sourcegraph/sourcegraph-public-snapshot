package inference

import (
	"regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

// IndexJobRecognizer infers index jobs from repository structure.
type IndexJobRecognizer interface {
	// Patterns returns a set of regular expressions that match file paths
	// which are of interest to InferIndexJobs.
	Patterns() []*regexp.Regexp

	// CanIndexRepo returns true if the given list of file paths describe
	// a repository for which InferIndexJobs is likely to generate jobs.
	CanIndexRepo(gitserver GitClient, paths []string) bool

	// InferIndexJobs returns a set of index jobs which are likely to be
	// correct given the list of file paths that describe a repository.
	// The given file paths should be all of the file path matches in the
	// repository that matches any pattern returned from Patterns.
	InferIndexJobs(gitserver GitClient, paths []string) []config.IndexJob
}

// Recognizers is a list of registered index job recognizers.
var Recognizers = map[string]IndexJobRecognizer{
	"go":  recognizer{GoPatterns, CanIndexGoRepo, InferGoIndexJobs},
	"tsc": recognizer{TypeScriptPatterns, CanIndexTypeScriptRepo, InferTypeScriptIndexJobs},
}

type recognizer struct {
	patterns       func() []*regexp.Regexp
	canIndexRepo   func(gitserver GitClient, paths []string) bool
	inferIndexJobs func(gitserver GitClient, paths []string) []config.IndexJob
}

func (r recognizer) Patterns() []*regexp.Regexp {
	return r.patterns()
}

func (r recognizer) CanIndexRepo(gitserver GitClient, paths []string) bool {
	return r.canIndexRepo(gitserver, paths)
}

func (r recognizer) InferIndexJobs(gitserver GitClient, paths []string) []config.IndexJob {
	return r.inferIndexJobs(gitserver, paths)
}

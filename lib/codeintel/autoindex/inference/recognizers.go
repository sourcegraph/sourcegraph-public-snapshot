package inference

import (
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

// IndexJobRecognizer infers index jobs from repository structure.
type IndexJobRecognizer interface {
	// Patterns returns a set of regular expressions that match file paths
	// which are of interest to InferIndexJobs.
	Patterns() []*regexp.Regexp

	// InferIndexJobs returns a set of index jobs which are likely to be
	// correct given the list of file paths that describe a repository.
	// The given file paths should be all of the file path matches in the
	// repository that matches any pattern returned from Patterns.
	InferIndexJobs(gitserver GitClient, paths []string) []config.IndexJob

	// InferIndexJobHints returns a set of hints as to the indexers and index
	// roots that seem likely to support indexing but for which we could/can not
	// infer sufficient index job configurations.
	InferIndexJobHints(gitserver GitClient, paths []string) []config.IndexJobHint
}

// Recognizers is a list of registered index job recognizers.
var Recognizers = map[string]IndexJobRecognizer{
	"go":    recognizer{GoPatterns, InferGoIndexJobs, nil},
	"tsc":   recognizer{TypeScriptPatterns, InferTypeScriptIndexJobs, nil},
	"java":  recognizer{JavaPatterns, InferJavaIndexJobs, InferJavaIndexJobHints},
	"rust":  recognizer{RustPatterns, InferRustIndexJobs, nil},
	"clang": recognizer{ClangPatterns, nil, InferClangIndexJobHints},
}

type recognizer struct {
	patterns           func() []*regexp.Regexp
	inferIndexJobs     func(gitserver GitClient, paths []string) []config.IndexJob
	inferIndexJobHints func(gitserver GitClient, paths []string) []config.IndexJobHint
}

func (r recognizer) Patterns() []*regexp.Regexp {
	return r.patterns()
}

func (r recognizer) InferIndexJobs(gitserver GitClient, paths []string) []config.IndexJob {
	if r.inferIndexJobs == nil {
		return nil
	}
	return r.inferIndexJobs(gitserver, paths)
}

func (r recognizer) InferIndexJobHints(gitserver GitClient, paths []string) []config.IndexJobHint {
	if r.inferIndexJobHints == nil {
		return nil
	}
	return r.inferIndexJobHints(gitserver, paths)
}

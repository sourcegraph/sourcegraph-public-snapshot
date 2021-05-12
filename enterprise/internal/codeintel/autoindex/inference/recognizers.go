package inference

import (
	"context"
	"regexp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
)

// IndexJobRecognizer infers index jobs from repository structure.
type IndexJobRecognizer interface {
	// CanIndexRepo returns true if the given list of file paths describe a
	// repository for which InferIndexJobs is likely to recognize.
	CanIndexRepo(paths []string, gitserver GitserverClientWrapper) bool

	// InferIndexJobs returns a set of index jobs which are likely to be
	// correct given the list of file paths that describe a repository.
	// The given file paths should be all of the file path matches in the
	// repository that matches any pattern returned from Patterns.
	InferIndexJobs(paths []string, gitserver GitserverClientWrapper) []config.IndexJob

	// Patterns returns a set of regular expressions that match file paths
	// which are of interest to InferIndexJobs.
	Patterns() []*regexp.Regexp

	// EnsurePackageRepo checks whether the package is proxied through gitserver,
	// and creates the proxy repo otherwise. The name of the repo is returned.
	EnsurePackageRepo(ctx context.Context, pkg semantic.Package, repoUpdater RepoUpdaterClient, gitserver GitserverClient) (int, string, bool, error)

	// InferPackageIndexJobs returns a set of index jobs which are likely to be correct
	// given a particular package. It's expected that EnsurePackageRepo is called prior to
	// this function.
	InferPackageIndexJobs(ctx context.Context, pkg semantic.Package, gitserver GitserverClientWrapper) ([]config.IndexJob, error)
}

// Recognizers is a list of registered index job recognizers.
var Recognizers = map[string]IndexJobRecognizer{
	"go":  lsifGoJobRecognizer{},
	"tsc": lsifTscJobRecognizer{},
}

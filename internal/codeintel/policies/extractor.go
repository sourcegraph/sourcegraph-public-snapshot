package policies

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
)

// Extractor returns a max age and include intermediate commits flags from a policy. These fields exist for
// both data retention and auto-index scheduling.
type Extractor func(policy shared.ConfigurationPolicy) (maxAge *time.Duration, includeIntermediateCommits bool)

// NoopExtractor returns nil and false.
func NoopExtractor(policy shared.ConfigurationPolicy) (*time.Duration, bool) {
	return nil, false
}

// RetentionExtractor returns the max age of a precise code intelligence upload the given policy as well as a
// flag indicating whether commits on branches (but not the tip) should be included.
func RetentionExtractor(policy shared.ConfigurationPolicy) (*time.Duration, bool) {
	return policy.RetentionDuration, policy.RetainIntermediateCommits
}

// IndexingExtractor returns the max age of a commit that can be auto-indexed the given policy as well as a
// flag indicating whether commits on branches (but not the tip) should be included.
func IndexingExtractor(policy shared.ConfigurationPolicy) (*time.Duration, bool) {
	return policy.IndexCommitMaxAge, policy.IndexIntermediateCommits
}

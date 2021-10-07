package janitor

import (
	"context"
	"time"

	"github.com/gobwas/glob"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
)

func isUploadCommitProtectedByPolicy(
	ctx context.Context,
	policies []dbstore.ConfigurationPolicy,
	patterns map[string]glob.Glob,
	refDescriptions map[string][]gitserver.RefDescription,
	repositoryState *repositoryExpirationState,
	upload dbstore.Upload,
	commits []string,
	now time.Time,
) (bool, error) {
	filteredCommits := make([]string, 0, len(commits))
	for _, commit := range commits {
		// See if this commit was already shown to be protected
		if repositoryState.IsProtected(commit) {
			return true, nil
		}

		// Never expire tip of the default branch
		for _, refDescription := range refDescriptions[commit] {
			if refDescription.IsDefaultBranch {
				repositoryState.MarkProtected(commit)
				return true, nil
			}
		}

		// See if this commit was shown to be unprotected through this upload's timeframe. We may have
		// previously been able to tell that a commit cannot be protected by any upload after a certain
		// date, which can potentially short-circuit a large number of both policy comparisons and calls
		// to gitserver.
		if threshold, ok := repositoryState.UnprotectedUntil(commit); ok && upload.FinishedAt.Before(now.Add(-threshold)) {
			continue
		}

		filteredCommits = append(filteredCommits, commit)
	}

	// Try fast path first to avoid another gitserver query. We check _all_ queries in this batch
	// first, as if we find one protected commit we can skip checking all of the others on the slow
	// path as well.
	if ok, err := isUploadCommitProtectedByPolicyFastPath(
		policies,
		patterns,
		refDescriptions,
		repositoryState,
		upload,
		filteredCommits,
		now,
	); err != nil || ok {
		return ok, err
	}

	// Fall back to slow path
	if ok, err := isUploadCommitProtectedByPolicySlowPath(
		ctx,
		policies,
		patterns,
		repositoryState,
		upload,
		filteredCommits,
		now,
	); err != nil || ok {
		return ok, err
	}

	// If this upload is not found to be protected at this point, then all of the commits from which
	// this commit is visible are also not protected _up to the time of the commit_. We only need to
	// re-check this commit in the future against a policy with a strictly shorter retention duration.
	// In the following loop, we find the largest retention duration of any policy that was not just
	// checked in the loops above.

	var nextLargestDuration time.Duration

	for _, policy := range policies {
		if policy.RetentionDuration == nil || now.Sub(*upload.FinishedAt) <= *policy.RetentionDuration {
			continue
		}

		if nextLargestDuration == 0 || nextLargestDuration < *policy.RetentionDuration {
			nextLargestDuration = *policy.RetentionDuration
		}
	}

	if nextLargestDuration > 0 {
		for _, commit := range filteredCommits {
			repositoryState.MarkUnprotectedUntil(commit, nextLargestDuration)
		}
	}

	return false, nil
}

// isUploadCommitProtectedByPolicyFastPath uses the information we already have about the tips of the
// repository's branches and tags. We will not be able to complete the protection check in this step as
// we don't yet have the data to consider commits contained by a branch, or policies with the retain
// intermediate commits option enabled. This will be completed in the next step, if the upload is not
// shown to be protected in this "fast path".
func isUploadCommitProtectedByPolicyFastPath(
	policies []dbstore.ConfigurationPolicy,
	patterns map[string]glob.Glob,
	refDescriptions map[string][]gitserver.RefDescription,
	repositoryState *repositoryExpirationState,
	upload dbstore.Upload,
	commits []string,
	now time.Time,
) (bool, error) {
	filters := []policyFilter{
		policyCoversTime(now.Sub(*upload.FinishedAt)),
	}

	return isUploadProtectedByPolicy(repositoryState, commits, func(commit string) (bool, error) {
		return filterAndMatchPolicies(policies, filters, []policyMatcher{
			makeTipPolicyMatcher(patterns, commit, refDescriptions[commit]),
		}), nil
	})
}

// isUploadCommitProtectedByPolicySlowPath completes the protection check by considering policies with
// the retain intermediate commits flag enabled. This does not necessarily mean that the branch defines
// the tip of the branch; that was already checked in the preceding loop. Commits contained by a branch
// are queried from gitserver on demand. Gitserver responses are stored in an in-memory LRU cache.
func isUploadCommitProtectedByPolicySlowPath(
	ctx context.Context,
	policies []dbstore.ConfigurationPolicy,
	patterns map[string]glob.Glob,
	repositoryState *repositoryExpirationState,
	upload dbstore.Upload,
	commits []string,
	now time.Time,
) (_ bool, err error) {
	filters := []policyFilter{
		policyRetainsIntermediateCommits,
		policyCoversTime(now.Sub(*upload.FinishedAt)),
	}

	return isUploadProtectedByPolicy(repositoryState, commits, func(commit string) (bool, error) {
		branches, err := repositoryState.GetBranchesFor(ctx, commit)
		if err != nil {
			return false, err
		}

		return filterAndMatchPolicies(policies, filters, []policyMatcher{
			makeContainsPolicyMatcher(patterns, commit, branches),
		}), nil
	})
}

type commitMatcher func(commit string) (bool, error)
type policyFilter func(policy dbstore.ConfigurationPolicy) bool
type policyMatcher func(policy dbstore.ConfigurationPolicy) bool

// isUploadProtectedByPolicy returns true if the given matcher returns true when invoked with
// any of the given commit values. The first found passing commit will be marked as protected
// in teh given repository expiration state object. This function short-circuits and marks at
// most one commit as protected per invocation.
func isUploadProtectedByPolicy(repositoryState *repositoryExpirationState, commits []string, matcher commitMatcher) (bool, error) {
	for _, commit := range commits {
		ok, err := matcher(commit)
		if err != nil {
			return false, err
		}

		if ok {
			repositoryState.MarkProtected(commit)
			return true, nil
		}
	}

	return false, nil
}

// policyCoversTime creates a policy filter that returns true if the given policy's retention
// duration is no smaller than the given duration. Policies with no retention duration matches
// commits regardless of time, hence will always return true for such policies.
func policyCoversTime(duration time.Duration) policyFilter {
	return func(policy dbstore.ConfigurationPolicy) bool {
		return policy.RetentionDuration == nil || duration <= *policy.RetentionDuration
	}
}

// policyRetainsIntermediateCommits is a policyFilter that returns true if the policy considers
// commits that are contained in a branch but do not define the branch's tip.
func policyRetainsIntermediateCommits(policy dbstore.ConfigurationPolicy) bool {
	return policy.RetainIntermediateCommits
}

// filterAndMatchPolicies applies the given policy matchers over the policies that pass ALL of the
// given policy filters. This function returns true if any policy/matcher combination returns true.
func filterAndMatchPolicies(policies []dbstore.ConfigurationPolicy, filters []policyFilter, matchers []policyMatcher) bool {
policyLoop:
	for _, policy := range policies {
		for _, filter := range filters {
			if !filter(policy) {
				continue policyLoop
			}
		}

		for _, matcher := range matchers {
			if matcher(policy) {
				return true
			}
		}
	}

	return false
}

// makeTipPolicyMatcher creates a policy matcher that returns true when the given policy pattern
// matches the given commit, or is the commit defining a tag or the tip of a branch in the given
// reference descriptions.
func makeTipPolicyMatcher(patterns map[string]glob.Glob, commit string, refDescriptions []gitserver.RefDescription) policyMatcher {
	branches, tags := refNamesByType(refDescriptions)

	return func(policy dbstore.ConfigurationPolicy) bool {
		if policy.Type == dbstore.GitObjectTypeCommit && patterns[policy.Pattern].Match(commit) {
			return true
		} else if policy.Type == dbstore.GitObjectTypeTag && patternMatchesAnyValue(patterns[policy.Pattern], tags) {
			return true
		} else if policy.Type == dbstore.GitObjectTypeTree && patternMatchesAnyValue(patterns[policy.Pattern], branches) {
			return true
		}

		return false
	}
}

// makeContainsPolicyMatcher creates a policy matcher that returns true when the given policy pattern
// matches one of the given branch names.
func makeContainsPolicyMatcher(patterns map[string]glob.Glob, commit string, branches []string) policyMatcher {
	return func(policy dbstore.ConfigurationPolicy) bool {
		return patternMatchesAnyValue(patterns[policy.Pattern], branches)
	}
}

// refNamesByType returns slices of names of the given references description bucketed by their type.
func refNamesByType(refDescriptions []gitserver.RefDescription) (branches, tags []string) {
	branches = make([]string, 0, len(refDescriptions))
	tags = make([]string, 0, len(refDescriptions))

	for _, refDescription := range refDescriptions {
		if refDescription.Type == gitserver.RefTypeBranch {
			branches = append(branches, refDescription.Name)
		} else if refDescription.Type == gitserver.RefTypeTag {
			tags = append(tags, refDescription.Name)
		}
	}

	return branches, tags
}

// patternMatchesAnyValue returns true if the given pattern matches at least one of the given values.
func patternMatchesAnyValue(pattern glob.Glob, values []string) bool {
	for _, value := range values {
		if pattern.Match(value) {
			return true
		}
	}

	return false
}

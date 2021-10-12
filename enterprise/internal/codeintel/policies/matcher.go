package policies

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gobwas/glob"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

type Matcher struct {
	gitserverClient           GitserverClient
	extractor                 Extractor
	includeTipOfDefaultBranch bool
	filterByCreatedDate       bool
}

// PolicyMatch indicates the name of the matching branch or tag associated with some commit. The policy
// identifier field is set unless the policy match exists due to a `includeTipOfDefaultBranch` match. The
// policy duration field is set if the matching policy specifies a duration.
type PolicyMatch struct {
	Name           string
	PolicyID       *int
	PolicyDuration *time.Duration
}

type matcherContext struct {
	repositoryID   int
	policies       []dbstore.ConfigurationPolicy
	patterns       map[string]glob.Glob
	commitMap      map[string][]PolicyMatch
	branchRequests map[string]branchRequestMeta
}

// branchRequestMeta is bookeeping for the types of gitserver queries we'll need to make to resolve commits
// that are contained on branches but are not their tip.
type branchRequestMeta struct {
	isDefaultBranch     bool
	policyDurationByIDs map[int]*time.Duration
}

func NewMatcher(
	gitserverClient GitserverClient,
	extractor Extractor,
	includeTipOfDefaultBranch bool,
	filterByCreatedDate bool,
) *Matcher {
	return &Matcher{
		gitserverClient:           gitserverClient,
		extractor:                 extractor,
		includeTipOfDefaultBranch: includeTipOfDefaultBranch,
		filterByCreatedDate:       filterByCreatedDate,
	}
}

// CommitsDescribedByPolicy returns a map from commits within the given repository to a set of policy matches
// with respect to the given policies.
//
// If includeTipOfDefaultBranch is true, the there will exist a match for the tip default branch with a nil
// policy identifier, even if no policies are supplied. This is set to true for data retention but not for
// auto-indexing.
//
// If filterByCreatedDate is true, then commits that are older than the matching policy duration will be
// filtered out. If false, policy duration is not considered. This is set to true for auto-indexing, but false
// for data retention as we need to compare the policy duration against the associated upload date, not the
// commit date.
func (m *Matcher) CommitsDescribedByPolicy(ctx context.Context, repositoryID int, policies []dbstore.ConfigurationPolicy, now time.Time) (map[string][]PolicyMatch, error) {
	if len(policies) == 0 && !m.includeTipOfDefaultBranch {
		return nil, nil
	}

	patterns, err := compilePatterns(policies)
	if err != nil {
		return nil, err
	}

	// mutable context
	context := matcherContext{
		repositoryID:   repositoryID,
		policies:       policies,
		patterns:       patterns,
		commitMap:      map[string][]PolicyMatch{},
		branchRequests: map[string]branchRequestMeta{},
	}

	refDescriptions, err := m.gitserverClient.RefDescriptions(ctx, repositoryID)
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.RefDescriptions")
	}

	for commit, refDescriptions := range refDescriptions {
		for _, refDescription := range refDescriptions {
			switch refDescription.Type {
			case gitserver.RefTypeTag:
				// Resolve tagged commits
				m.resolveTagReference(context, commit, refDescription, now)

			case gitserver.RefTypeBranch:
				// Resolve tips of branches
				m.resolveBranchReference(context, commit, refDescription, now)
			}
		}
	}

	// Resolve commits on branches but not at tip
	if err := m.resolveBranchMembership(ctx, context, now); err != nil {
		return nil, err
	}

	// Resolve comments via rev-parse
	if err := m.resolveCommitPolicies(ctx, context, now); err != nil {
		return nil, err
	}

	return context.commitMap, nil
}

func (m *Matcher) resolveTagReference(context matcherContext, commit string, refDescription gitserver.RefDescription, now time.Time) {
	visitor := func(policy dbstore.ConfigurationPolicy) {
		policyDuration, _ := m.extractor(policy)

		context.commitMap[commit] = append(context.commitMap[commit], PolicyMatch{
			Name:           refDescription.Name,
			PolicyID:       &policy.ID,
			PolicyDuration: policyDuration,
		})
	}

	m.forEachMatchingPolicy(context, refDescription, dbstore.GitObjectTypeTag, visitor, now)
}

func (m *Matcher) resolveBranchReference(context matcherContext, commit string, refDescription gitserver.RefDescription, now time.Time) {
	// Add fake match for tip of default branch
	if refDescription.IsDefaultBranch && m.includeTipOfDefaultBranch {
		context.commitMap[commit] = append(context.commitMap[commit], PolicyMatch{
			Name:           refDescription.Name,
			PolicyID:       nil,
			PolicyDuration: nil,
		})
	}

	visitor := func(policy dbstore.ConfigurationPolicy) {
		policyDuration, _ := m.extractor(policy)

		context.commitMap[commit] = append(context.commitMap[commit], PolicyMatch{
			Name:           refDescription.Name,
			PolicyID:       &policy.ID,
			PolicyDuration: policyDuration,
		})

		// If we include intermediate commits for this policy, we need to query the
		// set of comits that belong to any branch matching this policy's pattern.
		// We store this information in the branchRequests map so that we perform a
		// query for each matching branch only once later.
		if policyDuration, includeIntermediateCommits := m.extractor(policy); includeIntermediateCommits {
			meta, ok := context.branchRequests[refDescription.Name]
			if !ok {
				meta.policyDurationByIDs = map[int]*time.Duration{}
			}

			meta.policyDurationByIDs[policy.ID] = policyDuration
			meta.isDefaultBranch = meta.isDefaultBranch || refDescription.IsDefaultBranch
			context.branchRequests[refDescription.Name] = meta
		}
	}

	m.forEachMatchingPolicy(context, refDescription, dbstore.GitObjectTypeTree, visitor, now)
}

func (m *Matcher) resolveBranchMembership(ctx context.Context, context matcherContext, now time.Time) error {
	for branchName, branchRequestMeta := range context.branchRequests {
		maxCommitAge := getMaxAge(branchRequestMeta.policyDurationByIDs, now)

		if !m.filterByCreatedDate {
			// Do not filter out any commits by date
			maxCommitAge = nil
		}

		commitDates, err := m.gitserverClient.CommitsUniqueToBranch(ctx, context.repositoryID, branchName, branchRequestMeta.isDefaultBranch, maxCommitAge)
		if err != nil {
			return errors.Wrap(err, "gitserver.CommitsUniqueToBranch")
		}

		for commit, commitDate := range commitDates {
		policyLoop:
			for policyID, policyDuration := range branchRequestMeta.policyDurationByIDs {
				for _, match := range context.commitMap[commit] {
					if match.PolicyID != nil && *match.PolicyID == policyID {
						// Skip duplicates (can happen at head of branch)
						continue policyLoop
					}
				}

				if m.filterByCreatedDate && policyDuration != nil && now.Sub(commitDate) > *policyDuration {
					// Policy duration was less than max age and re-check failed
					continue policyLoop
				}

				// Don't capture loop variable pointer
				localPolicyID := policyID

				context.commitMap[commit] = append(context.commitMap[commit], PolicyMatch{
					Name:           branchName,
					PolicyID:       &localPolicyID,
					PolicyDuration: policyDuration,
				})
			}
		}
	}

	return nil
}

func (m *Matcher) resolveCommitPolicies(ctx context.Context, context matcherContext, now time.Time) error {
	for _, policy := range context.policies {
		if policy.Type == dbstore.GitObjectTypeCommit {
			commitDate, err := m.gitserverClient.CommitDate(ctx, context.repositoryID, policy.Pattern)
			if err != nil {
				if errcode.IsNotFound(err) {
					return nil
				}

				return errors.Wrap(err, "gitserver.ResolveRevision")
			}

			policyDuration, _ := m.extractor(policy)

			if m.filterByCreatedDate && policyDuration != nil && now.Sub(commitDate) > *policyDuration {
				continue
			}

			context.commitMap[policy.Pattern] = append(context.commitMap[policy.Pattern], PolicyMatch{
				Name:           policy.Pattern,
				PolicyID:       &policy.ID,
				PolicyDuration: policyDuration,
			})
		}
	}

	return nil
}

func (m *Matcher) forEachMatchingPolicy(context matcherContext, refDescription gitserver.RefDescription, targetObjectType dbstore.GitObjectType, f func(policy dbstore.ConfigurationPolicy), now time.Time) {
	for _, policy := range context.policies {
		if policy.Type == targetObjectType && m.policyMatchesRefDescription(context, policy, refDescription, now) {
			f(policy)
		}
	}
}

func (m *Matcher) policyMatchesRefDescription(context matcherContext, policy dbstore.ConfigurationPolicy, refDescription gitserver.RefDescription, now time.Time) bool {
	if !context.patterns[policy.Pattern].Match(refDescription.Name) {
		// Name doesn't match policy's pattern
		return false
	}

	if policyDuration, _ := m.extractor(policy); m.filterByCreatedDate && policyDuration != nil && now.Sub(refDescription.CreatedDate) > *policyDuration {
		// Policy is not unbounded, we are filtering by commit date, commit is moo old
		return false
	}

	return true
}

// compilePatterns constructs a map from patterns in each given policy to a compiled glob object used
// to match to commits, branch names, and tag names. If there are multiple policies with the same pattern,
// the pattern is compiled only once.
func compilePatterns(policies []dbstore.ConfigurationPolicy) (map[string]glob.Glob, error) {
	patterns := make(map[string]glob.Glob, len(policies))
	for _, policy := range policies {
		if _, ok := patterns[policy.Pattern]; ok || policy.Type == dbstore.GitObjectTypeCommit {
			continue
		}

		pattern, err := glob.Compile(policy.Pattern)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to compile glob pattern `%s` in configuration policy %d", policy.Pattern, policy.ID))
		}

		patterns[policy.Pattern] = pattern
	}

	return patterns, nil
}

func getMaxAge(policyDurationByIDs map[int]*time.Duration, now time.Time) *time.Time {
	var maxDuration *time.Duration
	for _, duration := range policyDurationByIDs {
		if duration == nil {
			// If any duration is nil, the policy is unbounded
			return nil
		}
		if maxDuration == nil || *maxDuration < *duration {
			maxDuration = duration
		}
	}
	if maxDuration == nil {
		return nil
	}

	maxAge := now.Add(-*maxDuration)
	return &maxAge
}

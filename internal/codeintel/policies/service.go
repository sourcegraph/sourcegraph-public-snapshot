package policies

import (
	"context"
	"sort"
	"time"

	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ service = (*Service)(nil)

type service interface {
	// Not used yet.
	List(ctx context.Context, opts ListOpts) (policies []Policy, err error)
	Get(ctx context.Context, id int) (policy Policy, ok bool, err error)
	Create(ctx context.Context, policy Policy) (hydratedPolicy Policy, err error)
	Update(ctx context.Context, policy Policy) (hydratedPolicy Policy, err error)
	Delete(ctx context.Context, id int) (err error)
	CommitsMatchingRetentionPolicies(ctx context.Context, repoID int, policies []Policy, instant time.Time, commitSubset ...string) (commitsToPolicies map[string][]Policy, err error)
	CommitsMatchingIndexingPolicies(ctx context.Context, repoID int, policies []Policy, instant time.Time) (commitsToPolicies map[string][]Policy, err error)

	// Configurations
	GetConfigurationPolicies(ctx context.Context, opts shared.GetConfigurationPoliciesOptions) ([]shared.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (_ shared.ConfigurationPolicy, _ bool, err error)
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy shared.ConfigurationPolicy) (shared.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy shared.ConfigurationPolicy) (err error)
	DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error)

	GetRetentionPolicyOverview(ctx context.Context, upload shared.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []shared.RetentionPolicyMatchCandidate, totalCount int, err error)
}

type Service struct {
	store      store.Store
	uploadSvc  UploadService
	gitserver  GitserverClient
	operations *operations
}

func newService(policiesStore store.Store, uploadSvc UploadService, gitserver GitserverClient, observationContext *observation.Context) *Service {
	return &Service{
		store:      policiesStore,
		uploadSvc:  uploadSvc,
		gitserver:  gitserver,
		operations: newOperations(observationContext),
	}
}

type Policy = shared.Policy

type ListOpts struct {
	Limit int
}

func (s *Service) getPolicyMatcherFactory(gitserver GitserverClient, extractor policies.Extractor, includeTipOfDefaultBranch bool, filterByCreatedDate bool) *policies.Matcher {
	return policies.NewMatcher(gitserver, extractor, includeTipOfDefaultBranch, filterByCreatedDate)
}

func (s *Service) List(ctx context.Context, opts ListOpts) (policies []Policy, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.List(ctx, store.ListOpts(opts))
}

func (s *Service) Get(ctx context.Context, id int) (policy Policy, ok bool, err error) {
	ctx, _, endObservation := s.operations.get.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33376
	_ = ctx
	return Policy{}, false, errors.Newf("unimplemented: policies.Get")
}

func (s *Service) Create(ctx context.Context, policy Policy) (hydratedPolicy Policy, err error) {
	ctx, _, endObservation := s.operations.create.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33376
	_ = ctx
	return Policy{}, errors.Newf("unimplemented: policies.Create")
}

func (s *Service) Update(ctx context.Context, policy Policy) (hydratedPolicy Policy, err error) {
	ctx, _, endObservation := s.operations.update.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33376
	_ = ctx
	return Policy{}, errors.Newf("unimplemented: policies.Update")
}

func (s *Service) Delete(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.delete.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33376
	_ = ctx
	return errors.Newf("unimplemented: policies.Delete")
}

func (s *Service) CommitsMatchingRetentionPolicies(ctx context.Context, repoID int, policies []Policy, instant time.Time, commitSubset ...string) (commitsToPolicies map[string][]Policy, err error) {
	ctx, _, endObservation := s.operations.commitsMatchingRetentionPolicies.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33376
	_ = ctx
	return nil, errors.Newf("unimplemented: policies.CommitsMatchingRetentionPolicies")
}

func (s *Service) CommitsMatchingIndexingPolicies(ctx context.Context, repoID int, policies []Policy, instant time.Time) (commitsToPolicies map[string][]Policy, err error) {
	ctx, _, endObservation := s.operations.commitsMatchingIndexingPolicies.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33376
	_ = ctx
	return nil, errors.Newf("unimplemented: policies.CommitsMatchingIndexingPolicies")
}

func (s *Service) GetConfigurationPolicies(ctx context.Context, opts shared.GetConfigurationPoliciesOptions) (_ []shared.ConfigurationPolicy, totalCount int, err error) {
	ctx, _, endObservation := s.operations.getConfigurationPolicies.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetConfigurationPolicies(ctx, opts)
}

func (s *Service) GetConfigurationPolicyByID(ctx context.Context, id int) (_ shared.ConfigurationPolicy, _ bool, err error) {
	ctx, _, endObservation := s.operations.getConfigurationPoliciesByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetConfigurationPolicyByID(ctx, id)
}

func (s *Service) CreateConfigurationPolicy(ctx context.Context, configurationPolicy shared.ConfigurationPolicy) (_ shared.ConfigurationPolicy, err error) {
	ctx, _, endObservation := s.operations.createConfigurationPolicy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.CreateConfigurationPolicy(ctx, configurationPolicy)
}

func (s *Service) UpdateConfigurationPolicy(ctx context.Context, policy shared.ConfigurationPolicy) (err error) {
	ctx, _, endObservation := s.operations.updateConfigurationPolicy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateConfigurationPolicy(ctx, policy)
}

func (s *Service) DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.deleteConfigurationPolicyByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteConfigurationPolicyByID(ctx, id)
}

func (s *Service) GetRetentionPolicyOverview(ctx context.Context, upload shared.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []shared.RetentionPolicyMatchCandidate, totalCount int, err error) {
	policyMatcher := s.getPolicyMatcherFactory(s.gitserver, policies.RetentionExtractor, true, false)

	policies, _, err := s.GetConfigurationPolicies(ctx, shared.GetConfigurationPoliciesOptions{
		RepositoryID:     upload.RepositoryID,
		Term:             query,
		ForDataRetention: true,
		Limit:            first,
		Offset:           int(after),
	})
	if err != nil {
		return nil, 0, err
	}

	visibileCommits, err := s.getCommitsVisibleToUpload(ctx, upload)
	if err != nil {
		return nil, 0, err
	}

	matchingPolicies, err := policyMatcher.CommitsDescribedByPolicyInternal(ctx, upload.RepositoryID, policies, time.Now(), visibileCommits...)
	if err != nil {
		return nil, 0, err
	}

	var (
		potentialMatchIndexSet map[int]int // map of polciy ID to array index
		potentialMatches       []shared.RetentionPolicyMatchCandidate
	)

	potentialMatches, potentialMatchIndexSet = s.populateMatchingCommits(visibileCommits, upload, matchingPolicies, policies, now)

	if !matchesOnly {
		// populate with remaining unmatched policies
		for _, policy := range policies {
			policy := policy
			if _, ok := potentialMatchIndexSet[policy.ID]; !ok {
				potentialMatches = append(potentialMatches, shared.RetentionPolicyMatchCandidate{
					ConfigurationPolicy: &policy,
					Matched:             false,
				})
			}
		}
	}

	sort.Slice(potentialMatches, func(i, j int) bool {
		if potentialMatches[i].ConfigurationPolicy == nil {
			return true
		} else if potentialMatches[j].ConfigurationPolicy == nil {
			return false
		}
		return potentialMatches[i].ID < potentialMatches[j].ID
	})

	return potentialMatches, len(potentialMatches), nil
}

func (s *Service) getCommitsVisibleToUpload(ctx context.Context, upload shared.Upload) (commits []string, err error) {
	var token *string
	for first := true; first || token != nil; first = false {
		cs, nextToken, err := s.uploadSvc.GetCommitsVisibleToUpload(ctx, upload.ID, 50, token)
		if err != nil {
			return nil, errors.Wrap(err, "uploadSvc.GetCommitsVisibleToUpload")
		}
		token = nextToken

		commits = append(commits, cs...)
	}

	return
}

// populateMatchingCommits builds a slice of all retention policies that, either directly or via
// a visibile upload, apply to the upload. It returns the slice of policies and the set of matching
// policy IDs mapped to their index in the slice.
func (s *Service) populateMatchingCommits(
	visibileCommits []string,
	upload shared.Upload,
	matchingPolicies map[string][]policies.PolicyMatch,
	policies []shared.ConfigurationPolicy,
	now time.Time,
) ([]shared.RetentionPolicyMatchCandidate, map[int]int) {
	var (
		potentialMatches       = make([]shared.RetentionPolicyMatchCandidate, 0, len(policies))
		potentialMatchIndexSet = make(map[int]int, len(policies))
	)

	// First add all matches for the commit of this upload. We do this to ensure that if a policy matches both the upload's commit
	// and a visible commit, we ensure an entry for that policy is only added for the upload's commit. This makes the logic in checking
	// the visible commits a bit simpler, as we don't have to check if policy X has already been added for a visible commit in the case
	// that the upload's commit is not first in the list.
	if policyMatches, ok := matchingPolicies[upload.Commit]; ok {
		for _, policyMatch := range policyMatches {
			if policyMatch.PolicyDuration == nil || now.Sub(upload.UploadedAt) < *policyMatch.PolicyDuration {
				policyID := -1
				if policyMatch.PolicyID != nil {
					policyID = *policyMatch.PolicyID
				}
				potentialMatches = append(potentialMatches, shared.RetentionPolicyMatchCandidate{
					ConfigurationPolicy: policyByID(policies, policyID),
					Matched:             true,
				})
				potentialMatchIndexSet[policyID] = len(potentialMatches) - 1
			}
		}
	}

	for _, commit := range visibileCommits {
		if commit == upload.Commit {
			continue
		}
		if policyMatches, ok := matchingPolicies[commit]; ok {
			for _, policyMatch := range policyMatches {
				if policyMatch.PolicyDuration == nil || now.Sub(upload.UploadedAt) < *policyMatch.PolicyDuration {
					policyID := -1
					if policyMatch.PolicyID != nil {
						policyID = *policyMatch.PolicyID
					}
					if index, ok := potentialMatchIndexSet[policyID]; ok && potentialMatches[index].ProtectingCommits != nil {
						//  If an entry for the policy already exists and it has > 1 "protecting commits", add this commit too.
						potentialMatches[index].ProtectingCommits = append(potentialMatches[index].ProtectingCommits, commit)
					} else if !ok {
						// Else if theres no entry for the policy, create an entry with this commit as the first "protecting commit".
						// This should never override an entry for a policy matched directly, see the first comment on how this is avoided.
						potentialMatches = append(potentialMatches, shared.RetentionPolicyMatchCandidate{
							ConfigurationPolicy: policyByID(policies, policyID),
							Matched:             true,
							ProtectingCommits:   []string{commit},
						})
						potentialMatchIndexSet[policyID] = len(potentialMatches) - 1
					}
				}
			}
		}
	}

	return potentialMatches, potentialMatchIndexSet
}

func policyByID(policies []shared.ConfigurationPolicy, id int) *shared.ConfigurationPolicy {
	if id == -1 {
		return nil
	}
	for _, policy := range policies {
		if policy.ID == id {
			return &policy
		}
	}
	return nil
}

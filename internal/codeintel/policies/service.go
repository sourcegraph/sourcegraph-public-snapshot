package policies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	policiesStore store.Store
	operations    *operations
}

func newService(policiesStore store.Store, observationContext *observation.Context) *Service {
	return &Service{
		policiesStore: policiesStore,
		operations:    newOperations(observationContext),
	}
}

type Policy = shared.Policy

type ListOpts struct {
	Limit int
}

func (s *Service) List(ctx context.Context, opts ListOpts) (policies []Policy, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.policiesStore.List(ctx, store.ListOpts(opts))
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

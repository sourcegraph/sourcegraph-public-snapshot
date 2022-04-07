package policies

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	policiesStore Store
	operations    *operations
}

func newService(policiesStore Store, observationContext *observation.Context) *Service {
	return &Service{
		policiesStore: policiesStore,
		operations:    newOperations(observationContext),
	}
}

type Policy struct {
	// TODO
}

type ListOpts struct {
	// TODO
}

func (s *Service) List(ctx context.Context, opts ListOpts) (policies []Policy, err error) {
	ctx, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}

func (s *Service) Get(ctx context.Context, id int) (policy Policy, ok bool, err error) {
	ctx, endObservation := s.operations.get.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return Policy{}, false, nil
}

func (s *Service) Create(ctx context.Context, policy Policy) (hydratedPolicy Policy, err error) {
	ctx, endObservation := s.operations.create.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return Policy{}, nil
}

func (s *Service) Update(ctx context.Context, policy Policy) (hydratedPolicy Policy, err error) {
	ctx, endObservation := s.operations.update.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return Policy{}, nil
}

func (s *Service) Delete(ctx context.Context, id int) (err error) {
	ctx, endObservation := s.operations.delete.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil
}

// TODO
func (s *Service) FindMatches(ctx context.Context) (err error) {
	ctx, endObservation := s.operations.findMatches.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil
}

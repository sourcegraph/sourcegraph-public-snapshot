package sentinel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	store      store.Store
	operations *operations
}

func newService(
	observationCtx *observation.Context,
	store store.Store,
) *Service {
	return &Service{
		store:      store,
		operations: newOperations(observationCtx),
	}
}

func (s *Service) Foo(ctx context.Context) (err error) {
	ctx, _, endObservation := s.operations.foo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	_ = ctx
	return nil
}

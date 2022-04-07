package documents

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	documentsStore Store
	operations     *operations
}

func newService(documentsStore Store, observationContext *observation.Context) *Service {
	return &Service{
		documentsStore: documentsStore,
		operations:     newOperations(observationContext),
	}
}

type Document struct {
	// TODO
}

type DocumentOpts struct {
	// TODO
}

func (s *Service) Document(ctx context.Context, opts DocumentOpts) (documents []Document, err error) {
	ctx, endObservation := s.operations.document.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}

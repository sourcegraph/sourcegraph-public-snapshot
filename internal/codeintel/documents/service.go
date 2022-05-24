package documents

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/documents/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/documents/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	documentsStore store.Store
	operations     *operations
}

func newService(documentsStore store.Store, observationContext *observation.Context) *Service {
	return &Service{
		documentsStore: documentsStore,
		operations:     newOperations(observationContext),
	}
}

type Document = shared.Document

type DocumentOpts struct{}

func (s *Service) Document(ctx context.Context, opts DocumentOpts) (documents []Document, err error) {
	ctx, _, endObservation := s.operations.document.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33373
	_, _ = s.documentsStore.List(ctx, store.ListOpts{})
	return nil, errors.Newf("unimplemented: documents.Document")
}

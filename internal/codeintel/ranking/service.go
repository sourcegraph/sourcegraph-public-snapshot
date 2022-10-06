package ranking

import (
	"context"
	"errors"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	uploadSvc  *uploads.Service
	operations *operations
	logger     log.Logger
}

func newService(
	uploadSvc *uploads.Service,
	observationContext *observation.Context,
) *Service {
	return &Service{
		uploadSvc:  uploadSvc,
		operations: newOperations(observationContext),
		logger:     observationContext.Logger,
	}
}

func (s *Service) GetRepoRank(ctx context.Context, repoName api.RepoName) (_ float64, err error) {
	_, _, endObservation := s.operations.getRepoRank.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return 0, errors.New("codeintel.ranking.service.GetRepoRank unimplemented")
}

func (s *Service) GetDocumentRanks(ctx context.Context, repoName api.RepoName) (_ map[string]float64, err error) {
	_, _, endObservation := s.operations.getDocumentRanks.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return nil, errors.New("codeintel.ranking.service.GetDocumentRanks unimplemented")
}

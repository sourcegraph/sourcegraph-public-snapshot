package sentinel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
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

func (s *Service) GetVulnerabilitiesByIDs(ctx context.Context, ids ...int) ([]shared.Vulnerability, error) {
	return s.store.GetVulnerabilitiesByIDs(ctx, ids...)
}

func (s *Service) VulnerabilityByID(ctx context.Context, id int) (shared.Vulnerability, bool, error) {
	return s.store.VulnerabilityByID(ctx, id)
}

func (s *Service) VulnerabilityMatchByID(ctx context.Context, id int) (shared.VulnerabilityMatch, bool, error) {
	return s.store.VulnerabilityMatchByID(ctx, id)
}

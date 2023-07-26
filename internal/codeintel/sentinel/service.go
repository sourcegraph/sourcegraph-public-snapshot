package sentinel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"
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

func (s *Service) GetVulnerabilities(ctx context.Context, args shared.GetVulnerabilitiesArgs) ([]shared.Vulnerability, int, error) {
	return s.store.GetVulnerabilities(ctx, args)
}

func (s *Service) GetVulnerabilityMatches(ctx context.Context, args shared.GetVulnerabilityMatchesArgs) ([]shared.VulnerabilityMatch, int, error) {
	return s.store.GetVulnerabilityMatches(ctx, args)
}

func (s *Service) GetVulnerabilityMatchesSummaryCounts(ctx context.Context) (shared.GetVulnerabilityMatchesSummaryCounts, error) {
	return s.store.GetVulnerabilityMatchesSummaryCount(ctx)
}

func (s *Service) GetVulnerabilityMatchesCountByRepository(ctx context.Context, args shared.GetVulnerabilityMatchesCountByRepositoryArgs) ([]shared.VulnerabilityMatchesByRepository, int, error) {
	return s.store.GetVulnerabilityMatchesCountByRepository(ctx, args)
}

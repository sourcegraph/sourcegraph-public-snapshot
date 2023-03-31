package sentinel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/shared"
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

type (
	Vulnerability                                = shared.Vulnerability
	VulnerabilityMatch                           = shared.VulnerabilityMatch
	GetVulnerabilitiesArgs                       = shared.GetVulnerabilitiesArgs
	GetVulnerabilityMatchesArgs                  = shared.GetVulnerabilityMatchesArgs
	GetVulnerabilityMatchesSummaryCounts         = shared.GetVulnerabilityMatchesSummaryCounts
	GetVulnerabilityMatchesCountByRepositoryArgs = shared.GetVulnerabilityMatchesCountByRepositoryArgs
	VulnerabilityMatchesByRepository             = shared.VulnerabilityMatchesByRepository
	AffectedPackage                              = shared.AffectedPackage
	AffectedSymbol                               = shared.AffectedSymbol
)

func (s *Service) GetVulnerabilitiesByIDs(ctx context.Context, ids ...int) ([]Vulnerability, error) {
	return s.store.GetVulnerabilitiesByIDs(ctx, ids...)
}

func (s *Service) VulnerabilityByID(ctx context.Context, id int) (Vulnerability, bool, error) {
	return s.store.VulnerabilityByID(ctx, id)
}

func (s *Service) VulnerabilityMatchByID(ctx context.Context, id int) (VulnerabilityMatch, bool, error) {
	return s.store.VulnerabilityMatchByID(ctx, id)
}

func (s *Service) GetVulnerabilities(ctx context.Context, args GetVulnerabilitiesArgs) ([]Vulnerability, int, error) {
	return s.store.GetVulnerabilities(ctx, args)
}

func (s *Service) GetVulnerabilityMatches(ctx context.Context, args GetVulnerabilityMatchesArgs) ([]VulnerabilityMatch, int, error) {
	return s.store.GetVulnerabilityMatches(ctx, args)
}

func (s *Service) GetVulnerabilityMatchesSummaryCounts(ctx context.Context) (GetVulnerabilityMatchesSummaryCounts, error) {
	return s.store.GetVulnerabilityMatchesSummaryCount(ctx)
}

func (s *Service) GetVulnerabilityMatchesCountByRepository(ctx context.Context, args GetVulnerabilityMatchesCountByRepositoryArgs) ([]VulnerabilityMatchesByRepository, int, error) {
	return s.store.GetVulnerabilityMatchesCountByRepository(ctx, args)
}

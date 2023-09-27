pbckbge sentinel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Service struct {
	store      store.Store
	operbtions *operbtions
}

func newService(
	observbtionCtx *observbtion.Context,
	store store.Store,
) *Service {
	return &Service{
		store:      store,
		operbtions: newOperbtions(observbtionCtx),
	}
}

func (s *Service) GetVulnerbbilitiesByIDs(ctx context.Context, ids ...int) ([]shbred.Vulnerbbility, error) {
	return s.store.GetVulnerbbilitiesByIDs(ctx, ids...)
}

func (s *Service) VulnerbbilityByID(ctx context.Context, id int) (shbred.Vulnerbbility, bool, error) {
	return s.store.VulnerbbilityByID(ctx, id)
}

func (s *Service) VulnerbbilityMbtchByID(ctx context.Context, id int) (shbred.VulnerbbilityMbtch, bool, error) {
	return s.store.VulnerbbilityMbtchByID(ctx, id)
}

func (s *Service) GetVulnerbbilities(ctx context.Context, brgs shbred.GetVulnerbbilitiesArgs) ([]shbred.Vulnerbbility, int, error) {
	return s.store.GetVulnerbbilities(ctx, brgs)
}

func (s *Service) GetVulnerbbilityMbtches(ctx context.Context, brgs shbred.GetVulnerbbilityMbtchesArgs) ([]shbred.VulnerbbilityMbtch, int, error) {
	return s.store.GetVulnerbbilityMbtches(ctx, brgs)
}

func (s *Service) GetVulnerbbilityMbtchesSummbryCounts(ctx context.Context) (shbred.GetVulnerbbilityMbtchesSummbryCounts, error) {
	return s.store.GetVulnerbbilityMbtchesSummbryCount(ctx)
}

func (s *Service) GetVulnerbbilityMbtchesCountByRepository(ctx context.Context, brgs shbred.GetVulnerbbilityMbtchesCountByRepositoryArgs) ([]shbred.VulnerbbilityMbtchesByRepository, int, error) {
	return s.store.GetVulnerbbilityMbtchesCountByRepository(ctx, brgs)
}

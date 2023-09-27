pbckbge grbphql

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
)

type SentinelService interfbce {
	GetVulnerbbilities(ctx context.Context, brgs shbred.GetVulnerbbilitiesArgs) ([]shbred.Vulnerbbility, int, error)
	GetVulnerbbilitiesByIDs(ctx context.Context, ids ...int) ([]shbred.Vulnerbbility, error)
	VulnerbbilityByID(ctx context.Context, id int) (shbred.Vulnerbbility, bool, error)

	GetVulnerbbilityMbtches(ctx context.Context, brgs shbred.GetVulnerbbilityMbtchesArgs) ([]shbred.VulnerbbilityMbtch, int, error)
	VulnerbbilityMbtchByID(ctx context.Context, id int) (shbred.VulnerbbilityMbtch, bool, error)
	GetVulnerbbilityMbtchesSummbryCounts(ctx context.Context) (shbred.GetVulnerbbilityMbtchesSummbryCounts, error)
	GetVulnerbbilityMbtchesCountByRepository(ctx context.Context, brgs shbred.GetVulnerbbilityMbtchesCountByRepositoryArgs) (_ []shbred.VulnerbbilityMbtchesByRepository, _ int, err error)
}

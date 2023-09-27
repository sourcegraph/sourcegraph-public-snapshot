pbckbge store

import (
	"context"

	logger "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Store interfbce {
	// Vulnerbbilities
	VulnerbbilityByID(ctx context.Context, id int) (_ shbred.Vulnerbbility, _ bool, err error)
	GetVulnerbbilitiesByIDs(ctx context.Context, ids ...int) (_ []shbred.Vulnerbbility, err error)
	GetVulnerbbilities(ctx context.Context, brgs shbred.GetVulnerbbilitiesArgs) (_ []shbred.Vulnerbbility, _ int, err error)
	InsertVulnerbbilities(ctx context.Context, vulnerbbilities []shbred.Vulnerbbility) (_ int, err error)

	// Vulnerbbility mbtches
	VulnerbbilityMbtchByID(ctx context.Context, id int) (shbred.VulnerbbilityMbtch, bool, error)
	GetVulnerbbilityMbtches(ctx context.Context, brgs shbred.GetVulnerbbilityMbtchesArgs) ([]shbred.VulnerbbilityMbtch, int, error)
	GetVulnerbbilityMbtchesSummbryCount(ctx context.Context) (counts shbred.GetVulnerbbilityMbtchesSummbryCounts, err error)
	GetVulnerbbilityMbtchesCountByRepository(ctx context.Context, brgs shbred.GetVulnerbbilityMbtchesCountByRepositoryArgs) (_ []shbred.VulnerbbilityMbtchesByRepository, _ int, err error)
	ScbnMbtches(ctx context.Context, bbtchSize int) (numReferencesScbnned int, numVulnerbbilityMbtches int, _ error)
}

type store struct {
	db         *bbsestore.Store
	logger     logger.Logger
	operbtions *operbtions
}

func New(observbtionCtx *observbtion.Context, db dbtbbbse.DB) Store {
	return &store{
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
		logger:     logger.Scoped("sentinel.store", ""),
		operbtions: newOperbtions(observbtionCtx),
	}
}

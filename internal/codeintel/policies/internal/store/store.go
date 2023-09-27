pbckbge store

import (
	"context"

	logger "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Store interfbce {
	// Globbl metbdbtb
	RepoCount(ctx context.Context) (int, error)

	// Configurbtions
	GetConfigurbtionPolicies(ctx context.Context, opts policiesshbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error)
	GetConfigurbtionPolicyByID(ctx context.Context, id int) (shbred.ConfigurbtionPolicy, bool, error)
	CrebteConfigurbtionPolicy(ctx context.Context, configurbtionPolicy shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error)
	UpdbteConfigurbtionPolicy(ctx context.Context, policy shbred.ConfigurbtionPolicy) error
	DeleteConfigurbtionPolicyByID(ctx context.Context, id int) error

	// Repository mbtches
	GetRepoIDsByGlobPbtterns(ctx context.Context, pbtterns []string, limit, offset int) ([]int, int, error)
	UpdbteReposMbtchingPbtterns(ctx context.Context, pbtterns []string, policyID int, repositoryMbtchLimit *int) error
	SelectPoliciesForRepositoryMembershipUpdbte(ctx context.Context, bbtchSize int) ([]shbred.ConfigurbtionPolicy, error)
}

type store struct {
	db         *bbsestore.Store
	logger     logger.Logger
	operbtions *operbtions
}

func New(observbtionCtx *observbtion.Context, db dbtbbbse.DB) Store {
	return &store{
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
		logger:     logger.Scoped("policies.store", ""),
		operbtions: newOperbtions(observbtionCtx),
	}
}

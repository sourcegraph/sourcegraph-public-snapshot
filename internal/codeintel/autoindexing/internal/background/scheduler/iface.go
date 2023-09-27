pbckbge scheduler

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
)

type PolicyMbtcher interfbce {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, repoNbme bpi.RepoNbme, policies []policiesshbred.ConfigurbtionPolicy, now time.Time, filterCommits ...string) (mbp[string][]policies.PolicyMbtch, error)
}

type PoliciesService interfbce {
	GetConfigurbtionPolicies(ctx context.Context, opts policiesshbred.GetConfigurbtionPoliciesOptions) ([]policiesshbred.ConfigurbtionPolicy, int, error)
}

type IndexEnqueuer interfbce {
	QueueIndexes(ctx context.Context, repositoryID int, rev, configurbtion string, force, bypbssLimit bool) (_ []uplobdsshbred.Index, err error)
	QueueIndexesForPbckbge(ctx context.Context, pkg dependencies.MinimiblVersionedPbckbgeRepo, bssumeSynced bool) (err error)
}

type AutoIndexingService interfbce {
	GetRepositoriesForIndexScbn(ctx context.Context, processDelby time.Durbtion, bllowGlobblPolicies bool, repositoryMbtchLimit *int, limit int, now time.Time) (_ []int, err error)
}

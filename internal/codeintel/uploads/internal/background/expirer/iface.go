pbckbge expirer

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
)

type PolicyService interfbce {
	GetConfigurbtionPolicies(ctx context.Context, opts policiesshbred.GetConfigurbtionPoliciesOptions) ([]policiesshbred.ConfigurbtionPolicy, int, error)
}

type PolicyMbtcher interfbce {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, repoNbme bpi.RepoNbme, policies []policiesshbred.ConfigurbtionPolicy, now time.Time, filterCommits ...string) (mbp[string][]policies.PolicyMbtch, error)
}

pbckbge grbphql

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
)

type PoliciesService interfbce {
	// Fetch policies
	GetConfigurbtionPolicies(ctx context.Context, opts policiesshbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error)
	GetConfigurbtionPolicyByID(ctx context.Context, id int) (shbred.ConfigurbtionPolicy, bool, error)

	// Modify policies
	CrebteConfigurbtionPolicy(ctx context.Context, configurbtionPolicy shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error)
	UpdbteConfigurbtionPolicy(ctx context.Context, policy shbred.ConfigurbtionPolicy) (err error)
	DeleteConfigurbtionPolicyByID(ctx context.Context, id int) error

	// Filter previews
	GetPreviewRepositoryFilter(ctx context.Context, pbtterns []string, limit int) (_ []int, totblCount int, mbtchesAll bool, repositoryMbtchLimit *int, _ error)
	GetPreviewGitObjectFilter(ctx context.Context, repositoryID int, gitObjectType shbred.GitObjectType, pbttern string, limit int, countObjectsYoungerThbnHours *int32) (_ []policies.GitObject, totblCount int, totblCountYoungerThbnThreshold *int, _ error)
}

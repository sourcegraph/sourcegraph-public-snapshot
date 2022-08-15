package expiration

import (
	"context"
	"time"

	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	policyShared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	uploadShared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type PolicyService interface {
	GetConfigurationPolicies(ctx context.Context, opts policyShared.GetConfigurationPoliciesOptions) ([]policyShared.ConfigurationPolicy, int, error)
}

type UploadService interface {
	// Uploads
	GetUploads(ctx context.Context, opts uploadShared.GetUploadsOptions) (uploads []uploadShared.Upload, totalCount int, err error)
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error)

	// Commits
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)

	// Repositories
	SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error)
}

type PolicyMatcher interface {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, policies []dbstore.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

package expiration

import (
	"context"
	"time"

	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
)

type PolicyService interface {
	GetConfigurationPolicies(ctx context.Context, opts types.GetConfigurationPoliciesOptions) ([]types.ConfigurationPolicy, int, error)
}

type UploadService interface {
	// Uploads
	GetUploads(ctx context.Context, opts types.GetUploadsOptions) (uploads []types.Upload, totalCount int, err error)
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error)
	BackfillReferenceCountBatch(ctx context.Context, batchSize int) error

	// Commits
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)

	// Repositories
	SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error)
}

type PolicyMatcher interface {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, policies []types.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

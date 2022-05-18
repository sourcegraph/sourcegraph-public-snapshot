package janitor

import (
	"context"
	"time"

	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type DBStore interface {
	basestore.ShareableStore

	Handle() *basestore.TransactableHandle
	Transact(ctx context.Context) (DBStore, error)
	Done(err error) error

	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	HardDeleteUploadByID(ctx context.Context, ids ...int) error
	GetConfigurationPolicies(ctx context.Context, opts dbstore.GetConfigurationPoliciesOptions) ([]dbstore.ConfigurationPolicy, int, error)
	SelectRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) ([]int, error)
	CommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) ([]string, *string, error)
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) error
	SoftDeleteExpiredUploads(ctx context.Context) (int, error)
	DirtyRepositories(ctx context.Context) (map[int]int, error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (int, error)
	StaleSourcedCommits(ctx context.Context, threshold time.Duration, limit int, now time.Time) ([]dbstore.SourcedCommits, error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (int, int, error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (int, int, int, error)
	SelectPoliciesForRepositoryMembershipUpdate(ctx context.Context, batchSize int) (configurationPolicies []dbstore.ConfigurationPolicy, err error)
	UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) (err error)
	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (int, error)
}

type DBStoreShim struct {
	*dbstore.Store
}

func (s *DBStoreShim) Transact(ctx context.Context) (DBStore, error) {
	store, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &DBStoreShim{store}, nil
}

type LSIFStore interface {
	Transact(ctx context.Context) (LSIFStore, error)
	Done(err error) error

	Clear(ctx context.Context, bundleIDs ...int) error
}

type LSIFStoreShim struct {
	*lsifstore.Store
}

func (s *LSIFStoreShim) Transact(ctx context.Context) (LSIFStore, error) {
	store, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &LSIFStoreShim{store}, nil
}

type PolicyMatcher interface {
	CommitsDescribedByPolicy(ctx context.Context, repositoryID int, policies []dbstore.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

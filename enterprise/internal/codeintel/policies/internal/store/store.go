package store

import (
	"context"

	logger "github.com/sourcegraph/log"

	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Store provides the interface for policies storage.
type Store interface {
	// Configurations
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]types.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (types.ConfigurationPolicy, bool, error)
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy types.ConfigurationPolicy) (types.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy types.ConfigurationPolicy) error
	DeleteConfigurationPolicyByID(ctx context.Context, id int) error

	// Repositories
	RepoCount(ctx context.Context) (int, error)
	GetRepoIDsByGlobPatterns(ctx context.Context, patterns []string, limit, offset int) ([]int, int, error)
	UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) error

	// Utilities
	SelectPoliciesForRepositoryMembershipUpdate(ctx context.Context, batchSize int) ([]types.ConfigurationPolicy, error)
}

// store manages the policies store.
type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new policies store.
func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("policies.store", ""),
		operations: newOperations(observationCtx),
	}
}

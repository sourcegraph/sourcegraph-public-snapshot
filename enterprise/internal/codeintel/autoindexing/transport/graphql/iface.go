package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type AutoIndexingService interface {
	// Inference configuration
	GetInferenceScript(ctx context.Context) (string, error)
	SetInferenceScript(ctx context.Context, script string) error

	// Repository configuration
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (shared.IndexConfiguration, bool, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) error

	// Inference
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force bool, bypassLimit bool) ([]types.Index, error)
	InferIndexConfiguration(ctx context.Context, repositoryID int, commit string, localOverrideScript string, bypassLimit bool) (*config.IndexConfiguration, []config.IndexJobHint, error)
	InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, localOverrideScript string, bypassLimit bool) ([]config.IndexJob, error)
}

type (
	UploadsService = uploadsgraphql.UploadsService
	PolicyService  = uploadsgraphql.PolicyService
)

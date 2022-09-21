package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	policy "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type UploadsService interface {
	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []shared.UploadsWithRepositoryNamespace, err error)
	GetUploads(ctx context.Context, opts types.GetUploadsOptions) (uploads []types.Upload, totalCount int, err error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []types.UploadLog, err error)
	GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
	GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) ([]string, int, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error)
}

type PolicyResolver interface {
	PolicyResolverFactory(ctx context.Context) (_ policy.PolicyResolver, err error)
}

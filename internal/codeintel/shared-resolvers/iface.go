package shared

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	autoindex "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	policy "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type AutoIndexingService interface {
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []autoindex.Index, err error)
	GetUnsafeDB() database.DB
	GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
}

type UploadsService interface {
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []types.UploadLog, err error)
	GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
	GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) ([]string, int, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error)
}

type PolicyResolver interface {
	PolicyResolverFactory(ctx context.Context) (_ policy.PolicyResolver, err error)
}

type CachedLocationResolverz interface {
	Repository(ctx context.Context, id api.RepoID) (*gql.RepositoryResolver, error)
}

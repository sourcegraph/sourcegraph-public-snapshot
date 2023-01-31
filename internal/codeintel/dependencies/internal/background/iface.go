package background

import (
	"context"
	"io"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type GitserverClient interface {
	ArchiveReader(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, options gitserver.ArchiveOptions) (io.ReadCloser, error)
	RequestRepoUpdate(context.Context, api.RepoName, time.Duration) (*protocol.RepoUpdateResponse, error)
}

type ExternalServiceStore interface {
	List(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
	Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error)
}

type DependenciesService interface {
	InsertPackageRepoRefs(ctx context.Context, deps []shared.MinimalPackageRepoRef) (_ []shared.PackageRepoReference, _ []shared.PackageRepoRefVersion, err error)
}

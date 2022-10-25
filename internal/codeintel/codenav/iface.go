package codenav

import (
	"context"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type UploadService interface {
	GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []types.Dump, err error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) (ids []int, recordsScanned int, totalCount int, err error)
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []types.Dump, err error)
	InferClosestUploads(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []types.Dump, err error)
}

type GitserverClient interface {
	CommitsExist(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error)
	DiffPath(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, sourceCommit, targetCommit, path string) ([]*diff.Hunk, error)
}

type DBStore interface {
	RepoName(ctx context.Context, repositoryID int) (string, error)
	RepoNames(ctx context.Context, repositoryIDs ...int) (map[int]string, error)
}

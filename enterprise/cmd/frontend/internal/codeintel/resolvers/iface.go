package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/api"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

type CodeIntelAPI interface {
	FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) ([]store.Dump, error)
	Ranges(ctx context.Context, file string, startLine, endLine, uploadID int) ([]api.ResolvedCodeIntelligenceRange, error)
	Definitions(ctx context.Context, file string, line, character, uploadID int) ([]api.ResolvedLocation, error)
	References(ctx context.Context, repositoryID int, commit string, limit int, cursor api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error)
	Hover(ctx context.Context, file string, line, character, uploadID int) (string, lsifstore.Range, bool, error)
	Diagnostics(ctx context.Context, prefix string, uploadID, limit, offset int) ([]api.ResolvedDiagnostic, int, error)
}

type DBStore interface {
	gitserver.DBStore

	GetUploadByID(ctx context.Context, id int) (dbstore.Upload, bool, error)
	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	DeleteUploadByID(ctx context.Context, id int) (bool, error)
	GetDumpByID(ctx context.Context, id int) (dbstore.Dump, bool, error)
	FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) ([]dbstore.Dump, error)
	FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, graph *gitserver.CommitGraph) ([]dbstore.Dump, error)
	GetPackage(ctx context.Context, scheme, name, version string) (dbstore.Dump, bool, error)
	SameRepoPager(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (int, api.ReferencePager, error)
	PackageReferencePager(ctx context.Context, scheme, name, version string, repositoryID, limit int) (int, api.ReferencePager, error)
	HasRepository(ctx context.Context, repositoryID int) (bool, error)
	HasCommit(ctx context.Context, repositoryID int, commit string) (bool, error)
	MarkRepositoryAsDirty(ctx context.Context, repositoryID int) error
	GetIndexByID(ctx context.Context, id int) (dbstore.Index, bool, error)
	GetIndexes(ctx context.Context, opts dbstore.GetIndexesOptions) ([]dbstore.Index, int, error)
	DeleteIndexByID(ctx context.Context, id int) (bool, error)
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (store.IndexConfiguration, bool, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) error
}

type DBStoreShim struct {
	*dbstore.Store
}

func (s *DBStoreShim) SameRepoPager(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (int, api.ReferencePager, error) {
	count, pager, err := s.Store.SameRepoPager(ctx, repositoryID, commit, scheme, name, version, limit)
	if err != nil {
		return 0, nil, err
	}

	return count, pager, nil
}

func (s *DBStoreShim) PackageReferencePager(ctx context.Context, scheme, name, version string, repositoryID, limit int) (int, api.ReferencePager, error) {
	count, pager, err := s.Store.PackageReferencePager(ctx, scheme, name, version, repositoryID, limit)
	if err != nil {
		return 0, nil, err
	}

	return count, pager, nil
}

type LSIFStore interface {
	api.LSIFStore
}

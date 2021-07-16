package indexing

import (
	"context"
	"regexp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/schema"
)

type DBStore interface {
	With(other basestore.ShareableStore) DBStore
	GetRepositoriesWithIndexConfiguration(ctx context.Context) ([]int, error)
	GetAutoindexDisabledRepositories(ctx context.Context) ([]int, error)
	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	GetUploadByID(ctx context.Context, id int) (dbstore.Upload, bool, error)
	ReferencesForUpload(ctx context.Context, uploadID int) (dbstore.PackageReferenceScanner, error)
	InsertCloneableDependencyRepo(ctx context.Context, dependency precise.Package) (bool, error)
}

type DBStoreShim struct {
	*dbstore.Store
}

type IndexingSettingStore interface {
	GetLastestSchemaSettings(context.Context, api.SettingsSubject) (*schema.Settings, error)
}

type IndexingRepoStore interface {
	ListRepoNames(ctx context.Context, opt database.ReposListOptions) (results []types.RepoName, err error)
	ListIndexableRepos(ctx context.Context, opts database.ListIndexableReposOptions) (results []types.RepoName, err error)
}

func (s *DBStoreShim) With(other basestore.ShareableStore) DBStore {
	return &DBStoreShim{s.Store.With(s)}
}

type ExternalServiceStore interface {
	List(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
	Upsert(ctx context.Context, svcs ...*types.ExternalService) (err error)
}

type GitserverClient interface {
	Head(ctx context.Context, repositoryID int) (string, bool, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
}

type IndexEnqueuer interface {
	QueueIndexesForRepository(ctx context.Context, repositoryID int) error
	QueueIndexesForPackage(ctx context.Context, pkg precise.Package) error
}

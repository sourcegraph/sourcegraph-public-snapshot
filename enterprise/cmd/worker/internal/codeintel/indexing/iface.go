package indexing

import (
	"context"
	"regexp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
	"github.com/sourcegraph/sourcegraph/schema"
)

type DBStore interface {
	With(other basestore.ShareableStore) DBStore
	GetRepositoriesWithIndexConfiguration(ctx context.Context) ([]int, error)
	GetAutoindexDisabledRepositories(ctx context.Context) ([]int, error)
	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	GetUploadByID(ctx context.Context, id int) (dbstore.Upload, bool, error)
	ReferencesForUpload(ctx context.Context, uploadID int) (dbstore.PackageReferenceScanner, error)
}

type DBStoreShim struct {
	*dbstore.Store
}

type IndexingSettingStore interface {
	GetLastestSchemaSettings(context.Context, api.SettingsSubject) (*schema.Settings, error)
}

type IndexingRepoStore interface {
	ListRepoNames(ctx context.Context, opt database.ReposListOptions) (results []types.RepoName, err error)
}

func (s *DBStoreShim) With(other basestore.ShareableStore) DBStore {
	return &DBStoreShim{s.Store.With(s)}
}

type GitserverClient interface {
	Head(ctx context.Context, repositoryID int) (string, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
}

type gitClient struct {
	client       GitserverClient
	repositoryID int
	commit       string
}

func newGitClient(client GitserverClient, repositoryID int, commit string) gitClient {
	return gitClient{
		client:       client,
		repositoryID: repositoryID,
		commit:       commit,
	}
}

func (c gitClient) ListFiles(ctx context.Context, pattern *regexp.Regexp) ([]string, error) {
	return c.client.ListFiles(ctx, c.repositoryID, c.commit, pattern)
}

func (c gitClient) FileExists(ctx context.Context, file string) (bool, error) {
	return c.client.FileExists(ctx, c.repositoryID, c.commit, file)
}

func (c gitClient) RawContents(ctx context.Context, file string) ([]byte, error) {
	return c.client.RawContents(ctx, c.repositoryID, c.commit, file)
}

type IndexEnqueuer interface {
	QueueIndexesForRepository(ctx context.Context, repositoryID int) error
	QueueIndexesForPackage(ctx context.Context, pkg semantic.Package) error
}

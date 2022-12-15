package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type SubRepoPermsStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) SubRepoPermsStore
	Transact(ctx context.Context) (SubRepoPermsStore, error)
	Done(err error) error
	Upsert(ctx context.Context, userID int32, repoID api.RepoID, perms authz.SubRepoPermissions) error
	UpsertWithSpec(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissions) error
	Get(ctx context.Context, userID int32, repoID api.RepoID) (*authz.SubRepoPermissions, error)
	GetByUser(ctx context.Context, userID int32) (map[api.RepoName]authz.SubRepoPermissions, error)
	// GetByUserAndService gets the sub repo permissions for a user, but filters down
	// to only repos that come from a specific external service.
	GetByUserAndService(ctx context.Context, userID int32, serviceType string, serviceID string) (map[api.ExternalRepoSpec]authz.SubRepoPermissions, error)
	RepoIDSupported(ctx context.Context, repoID api.RepoID) (bool, error)
	RepoSupported(ctx context.Context, repo api.RepoName) (bool, error)
	DeleteByUser(ctx context.Context, userID int32) error
}

// subRepoPermsStore is a no-op placeholder for the OSS version.
type subRepoPermsStore struct{}

var SubRepoPermsWith = func(other basestore.ShareableStore) SubRepoPermsStore {
	return &subRepoPermsStore{}
}

func (s *subRepoPermsStore) With(other basestore.ShareableStore) SubRepoPermsStore {
	return &subRepoPermsStore{}
}

func (s *subRepoPermsStore) Transact(ctx context.Context) (SubRepoPermsStore, error) {
	return &subRepoPermsStore{}, nil
}

func (s *subRepoPermsStore) Done(err error) error {
	return nil
}

func (s *subRepoPermsStore) Upsert(ctx context.Context, userID int32, repoID api.RepoID, perms authz.SubRepoPermissions) error {
	return nil
}

func (s *subRepoPermsStore) UpsertWithSpec(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissions) error {
	return nil
}

func (s *subRepoPermsStore) Get(ctx context.Context, userID int32, repoID api.RepoID) (*authz.SubRepoPermissions, error) {
	return nil, nil
}

func (s *subRepoPermsStore) GetByUser(ctx context.Context, userID int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
	return nil, nil
}

func (s *subRepoPermsStore) GetByUserAndService(ctx context.Context, userID int32, serviceType string, serviceID string) (map[api.ExternalRepoSpec]authz.SubRepoPermissions, error) {
	return nil, nil
}

func (s *subRepoPermsStore) RepoIDSupported(ctx context.Context, repoID api.RepoID) (bool, error) {
	return false, nil
}

func (s *subRepoPermsStore) RepoSupported(ctx context.Context, repo api.RepoName) (bool, error) {
	return false, nil
}

// DeleteByUser deletes all rows associated with the given user
func (s *subRepoPermsStore) DeleteByUser(ctx context.Context, userID int32) error {
	q := sqlf.Sprintf(`
DELETE FROM sub_repo_permissions WHERE user_id = %d
`, userID)
	return s.Exec(ctx, q)
}

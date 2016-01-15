package auth

import (
	"fmt"
	"os"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/platform/storage"
)

const permsBucket = "permissions"

type TokenNotFoundError struct {
	msg string
}

func (e TokenNotFoundError) Error() string { return e.msg }

// RepoPermsStore contains methods for storing user permissions
// for accessing private repos.
type RepoPermsStore struct{}

type RepoPermissions struct {
	Allow bool
}

func (s *RepoPermsStore) storage(ctx context.Context) storage.System {
	return storage.Namespace(ctx, "core.repo-perms", "")
}

func (s *RepoPermsStore) getKey(uid int, repo string) string {
	return fmt.Sprintf("%d:%s", uid, repo)
}

func (s *RepoPermsStore) Set(ctx context.Context, uid int, repo string) error {
	fs := s.storage(ctx)
	return storage.PutJSON(fs, permsBucket, s.getKey(uid, repo), RepoPermissions{Allow: true})
}

func (s *RepoPermsStore) Get(ctx context.Context, uid int, repo string) (RepoPermissions, error) {
	perms := RepoPermissions{}
	fs := s.storage(ctx)
	err := storage.GetJSON(fs, permsBucket, s.getKey(uid, repo), &perms)
	if err != nil {
		if os.IsNotExist(err) {
			return RepoPermissions{}, nil
		}
		return RepoPermissions{}, err
	}

	return perms, nil
}

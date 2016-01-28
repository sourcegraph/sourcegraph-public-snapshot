package store

import (
	"errors"

	"golang.org/x/net/context"
)

// RepoPerms defines the interface for getting and setting permissions
// for access to private repos on this server.
type RepoPerms interface {
	// Add adds permissions for the user to access the repo.
	Add(ctx context.Context, uid int32, repo string) error

	// Update updates the list of repos visible to the user to the
	// given slice of repo URIs. Repos that user previously had
	// access to but are not present in the given slice, are removed.
	Update(ctx context.Context, uid int32, repos []string) error

	// Delete removes permissions for a user to access a repo.
	Delete(ctx context.Context, uid int32, repo string) error

	// ListUserRepos list the repos that the user has access to.
	ListUserRepos(ctx context.Context, uid int32) ([]string, error)

	// ListRepoUsers list all users that have access to the repo.
	ListRepoUsers(ctx context.Context, repo string) ([]int32, error)

	// DeleteUser deletes all permissions records pertaining to the user.
	DeleteUser(ctx context.Context, uid int32) error

	// DeleteRepo deletes all permissions records pertaining to the repo.
	DeleteRepo(ctx context.Context, repo string) error
}

var (
	// ErrRepoPermissionExists occurs when a repo permission is already granted
	// to a user.
	ErrRepoPermissionExists = errors.New("user already has access to the repo")
)

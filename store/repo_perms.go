package store

import (
	"errors"

	"golang.org/x/net/context"
)

// RepoPerms defines the interface for getting and setting permissions
// for access to private repos on this server.
type RepoPerms interface {
	// ListRepoUsers list all users that have access to the repo.
	ListRepoUsers(ctx context.Context, repo string) ([]int32, error)
}

var (
	// ErrRepoPermissionExists occurs when a repo permission is already granted
	// to a user.
	ErrRepoPermissionExists = errors.New("user already has access to the repo")
)

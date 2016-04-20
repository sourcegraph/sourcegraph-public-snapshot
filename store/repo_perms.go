package store

import "golang.org/x/net/context"

// RepoPerms defines the interface for getting and setting permissions
// for access to private repos on this server.
type RepoPerms interface {
	// ListRepoUsers list all users that have access to the repo.
	ListRepoUsers(ctx context.Context, repo string) ([]int32, error)
}

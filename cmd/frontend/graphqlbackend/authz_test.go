package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// setupGlobalSubRepoDeny replaces the global constructor for sub repo permissions client
// for testing purposes.
//
// Where possible, prefer to use a local mock instead.
func setupGlobalSubRepoDeny(ctx context.Context, t *testing.T, denyPaths []string) context.Context {
	oldFn := subRepoPermsClient
	t.Cleanup(func() { subRepoPermsClient = oldFn })

	subRepoPermsClient = func(db dbutil.DB) authz.SubRepoPermissionChecker {
		m := authz.NewMockSubRepoPermissionChecker()
		m.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
			for _, p := range denyPaths {
				if p == content.Path {
					return authz.None, nil
				}
			}
			return authz.Read, nil
		})
		m.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		return m
	}

	return actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})
}

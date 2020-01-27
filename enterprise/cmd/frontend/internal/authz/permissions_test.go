package authz

import (
	"reflect"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func bitmap(ids ...uint32) *roaring.Bitmap {
	bm := roaring.NewBitmap()
	bm.AddMany(ids)
	return bm
}

func TestUserPermissions_AuthorizedRepos(t *testing.T) {
	tests := []struct {
		name        string
		repos       []*types.Repo
		p           *UserPermissions
		expectPerms []authz.RepoPerms
	}{
		{
			name:  "wrong permissions type",
			repos: []*types.Repo{},
			p: &UserPermissions{
				Perm:     authz.Read,
				Type:     "",
				IDs:      bitmap(),
				Provider: authz.ProviderSourcegraph,
			},
			expectPerms: []authz.RepoPerms{},
		},
		{
			name:  "nil bitmap",
			repos: []*types.Repo{},
			p: &UserPermissions{
				Perm:     authz.Read,
				Type:     authz.PermRepos,
				IDs:      nil,
				Provider: authz.ProviderSourcegraph,
			},
			expectPerms: []authz.RepoPerms{},
		},
		{
			name:  "empty bitmap",
			repos: []*types.Repo{},
			p: &UserPermissions{
				Perm:     authz.Read,
				Type:     authz.PermRepos,
				IDs:      bitmap(),
				Provider: authz.ProviderSourcegraph,
			},
			expectPerms: []authz.RepoPerms{},
		},

		{
			name: "filter out repos have id=0",
			repos: []*types.Repo{
				{ID: 0},
				{ID: 1},
			},
			p: &UserPermissions{
				Perm:     authz.Read,
				Type:     authz.PermRepos,
				IDs:      bitmap(1),
				Provider: authz.ProviderSourcegraph,
			},
			expectPerms: []authz.RepoPerms{
				{Repo: &types.Repo{ID: 1}, Perms: authz.Read},
			},
		},
		{
			name: "candidate list is a subset",
			repos: []*types.Repo{
				{ID: 1},
				{ID: 2},
			},
			p: &UserPermissions{
				Perm:     authz.Read,
				Type:     authz.PermRepos,
				IDs:      bitmap(1, 2, 3, 4),
				Provider: authz.ProviderSourcegraph,
			},
			expectPerms: []authz.RepoPerms{
				{Repo: &types.Repo{ID: 1}, Perms: authz.Read},
				{Repo: &types.Repo{ID: 2}, Perms: authz.Read},
			},
		},
		{
			name: "candidate list is a superset",
			repos: []*types.Repo{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
			p: &UserPermissions{
				Perm:     authz.Read,
				Type:     authz.PermRepos,
				IDs:      bitmap(1, 2),
				Provider: authz.ProviderSourcegraph,
			},
			expectPerms: []authz.RepoPerms{
				{Repo: &types.Repo{ID: 1}, Perms: authz.Read},
				{Repo: &types.Repo{ID: 2}, Perms: authz.Read},
			},
		},
		{
			name: "candidate list has intersection",
			repos: []*types.Repo{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
			p: &UserPermissions{
				Perm:     authz.Read,
				Type:     authz.PermRepos,
				IDs:      bitmap(1, 3, 5, 7),
				Provider: authz.ProviderSourcegraph,
			},
			expectPerms: []authz.RepoPerms{
				{Repo: &types.Repo{ID: 1}, Perms: authz.Read},
				{Repo: &types.Repo{ID: 3}, Perms: authz.Read},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			perms := test.p.AuthorizedRepos(test.repos)
			if !reflect.DeepEqual(test.expectPerms, perms) {
				t.Fatalf("perms: %s", cmp.Diff(test.expectPerms, perms))
			}
		})
	}
}

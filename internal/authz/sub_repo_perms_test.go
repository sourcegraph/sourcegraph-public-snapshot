package authz

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSubRepoPermsPermissions(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableSubRepoPermissions: true,
			},
		},
	})
	t.Cleanup(func() { conf.Mock(nil) })

	testCases := []struct {
		name     string
		userID   int32
		content  RepoContent
		clientFn func() *SubRepoPermsClient
		want     Perms
	}{
		{
			name:   "Empty path",
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "",
			},
			clientFn: func() *SubRepoPermsClient {
				return NewSubRepoPermsClient(NewMockSubRepoPermissionsGetter())
			},
			want: Read,
		},
		{
			name:   "No rules",
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "/dev/thing",
			},
			clientFn: func() *SubRepoPermsClient {
				getter := NewMockSubRepoPermissionsGetter()
				getter.GetByUserFunc.SetDefaultHook(func(ctx context.Context, i int32) (map[api.RepoName]SubRepoPermissions, error) {
					return map[api.RepoName]SubRepoPermissions{
						"sample": {
							PathIncludes: []string{},
							PathExcludes: []string{},
						},
					}, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			want: None,
		},
		{
			name:   "Exclude",
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "/dev/thing",
			},
			clientFn: func() *SubRepoPermsClient {
				getter := NewMockSubRepoPermissionsGetter()
				getter.GetByUserFunc.SetDefaultHook(func(ctx context.Context, i int32) (map[api.RepoName]SubRepoPermissions, error) {
					return map[api.RepoName]SubRepoPermissions{
						"sample": {
							PathIncludes: []string{},
							PathExcludes: []string{"/dev/*"},
						},
					}, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			want: None,
		},
		{
			name:   "Include",
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "/dev/thing",
			},
			clientFn: func() *SubRepoPermsClient {
				getter := NewMockSubRepoPermissionsGetter()
				getter.GetByUserFunc.SetDefaultHook(func(ctx context.Context, i int32) (map[api.RepoName]SubRepoPermissions, error) {
					return map[api.RepoName]SubRepoPermissions{
						"sample": {
							PathIncludes: []string{"*"},
						},
					}, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			want: None,
		},
		{
			name:   "Exclude takes precedence",
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "/dev/thing",
			},
			clientFn: func() *SubRepoPermsClient {
				getter := NewMockSubRepoPermissionsGetter()
				getter.GetByUserFunc.SetDefaultHook(func(ctx context.Context, i int32) (map[api.RepoName]SubRepoPermissions, error) {
					return map[api.RepoName]SubRepoPermissions{
						"sample": {
							PathIncludes: []string{"*"},
							PathExcludes: []string{"/dev/*"},
						},
					}, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			want: None,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			have, err := tc.clientFn().Permissions(context.Background(), tc.userID, tc.content)
			if err != nil {
				t.Fatal(err)
			}
			if have != tc.want {
				t.Fatalf("have %v, want %v", have, tc.want)
			}
		})
	}
}

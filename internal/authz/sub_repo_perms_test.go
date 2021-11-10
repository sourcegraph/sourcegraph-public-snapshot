package authz

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSubRepoPermsPermissions(t *testing.T) {
	baseGetter := func() *MockSubRepoPermissionsGetter {
		getter := NewMockSubRepoPermissionsGetter()
		getter.RepoSupportedFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (bool, error) {
			return true, nil
		})
		return getter
	}
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
		ctx      context.Context
		userID   int32
		content  RepoContent
		clientFn func() *subRepoPermsClient
		want     Perms
	}{
		{
			name:   "Not supported",
			ctx:    context.Background(),
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "",
			},
			clientFn: func() *subRepoPermsClient {
				getter := NewMockSubRepoPermissionsGetter()
				getter.RepoSupportedFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (bool, error) {
					return false, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			want: Read,
		},
		{
			name:   "Empty path",
			ctx:    context.Background(),
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "",
			},
			clientFn: func() *subRepoPermsClient {
				getter := baseGetter()
				return NewSubRepoPermsClient(getter)
			},
			want: Read,
		},
		{
			name:   "No rules",
			ctx:    context.Background(),
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "/dev/thing",
			},
			clientFn: func() *subRepoPermsClient {
				getter := baseGetter()
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
			ctx:    context.Background(),
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "/dev/thing",
			},
			clientFn: func() *subRepoPermsClient {
				getter := baseGetter()
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
			ctx:    context.Background(),
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "/dev/thing",
			},
			clientFn: func() *subRepoPermsClient {
				getter := baseGetter()
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
			ctx:    context.Background(),
			userID: 1,
			content: RepoContent{
				Repo: "sample",
				Path: "/dev/thing",
			},
			clientFn: func() *subRepoPermsClient {
				getter := baseGetter()
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
			have, err := tc.clientFn().Permissions(tc.ctx, tc.userID, tc.content)
			if err != nil {
				t.Fatal(err)
			}
			if have != tc.want {
				t.Fatalf("have %v, want %v", have, tc.want)
			}
		})
	}
}

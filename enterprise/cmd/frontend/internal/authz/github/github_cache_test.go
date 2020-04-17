package github

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// TestProvider_RepoPerms_cacheTTL tests that cache entries are invalidated after the cache TTL
// changes.
func TestProvider_RepoPerms_cacheTTL(t *testing.T) {
	cacheMisses := 0
	mockClient := &mockClient{
		MockGetRepositoriesByNodeIDFromAPI: func(ctx context.Context, nodeIDs []string) (map[string]*github.Repository, error) {
			cacheMisses++
			return map[string]*github.Repository{
				"u1/public": {},
				"u0/r0":     {},
			}, nil
		},
	}
	mockClient.MockWithToken = func(token string) client {
		return mockClient
	}

	p := NewProvider(mustURL(t, "https://github.com"), "base-token", 3*time.Hour, make(authz.MockCache))
	p.client = mockClient

	ctx := context.Background()

	userAccount := ua("u0", "t0")
	repos := []*types.Repo{
		rp("r0", "u0/r0", "https://github.com/"),
		rp("r1", "u1/r1", "https://github.com/"),
		rp("r2", "u1/public", "https://github.com/"),
	}
	wantPerms := []authz.RepoPerms{
		{Repo: repos[0], Perms: authz.Read},
		{Repo: repos[1], Perms: authz.None},
		{Repo: repos[2], Perms: authz.Read},
	}
	{
		perms, err := p.RepoPerms(ctx, userAccount, repos)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(wantPerms, perms); diff != "" {
			t.Fatal(diff)
		}
		if cacheMisses != 1 {
			t.Errorf("expected 1 cache misses, but got %d", cacheMisses)
		}
		cacheMisses = 0
	}
	{
		perms, gotErr := p.RepoPerms(ctx, userAccount, repos)
		if gotErr != nil {
			t.Fatal(gotErr)
		}
		if diff := cmp.Diff(wantPerms, perms); diff != "" {
			t.Fatal(diff)
		}
		if cacheMisses != 0 {
			t.Errorf("expected 0 cache misses, but got %d", cacheMisses)
		}
		cacheMisses = 0
	}

	p.cacheTTL = 1 * time.Hour // lower cache TTL
	{
		perms, gotErr := p.RepoPerms(ctx, userAccount, repos)
		if gotErr != nil {
			t.Fatal(gotErr)
		}
		if diff := cmp.Diff(wantPerms, perms); diff != "" {
			t.Fatal(diff)
		}
		if cacheMisses != 1 {
			t.Errorf("expected 1 cache misses, but got %d", cacheMisses)
		}
		cacheMisses = 0
	}
}

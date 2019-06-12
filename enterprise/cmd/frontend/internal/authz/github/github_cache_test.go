package github

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

// TestProvider_RepoPerms_cacheTTL tests that cache entries are invalidated after the cache TTL
// changes.
func TestProvider_RepoPerms_cacheTTL(t *testing.T) {
	githubMock := newMockGitHub([]*github.Repository{
		{ID: "u0/r0", IsPrivate: true},
		{ID: "u1/r1", IsPrivate: true},
		{ID: "u1/public"},
	}, map[string][]string{
		"t0": {"u0/r0"},
		"t1": {"u1/r1"},
	})
	github.GetRepositoryByNodeIDMock = githubMock.GetRepositoryByNodeID
	defer func() { github.GetRepositoryByNodeIDMock = nil }()
	github.GetRepositoriesByNodeIDFromAPIMock = githubMock.GetRepositoriesByNodeIDFromAPI
	defer func() { github.GetRepositoriesByNodeIDFromAPIMock = nil }()

	provider := NewProvider(mustURL(t, "https://github.com"), "base-token", 3*time.Hour, make(authz.MockCache))
	ctx := context.Background()

	githubMock.getRepositoriesByNodeIDCount = 0
	userAccount := ua("u0", "t0")
	repos := map[authz.Repo]struct{}{
		rp("r0", "u0/r0", "https://github.com/"):     {},
		rp("r1", "u1/r1", "https://github.com/"):     {},
		rp("r2", "u1/public", "https://github.com/"): {},
	}
	wantPerms := map[api.RepoName]map[authz.Perm]bool{
		"r0": readPerms,
		"r1": noPerms,
		"r2": readPerms,
	}
	{
		gotPerms, gotErr := provider.RepoPerms(ctx, userAccount, repos)
		if gotErr != nil {
			t.Fatal(gotErr)
		}
		if !reflect.DeepEqual(gotPerms, wantPerms) {
			dmp := diffmatchpatch.New()
			t.Errorf("wantPerms != gotPerms:\n%s",
				dmp.DiffPrettyText(dmp.DiffMain(spew.Sdump(wantPerms), spew.Sdump(gotPerms), false)))
		}
		if want, got := 1, githubMock.getRepositoriesByNodeIDCount; want != got {
			t.Errorf("expected %d cache misses, but got %d", want, got)
		}
		githubMock.getRepositoriesByNodeIDCount = 0
	}
	{
		gotPerms, gotErr := provider.RepoPerms(ctx, userAccount, repos)
		if gotErr != nil {
			t.Fatal(gotErr)
		}
		if !reflect.DeepEqual(gotPerms, wantPerms) {
			dmp := diffmatchpatch.New()
			t.Errorf("wantPerms != gotPerms:\n%s",
				dmp.DiffPrettyText(dmp.DiffMain(spew.Sdump(wantPerms), spew.Sdump(gotPerms), false)))
		}
		if want, got := 0, githubMock.getRepositoriesByNodeIDCount; want != got {
			t.Errorf("expected %d cache misses, but got %d", want, got)
		}
		githubMock.getRepositoriesByNodeIDCount = 0
	}

	provider.cacheTTL = 1 * time.Hour // lower cache TTL
	{
		gotPerms, gotErr := provider.RepoPerms(ctx, userAccount, repos)
		if gotErr != nil {
			t.Fatal(gotErr)
		}
		if !reflect.DeepEqual(gotPerms, wantPerms) {
			dmp := diffmatchpatch.New()
			t.Errorf("wantPerms != gotPerms:\n%s",
				dmp.DiffPrettyText(dmp.DiffMain(spew.Sdump(wantPerms), spew.Sdump(gotPerms), false)))
		}
		if want, got := 1, githubMock.getRepositoriesByNodeIDCount; want != got {
			t.Errorf("expected %d cache misses, but got %d", want, got)
		}
		githubMock.getRepositoriesByNodeIDCount = 0
	}
}

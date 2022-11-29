package repos

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGiteaSource_ListRepos(t *testing.T) {
	// This instance has a big warning that repos can be deleted and is only
	// for demo purposes. However, there are 3 year old repos on it and we
	// record the tests. If the tests break you can visit
	// https://try.gitea.io/api/swagger#/repository/repoSearch and create new
	// searches.
	conf := &schema.GiteaConnection{
		Url: "https://try.gitea.io",
		SearchQuery: []string{
			// Random user with 2 repos
			"?uid=3&exclusive=true",
			// Finds a 1 random repo
			"?q=random%20block&includeDesc=true",
		},
	}
	cf, save := newClientFactory(t, t.Name())
	defer save(t)

	svc := &types.ExternalService{
		Kind:   extsvc.KindGitea,
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, conf)),
	}

	ctx := context.Background()
	src, err := NewGiteaSource(ctx, svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	repos, err := listAll(context.Background(), src)
	if err != nil {
		t.Fatal(err)
	}

	sort.Sort(types.Repos(repos))

	testutil.AssertGolden(t, "testdata/sources/"+t.Name(), update(t.Name()), repos)
}

package repos

import (
	"context"
	"net/url"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/assert"
)

func TestGerritSource_ListRepos(t *testing.T) {
	t.Run("no filtering", func(t *testing.T) {
		conf := &schema.GerritConnection{
			Url: "https://gerrit-review.googlesource.com",
		}
		cf, save := newClientFactory(t, t.Name(), httpcli.GerritUnauthenticateMiddleware)
		defer save(t)

		svc := &types.ExternalService{
			Kind:   extsvc.KindGerrit,
			Config: extsvc.NewUnencryptedConfig(marshalJSON(t, conf)),
		}

		ctx := context.Background()
		src, err := NewGerritSource(ctx, svc, cf)
		if err != nil {
			t.Fatal(err)
		}

		src.perPage = 25

		repos, err := listAll(context.Background(), src)
		if err != nil {
			t.Fatal(err)
		}

		testutil.AssertGolden(t, "testdata/sources/GERRIT/"+t.Name(), update(t.Name()), repos)
	})

	t.Run("with filtering", func(t *testing.T) {
		conf := &schema.GerritConnection{
			Url: "https://gerrit-review.googlesource.com",
			Repos: []string{
				"apps/reviewit",
				"buck",
			},
		}
		cf, save := newClientFactory(t, t.Name(), httpcli.GerritUnauthenticateMiddleware)
		defer save(t)

		svc := &types.ExternalService{
			Kind:   extsvc.KindGerrit,
			Config: extsvc.NewUnencryptedConfig(marshalJSON(t, conf)),
		}

		ctx := context.Background()
		src, err := NewGerritSource(ctx, svc, cf)
		if err != nil {
			t.Fatal(err)
		}

		src.perPage = 25

		repos, err := listAll(context.Background(), src)
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, repos, 2)
		repoNames := make([]string, 0, len(repos))
		for _, repo := range repos {
			repoNames = append(repoNames, repo.ExternalRepo.ID)
		}
		assert.ElementsMatch(t, repoNames, []string{
			url.PathEscape("apps/reviewit"),
			"buck",
		})
	})
}

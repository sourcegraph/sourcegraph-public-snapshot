package repos

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGerritSource_ListRepos(t *testing.T) {
	ratelimit.SetupForTest(t)

	t.Run("no filtering", func(t *testing.T) {
		cf, save := NewClientFactory(t, t.Name())
		defer save(t)

		svc := typestest.MakeExternalService(t, extsvc.VariantGerrit, &schema.GerritConnection{
			Url:      "https://gerrit.sgdev.org",
			Username: os.Getenv("GERRIT_USERNAME"),
			Password: os.Getenv("GERRIT_PASSWORD"),
		})

		ctx := context.Background()
		src, err := NewGerritSource(ctx, svc, cf)
		require.NoError(t, err)

		src.perPage = 25

		repos, err := ListAll(ctx, src)
		require.NoError(t, err)

		testutil.AssertGolden(t, "testdata/sources/GERRIT/"+t.Name(), Update(t.Name()), repos)
	})

	t.Run("with filtering", func(t *testing.T) {
		cf, save := NewClientFactory(t, t.Name())
		defer save(t)

		svc := typestest.MakeExternalService(t, extsvc.VariantGerrit, &schema.GerritConnection{
			Projects: []string{
				"src-cli",
			},
			Url:      "https://gerrit.sgdev.org",
			Username: os.Getenv("GERRIT_USERNAME"),
			Password: os.Getenv("GERRIT_PASSWORD"),
		})

		ctx := context.Background()
		src, err := NewGerritSource(ctx, svc, cf)
		require.NoError(t, err)

		src.perPage = 25

		repos, err := ListAll(ctx, src)
		require.NoError(t, err)

		assert.Len(t, repos, 1)
		repoNames := make([]string, 0, len(repos))
		for _, repo := range repos {
			repoNames = append(repoNames, repo.ExternalRepo.ID)
		}
		assert.ElementsMatch(t, repoNames, []string{
			"src-cli",
		})
	})
}

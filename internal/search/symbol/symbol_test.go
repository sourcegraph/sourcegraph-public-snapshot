package symbol

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestFilterZoektResults(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				SubRepoPermissions: &schema.SubRepoPermissions{
					Enabled: true,
				},
			},
		},
	})
	t.Cleanup(func() { conf.Mock(nil) })

	repoName := api.RepoName("foo")
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})
	checker, err := authz.NewSimpleChecker(repoName, []string{"**"}, []string{"*_test.go"})
	if err != nil {
		t.Fatal(err)
	}
	results := []*result.SymbolMatch{
		{
			Symbol: result.Symbol{},
			File: &result.File{
				Path: "foo.go",
			},
		},
		{
			Symbol: result.Symbol{},
			File: &result.File{
				Path: "foo_test.go",
			},
		},
	}
	filtered, err := filterZoektResults(ctx, checker, repoName, results)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, filtered, 1)
	r := filtered[0]
	assert.Equal(t, r.File.Path, "foo.go")
}

package repos

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGoModulesSource_ListRepos(t *testing.T) {
	ctx := context.Background()
	depsSvc := testDependenciesService(ctx, t, []dependencies.Repo{
		{
			ID:      1,
			Scheme:  dependencies.GoModulesScheme,
			Name:    "github.com/gorilla/mux",
			Version: "v1.8.0", // test deduplication with version from config
		},
		{
			ID:      2,
			Scheme:  dependencies.GoModulesScheme,
			Name:    "github.com/gorilla/mux",
			Version: "v1.7.4", // test multiple versions of the same module
		},
		{
			ID:      3,
			Scheme:  dependencies.GoModulesScheme,
			Name:    "github.com/goware/urlx",
			Version: "v0.3.1",
		},
		{
			ID:      4,
			Scheme:  dependencies.GoModulesScheme,
			Name:    "github.com/foo/barbaz",
			Version: "v0.0.1", // Test missing modules are skipped.
		},
	})

	svc := types.ExternalService{
		Kind: extsvc.KindGoModules,
		Config: marshalJSON(t, &schema.GoModulesConnection{
			Urls: []string{
				"https://proxy.golang.org",
			},
			Dependencies: []string{
				"github.com/tsenart/vegeta/v12@v12.8.4",
				"github.com/coreos/go-oidc@v2.2.1+incompatible",
				"github.com/google/zoekt@v0.0.0-20211108135652-f8e8ada171c7",
				"github.com/gorilla/mux@v1.8.0",
			},
		}),
	}

	cf, save := newClientFactory(t, t.Name())
	t.Cleanup(func() { save(t) })

	src, err := NewGoModulesSource(&svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	src.SetDependenciesService(depsSvc)

	repos, err := listAll(ctx, src)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/sources/"+t.Name(), update(t.Name()), repos)
}

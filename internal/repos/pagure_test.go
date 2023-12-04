package repos

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestPagureSource_ListRepos(t *testing.T) {
	conf := &schema.PagureConnection{
		Url:     "https://src.fedoraproject.org",
		Pattern: "ac*",
	}
	cf, save := NewClientFactory(t, t.Name())
	defer save(t)

	svc := typestest.MakeExternalService(t, extsvc.VariantPagure, conf)

	ctx := context.Background()
	src, err := NewPagureSource(ctx, svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	src.perPage = 25 // 2 pages for 47 results

	repos, err := ListAll(context.Background(), src)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/sources/"+t.Name(), Update(t.Name()), repos)
}

package repos

import (
	"context"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestADOSource_ListRepos(t *testing.T) {
	conf := &azuredevops.ADOConnection{
		Username: "testuser",
		Token:    "testtoken",
		Projects: []string{"sgadotest/sgadotest"},
	}
	cf, save := newClientFactory(t, t.Name())
	defer save(t)

	svc := &types.ExternalService{
		Kind:   extsvc.KindADO,
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, conf)),
	}

	ctx := context.Background()
	src, err := NewADOSource(ctx, svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	repos, err := listAll(context.Background(), src)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/sources/AZUREDEVOPS/"+t.Name(), update(t.Name()), repos)
}

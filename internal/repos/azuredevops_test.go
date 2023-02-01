package repos

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAzureDevOpsSource_ListRepos(t *testing.T) {
	conf := &schema.AzureDevOpsConnection{
		Url:      "https://dev.azure.com",
		Username: "testuser",
		Token:    "testtoken",
		Projects: []string{"sgtestazure/sgtestazure", "sgtestazure/sg test with spaces"},
		Exclude: []*schema.ExcludedAzureDevOpsServerRepo{
			{
				Name: "sg test with spaces/sg test with spaces",
			},
			{
				Pattern: "^sgtestazure/sgtestazure[3-9]",
			},
		},
	}
	cf, save := newClientFactory(t, t.Name())
	defer save(t)

	svc := &types.ExternalService{
		Kind:   extsvc.KindAzureDevOps,
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, conf)),
	}

	ctx := context.Background()
	src, err := NewAzureDevOpsSource(ctx, nil, svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	repos, err := listAll(context.Background(), src)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/sources/AZUREDEVOPS/"+t.Name(), update(t.Name()), repos)
}

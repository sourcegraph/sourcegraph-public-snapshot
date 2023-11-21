package repos

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

// To update this test run:
//  1. Set the env AZURE_DEV_OPS_USERNAME and AZURE_DEV_OPS_TOKEN (the secrets can be found in 1Password if you search for Azure test credentials)
//  2. Run the test with the -update flag:
//     `go test -run='TestAzureDevOpsSource_ListRepos' -update=TestAzureDevOpsSource_ListRepos`
func TestAzureDevOpsSource_ListRepos(t *testing.T) {
	ratelimit.SetupForTest(t)

	cf, save := NewClientFactory(t, t.Name())
	defer save(t)

	svc := typestest.MakeExternalService(t, extsvc.VariantAzureDevOps, &schema.AzureDevOpsConnection{
		Url:      "https://dev.azure.com",
		Username: os.Getenv("AZURE_DEV_OPS_USERNAME"),
		Token:    os.Getenv("AZURE_DEV_OPS_TOKEN"),
		Projects: []string{"sgtestazure/sgtestazure", "sgtestazure/sg test with spaces"},
		Exclude: []*schema.ExcludedAzureDevOpsServerRepo{
			{
				Name: "sgtestazure/sg test with spaces/sg test with spaces",
			},
			{
				Pattern: "^sgtestazure/sgtestazure/sgtestazure[3-9]",
			},
		},
	})

	ctx := context.Background()
	src, err := NewAzureDevOpsSource(ctx, nil, svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	repos, err := ListAll(context.Background(), src)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/sources/AZUREDEVOPS/"+t.Name(), Update(t.Name()), repos)
}

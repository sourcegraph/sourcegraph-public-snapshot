package azuredevops

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestClient_AzureServicesProfile(t *testing.T) {
	cli, save := NewTestClient(t, "AzureServicesProfile", *update)
	t.Cleanup(save)

	resp, err := cli.GetAuthorizedProfile(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/AzureServicesProfile.json", *update, resp)
}

// To update this test run:
//  1. Set the env AZURE_DEV_OPS_USERNAME and AZURE_DEV_OPS_TOKEN (the secrets can be found in 1Password if you search for Azure test credentials)
//  2. Run the test with the -update flag:
//     `go test -run='TestClient_ListAuthorizedUserOrganizations' -update=true`
func TestClient_ListAuthorizedUserOrganizations(t *testing.T) {
	// cf, save := newClientFactory(t, t.Name())
	// defer save(t)

	// httpCli, err := cf.Doer()
	// if err != nil {
	// 	t.Fatalf("failed to created http client: %v", err)
	// }

	cli, save := NewTestClient(
		t,
		"ListAuthorizedUserOrganizations",
		*update,
	)
	defer save()

	ctx := context.Background()
	profile, err := cli.GetAuthorizedProfile(ctx)
	if err != nil {
		t.Fatalf("failed to authorized profile: %v", err)
	}

	orgs, err := cli.ListAuthorizedUserOrganizations(ctx, profile)
	if err != nil {
		t.Fatalf("failed to list authorized user origanizations: %v", err)
	}

	testutil.AssertGolden(t, "testdata/golden/ListAuthorizedUserOrganizations.json", *update, orgs)
}

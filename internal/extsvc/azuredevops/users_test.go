package azuredevops

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/stretchr/testify/require"
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
	cli, save := NewTestClient(
		t,
		"ListAuthorizedUserOrganizations",
		*update,
	)
	t.Cleanup(save)

	ctx := context.Background()
	profile, err := cli.GetAuthorizedProfile(ctx)
	require.NoError(t, err, "failed to get authorized profile")

	orgs, err := cli.ListAuthorizedUserOrganizations(ctx, profile)
	require.NoError(t, err, "failed to list authorized user organizations")

	testutil.AssertGolden(t, "testdata/golden/ListAuthorizedUserOrganizations.json", *update, orgs)
}

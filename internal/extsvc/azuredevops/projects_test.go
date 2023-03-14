package azuredevops

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestClient_GetProject(t *testing.T) {
	cli, save := NewTestClient(t, "GetProject", *update)
	t.Cleanup(save)

	resp, err := cli.GetProject(context.Background(), "sgtestazure", "sgtestazure")
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/GetProject.json", *update, resp)
}

// To update this test run:
//  1. Set the env AZURE_DEV_OPS_USERNAME and AZURE_DEV_OPS_TOKEN (the secrets can be found in 1Password if you search for Azure test credentials)
//  2. Run the test with the -update flag:
//     `go test -run='TestClient_ListAuthorizedUserOrganizations' -update=true`
func TestClient_ListAuthorizedUserProjects(t *testing.T) {
	client, save := NewTestClient(
		t,
		"ListAuthorizedUserProjects",
		*update,
	)
	t.Cleanup(save)

	ctx := context.Background()

	profile, err := client.GetAuthorizedProfile(ctx)
	if err != nil {
		t.Fatalf("failed to get authorized profile: %v", err)
	}

	orgs, err := client.ListAuthorizedUserOrganizations(ctx, profile)
	if err != nil {
		t.Fatalf("failed to list authorized user origanizations: %v", err)
	}

	allProjects := []Project{}
	for _, org := range orgs {
		projects, err := client.ListAuthorizedUserProjects(ctx, org.Name)
		if err != nil {
			t.Fatalf("failed to list authorized user origanizations: %v", err)
		}
		allProjects = append(allProjects, projects...)
	}
	testutil.AssertGolden(t, "testdata/golden/ListAuthorizedUserProjects.json", *update, allProjects)
}

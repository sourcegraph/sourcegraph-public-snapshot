package azuredevops

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestClient_AzureServicesProfile(t *testing.T) {
	cli, save := NewTestClient(t, "AzureServicesProfile", *update)
	t.Cleanup(save)

	resp, err := cli.AzureServicesProfile(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/AzureServicesProfile.json", *update, resp)
}

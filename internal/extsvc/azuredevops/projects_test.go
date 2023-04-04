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

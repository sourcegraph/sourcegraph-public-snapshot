package azuredevops

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestClient_GetRepository(t *testing.T) {
	cli, save := NewTestClient(t, "GetRepository", *update)
	t.Cleanup(save)

	args := OrgProjectRepoArgs{
		Org:          "sgtestazure",
		Project:      "sgtestazure",
		RepoNameOrID: "sgtestazure",
	}

	resp, err := cli.GetRepo(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/GetRepository.json", *update, resp)
}

func TestClient_ListRepositoriesByProjectOrOrg(t *testing.T) {
	cli, save := NewTestClient(t, "ListRepositoriesByProjectOrOrg", *update)
	t.Cleanup(save)

	opts := ListRepositoriesByProjectOrOrgArgs{
		ProjectOrOrgName: "sgtestazure",
	}

	resp, err := cli.ListRepositoriesByProjectOrOrg(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/ListProjects.json", *update, resp)
}

func TestClient_ForkRepository(t *testing.T) {
	cli, save := NewTestClient(t, "ForkRepository", *update)
	t.Cleanup(save)

	input := ForkRepositoryInput{
		Name: "sgtestazureforks2",
		Project: ForkRepositoryInputProject{
			ID: "dc493f7d-0b57-4de2-a59b-3f74ff3ea334",
		},
		ParentRepository: ForkRepositoryInputParentRepository{
			ID: "c4d186ef-18a6-4de4-a610-aa9ebd4e1faa",
			Project: ForkRepositoryInputProject{
				ID: "dc493f7d-0b57-4de2-a59b-3f74ff3ea334",
			},
		},
	}

	resp, err := cli.ForkRepository(context.Background(), "sgtestazure", input)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/ForkRepository.json", *update, resp)
}

func TestClient_GetRepositoryBranch(t *testing.T) {
	cli, save := NewTestClient(t, "GetRepositoryBranch", *update)
	t.Cleanup(save)

	args := OrgProjectRepoArgs{
		Org:          "sgtestazure",
		Project:      "sgtestazure",
		RepoNameOrID: "sgtestazure3",
	}

	resp, err := cli.GetRepositoryBranch(context.Background(), args, "main")
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/GetRepositoryBranch.json", *update, resp)
}

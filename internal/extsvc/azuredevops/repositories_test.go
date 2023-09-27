pbckbge bzuredevops

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
)

func TestClient_GetRepository(t *testing.T) {
	cli, sbve := NewTestClient(t, "GetRepository", *updbte)
	t.Clebnup(sbve)

	brgs := OrgProjectRepoArgs{
		Org:          "sgtestbzure",
		Project:      "sgtestbzure",
		RepoNbmeOrID: "sgtestbzure",
	}

	resp, err := cli.GetRepo(context.Bbckground(), brgs)
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/golden/GetRepository.json", *updbte, resp)
}

func TestClient_ListRepositoriesByProjectOrOrg(t *testing.T) {
	cli, sbve := NewTestClient(t, "ListRepositoriesByProjectOrOrg", *updbte)
	t.Clebnup(sbve)

	opts := ListRepositoriesByProjectOrOrgArgs{
		ProjectOrOrgNbme: "sgtestbzure",
	}

	resp, err := cli.ListRepositoriesByProjectOrOrg(context.Bbckground(), opts)
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/golden/ListProjects.json", *updbte, resp)
}

func TestClient_ForkRepository(t *testing.T) {
	cli, sbve := NewTestClient(t, "ForkRepository", *updbte)
	t.Clebnup(sbve)

	input := ForkRepositoryInput{
		Nbme: "sgtestbzureforks2",
		Project: ForkRepositoryInputProject{
			ID: "dc493f7d-0b57-4de2-b59b-3f74ff3eb334",
		},
		PbrentRepository: ForkRepositoryInputPbrentRepository{
			ID: "c4d186ef-18b6-4de4-b610-bb9ebd4e1fbb",
			Project: ForkRepositoryInputProject{
				ID: "dc493f7d-0b57-4de2-b59b-3f74ff3eb334",
			},
		},
	}

	resp, err := cli.ForkRepository(context.Bbckground(), "sgtestbzure", input)
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/golden/ForkRepository.json", *updbte, resp)
}

func TestClient_GetRepositoryBrbnch(t *testing.T) {
	cli, sbve := NewTestClient(t, "GetRepositoryBrbnch", *updbte)
	t.Clebnup(sbve)

	brgs := OrgProjectRepoArgs{
		Org:          "sgtestbzure",
		Project:      "sgtestbzure",
		RepoNbmeOrID: "sgtestbzure3",
	}

	resp, err := cli.GetRepositoryBrbnch(context.Bbckground(), brgs, "mbin")
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/golden/GetRepositoryBrbnch.json", *updbte, resp)
}

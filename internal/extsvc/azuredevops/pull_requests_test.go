package azuredevops

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestClient_CreatePullRequest(t *testing.T) {
	cli, save := NewTestClient(t, "CreatePullRequest", *update)
	t.Cleanup(save)

	args := OrgProjectRepoArgs{
		Org:          "sgtestazure",
		Project:      "sgtestazure",
		RepoNameOrID: "c4d186ef-18a6-4de4-a610-aa9ebd4e1faa",
	}

	input := CreatePullRequestInput{
		SourceRefName: "refs/heads/add-codeowners",
		TargetRefName: "refs/heads/main",
		Title:         "Test PR",
		Description:   "test description",
	}

	resp, err := cli.CreatePullRequest(context.Background(), args, input)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/CreatePullRequest.json", *update, resp)
}

func TestClient_AbandonPullRequest(t *testing.T) {
	cli, save := NewTestClient(t, "AbandonPullRequest", *update)
	t.Cleanup(save)

	// When updating this test make sure you point these args to an active PR.
	args := PullRequestCommonArgs{
		PullRequestID: "2",
		Org:           "sgtestazure",
		Project:       "sgtestazure",
		RepoNameOrID:  "sgtestazure3",
	}

	resp, err := cli.AbandonPullRequest(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/AbandonPullRequest.json", *update, resp)
}

func TestClient_GetPullRequest(t *testing.T) {
	cli, save := NewTestClient(t, "GetPullRequest", *update)
	t.Cleanup(save)

	// When updating this test make sure you point these args to an active PR.
	args := PullRequestCommonArgs{
		PullRequestID: "2",
		Org:           "sgtestazure",
		Project:       "sgtestazure",
		RepoNameOrID:  "sgtestazure3",
	}

	resp, err := cli.GetPullRequest(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/GetPullRequest.json", *update, resp)
}

func TestClient_GetPullRequestStatuses(t *testing.T) {
	cli, save := NewTestClient(t, "GetPullRequestStatuses", *update)
	t.Cleanup(save)

	// When updating this test make sure you point these args to an active PR.
	args := PullRequestCommonArgs{
		PullRequestID: "8",
		Org:           "sgtestazure",
		Project:       "sgtestazure",
		RepoNameOrID:  "sgtestazure3",
	}

	resp, err := cli.GetPullRequestStatuses(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/GetPullRequestStatuses.json", *update, resp)
}

func TestClient_UpdatePullRequest(t *testing.T) {
	cli, save := NewTestClient(t, "UpdatePullRequest", *update)
	t.Cleanup(save)

	// When updating this test make sure you point these args to an active PR.
	title := "new title"
	description := "new description"
	args := PullRequestCommonArgs{
		PullRequestID: "5",
		Org:           "sgtestazure",
		Project:       "sgtestazure",
		RepoNameOrID:  "sgtestazure3",
	}
	input := PullRequestUpdateInput{
		Title:       &title,
		Description: &description,
	}

	resp, err := cli.UpdatePullRequest(context.Background(), args, input)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/UpdatePullRequest.json", *update, resp)
}

func TestClient_CreatePullRequestCommentThread(t *testing.T) {
	cli, save := NewTestClient(t, "CreatePullRequestCommentThread", *update)
	t.Cleanup(save)

	// When updating this test make sure you point these args to an active PR.
	args := PullRequestCommonArgs{
		PullRequestID: "5",
		Org:           "sgtestazure",
		Project:       "sgtestazure",
		RepoNameOrID:  "sgtestazure3",
	}
	input := PullRequestCommentInput{
		Comments: []PullRequestCommentForInput{{
			ParentCommitID: 0,
			Content:        "new comment",
			CommentType:    1,
		}},
	}

	resp, err := cli.CreatePullRequestCommentThread(context.Background(), args, input)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/CreatePullRequestCommentThread.json", *update, resp)
}

func TestClient_CompletePullRequest(t *testing.T) {
	cli, save := NewTestClient(t, "CompletePullRequest", *update)
	t.Cleanup(save)

	// When updating this test make sure you point these args to an active PR.
	args := PullRequestCommonArgs{
		PullRequestID: "7",
		Org:           "sgtestazure",
		Project:       "sgtestazure",
		RepoNameOrID:  "sgtestazure3",
	}
	input := PullRequestCommitRef{
		CommitID: "7867eee4f671d69f9a9be0fcacb6a7af764ec962",
	}

	resp, err := cli.CompletePullRequest(context.Background(), args, input)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/CompletePullRequest.json", *update, resp)
}

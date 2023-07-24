package sources

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	adobatches "github.com/sourcegraph/sourcegraph/internal/batches/sources/azuredevops"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var (
	testRepoName              = "testrepo"
	testProjectName           = "testproject"
	testOrgName               = "testorg"
	testPRID                  = "42"
	testRepository            = azuredevops.Repository{ID: "testrepoid", Name: testRepoName, Project: azuredevops.Project{ID: "testprojectid", Name: testProjectName}, APIURL: fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s", testOrgName, testProjectName, testRepoName), CloneURL: fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s", testOrgName, testProjectName, testRepoName)}
	testCommonPullRequestArgs = azuredevops.PullRequestCommonArgs{Org: testOrgName, Project: testProjectName, RepoNameOrID: testRepoName, PullRequestID: testPRID}
	testOrgProjectRepoArgs    = azuredevops.OrgProjectRepoArgs{Org: testOrgName, Project: testProjectName, RepoNameOrID: testRepoName}
)

func TestAzureDevOpsSource_GitserverPushConfig(t *testing.T) {
	// This isn't a full blown test of all the possibilities of
	// gitserverPushConfig(), but we do need to validate that the authenticator
	// on the client affects the eventual URL in the correct way, and that
	// requires a bunch of boilerplate to make it look like we have a valid
	// external service and repo.
	//
	// So, cue the boilerplate:
	au := auth.BasicAuth{Username: "user", Password: "pass"}

	s, client := mockAzureDevOpsSource()
	client.AuthenticatorFunc.SetDefaultReturn(&au)

	repo := &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: extsvc.TypeAzureDevOps,
		},
		Metadata: &azuredevops.Repository{
			ID:   "testrepoid",
			Name: "testrepo",
			Project: azuredevops.Project{
				ID:   "testprojectid",
				Name: "testproject",
			},
		},
		Sources: map[string]*types.SourceInfo{
			"1": {
				ID:       "extsvc:azuredevops:1",
				CloneURL: "https://dev.azure.com/testorg/testproject/_git/testrepo",
			},
		},
	}

	pushConfig, err := s.GitserverPushConfig(repo)
	assert.Nil(t, err)
	assert.NotNil(t, pushConfig)
	assert.Equal(t, "https://user:pass@dev.azure.com/testorg/testproject/_git/testrepo", pushConfig.RemoteURL)
}

func TestAzureDevOpsSource_WithAuthenticator(t *testing.T) {
	t.Run("supports BasicAuth", func(t *testing.T) {
		newClient := NewStrictMockAzureDevOpsClient()
		au := &auth.BasicAuth{}
		s, client := mockAzureDevOpsSource()
		client.WithAuthenticatorFunc.SetDefaultHook(func(a auth.Authenticator) (azuredevops.Client, error) {
			assert.Same(t, au, a)
			return newClient, nil
		})

		newSource, err := s.WithAuthenticator(au)
		assert.Nil(t, err)
		assert.Same(t, newClient, newSource.(*AzureDevOpsSource).client)
	})
}

func TestAzureDevOpsSource_ValidateAuthenticator(t *testing.T) {
	ctx := context.Background()

	for name, want := range map[string]error{
		"nil":   nil,
		"error": errors.New("error"),
	} {
		t.Run(name, func(t *testing.T) {
			s, client := mockAzureDevOpsSource()
			client.GetAuthorizedProfileFunc.SetDefaultReturn(azuredevops.Profile{}, want)

			assert.Equal(t, want, s.ValidateAuthenticator(ctx))
		})
	}
}

func TestAzureDevOpsSource_LoadChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := errors.New("error")
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return azuredevops.PullRequest{}, want
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return azuredevops.PullRequest{}, &notFoundError{}
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		target := ChangesetNotFoundError{}
		assert.ErrorAs(t, err, &target)
		assert.Same(t, target.Changeset, cs)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := mockAzureDevOpsAnnotatePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotatePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})

		err := s.LoadChangeset(ctx, cs)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_CreateChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error creating pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()

		want := errors.New("error")
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.OrgProjectRepoArgs, pri azuredevops.CreatePullRequestInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testOrgProjectRepoArgs, r)
			assert.Equal(t, cs.Title, pri.Title)
			return azuredevops.PullRequest{}, want
		})

		exists, err := s.CreateChangeset(ctx, cs)
		assert.False(t, exists)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := mockAzureDevOpsAnnotatePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.OrgProjectRepoArgs, pri azuredevops.CreatePullRequestInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testOrgProjectRepoArgs, r)
			assert.Equal(t, cs.Title, pri.Title)
			return *pr, nil
		})

		exists, err := s.CreateChangeset(ctx, cs)
		assert.False(t, exists)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotatePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.OrgProjectRepoArgs, pri azuredevops.CreatePullRequestInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testOrgProjectRepoArgs, r)
			assert.Equal(t, cs.Title, pri.Title)
			assert.Nil(t, pri.ForkSource)
			return *pr, nil
		})

		exists, err := s.CreateChangeset(ctx, cs)
		assert.True(t, exists)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})

	t.Run("success with fork", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotatePullRequestSuccess(client)

		fork := &azuredevops.Repository{
			ID:   "forkedrepoid",
			Name: "forkedrepo",
		}
		cs.RemoteRepo = &types.Repo{
			Metadata: fork,
		}

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.OrgProjectRepoArgs, pri azuredevops.CreatePullRequestInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testOrgProjectRepoArgs, r)
			assert.Equal(t, cs.Title, pri.Title)
			assert.Equal(t, *fork, pri.ForkSource.Repository)
			return *pr, nil
		})

		exists, err := s.CreateChangeset(ctx, cs)
		assert.True(t, exists)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_CreateDraftChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error creating pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()

		want := errors.New("error")
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.OrgProjectRepoArgs, pri azuredevops.CreatePullRequestInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testOrgProjectRepoArgs, r)
			assert.Equal(t, cs.Title, pri.Title)
			return azuredevops.PullRequest{}, want
		})

		exists, err := s.CreateDraftChangeset(ctx, cs)
		assert.False(t, exists)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := mockAzureDevOpsAnnotatePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.OrgProjectRepoArgs, pri azuredevops.CreatePullRequestInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testOrgProjectRepoArgs, r)
			assert.Equal(t, cs.Title, pri.Title)
			return *pr, nil
		})

		exists, err := s.CreateDraftChangeset(ctx, cs)
		assert.False(t, exists)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotatePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.OrgProjectRepoArgs, pri azuredevops.CreatePullRequestInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testOrgProjectRepoArgs, r)
			assert.Equal(t, cs.Title, pri.Title)
			assert.Nil(t, pri.ForkSource)
			assert.True(t, pri.IsDraft)
			return *pr, nil
		})

		exists, err := s.CreateDraftChangeset(ctx, cs)
		assert.True(t, exists)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})

	t.Run("success with fork", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotatePullRequestSuccess(client)

		fork := &azuredevops.Repository{
			ID:   "forkedrepoid",
			Name: "forkedrepo",
		}
		cs.RemoteRepo = &types.Repo{
			Metadata: fork,
		}

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CreatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.OrgProjectRepoArgs, pri azuredevops.CreatePullRequestInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testOrgProjectRepoArgs, r)
			assert.Equal(t, cs.Title, pri.Title)
			assert.Equal(t, *fork, pri.ForkSource.Repository)
			assert.True(t, pri.IsDraft)
			return *pr, nil
		})

		exists, err := s.CreateDraftChangeset(ctx, cs)
		assert.True(t, exists)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_CloseChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error declining pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		want := errors.New("error")
		client.AbandonPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return azuredevops.PullRequest{}, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CloseChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := mockAzureDevOpsAnnotatePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.AbandonPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CloseChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotatePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.AbandonPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CloseChangeset(ctx, cs)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_UpdateChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := errors.New("error")
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return azuredevops.PullRequest{}, want
		})

		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error updating pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := errors.New("error")
		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})
		client.UpdatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, pri azuredevops.PullRequestUpdateInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Equal(t, cs.Title, *pri.Title)
			return azuredevops.PullRequest{}, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := mockAzureDevOpsAnnotatePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})
		client.UpdatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, pri azuredevops.PullRequestUpdateInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Equal(t, cs.Title, *pri.Title)
			return *pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotatePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})
		client.UpdatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, pri azuredevops.PullRequestUpdateInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Equal(t, cs.Title, *pri.Title)
			return *pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.UpdateChangeset(ctx, cs)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_UndraftChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error updating pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := errors.New("error")
		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.UpdatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, pri azuredevops.PullRequestUpdateInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Equal(t, cs.Title, *pri.Title)
			return azuredevops.PullRequest{}, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.UndraftChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := mockAzureDevOpsAnnotatePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.UpdatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, pri azuredevops.PullRequestUpdateInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Equal(t, cs.Title, *pri.Title)
			return *pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.UndraftChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotatePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.UpdatePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, pri azuredevops.PullRequestUpdateInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Equal(t, cs.Title, *pri.Title)
			assert.False(t, *pri.IsDraft)
			return *pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.UndraftChangeset(ctx, cs)
		assert.Nil(t, err)
		assertChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_CreateComment(t *testing.T) {
	ctx := context.Background()

	t.Run("error creating comment", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		want := errors.New("error")
		client.CreatePullRequestCommentThreadFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, ci azuredevops.PullRequestCommentInput) (azuredevops.PullRequestCommentResponse, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Equal(t, "comment", ci.Comments[0].Content)
			return azuredevops.PullRequestCommentResponse{}, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CreateComment(ctx, cs, "comment")
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CreatePullRequestCommentThreadFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, ci azuredevops.PullRequestCommentInput) (azuredevops.PullRequestCommentResponse, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Equal(t, "comment", ci.Comments[0].Content)
			return azuredevops.PullRequestCommentResponse{}, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.CreateComment(ctx, cs, "comment")
		assert.Nil(t, err)
	})
}

func TestAzureDevOpsSource_MergeChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error merging pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		want := errors.New("error")
		client.CompletePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, input azuredevops.PullRequestCompleteInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Nil(t, input.MergeStrategy)
			return azuredevops.PullRequest{}, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		target := ChangesetNotMergeableError{}
		assert.ErrorAs(t, err, &target)
		assert.Equal(t, want.Error(), target.ErrorMsg)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		want := &notFoundError{}
		client.CompletePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, input azuredevops.PullRequestCompleteInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Nil(t, input.MergeStrategy)
			return azuredevops.PullRequest{}, want
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("error setting changeset metadata", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChangeset()
		s, client := mockAzureDevOpsSource()
		want := mockAzureDevOpsAnnotatePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CompletePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, input azuredevops.PullRequestCompleteInput) (azuredevops.PullRequest, error) {
			assert.Equal(t, testCommonPullRequestArgs, r)
			assert.Nil(t, input.MergeStrategy)
			return *pr, nil
		})

		annotateChangesetWithPullRequest(cs, pr)
		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		squash := azuredevops.PullRequestMergeStrategySquash
		for name, tc := range map[string]struct {
			squash bool
			want   *azuredevops.PullRequestMergeStrategy
		}{
			"no squash": {false, nil},
			"squash":    {true, &squash},
		} {
			t.Run(name, func(t *testing.T) {
				cs, _ := mockAzureDevOpsChangeset()
				s, client := mockAzureDevOpsSource()
				mockAzureDevOpsAnnotatePullRequestSuccess(client)

				pr := mockAzureDevOpsPullRequest(&testRepository)
				client.CompletePullRequestFunc.SetDefaultHook(func(ctx context.Context, r azuredevops.PullRequestCommonArgs, input azuredevops.PullRequestCompleteInput) (azuredevops.PullRequest, error) {
					assert.Equal(t, testCommonPullRequestArgs, r)
					assert.Equal(t, tc.want, input.MergeStrategy)
					return *pr, nil
				})

				annotateChangesetWithPullRequest(cs, pr)
				err := s.MergeChangeset(ctx, cs, tc.squash)
				assert.Nil(t, err)
				assertChangesetMatchesPullRequest(t, cs, pr)
			})
		}
	})
}

func TestAzureDevOpsSource_GetFork(t *testing.T) {
	ctx := context.Background()

	upstream := testRepository
	urn := extsvc.URN(extsvc.KindAzureDevOps, 1)
	upstreamRepo := &types.Repo{Metadata: &upstream, Sources: map[string]*types.SourceInfo{
		urn: {
			ID:       urn,
			CloneURL: "https://dev.azure.com/testorg/testproject/_git/testrepo",
		},
	}}

	args := azuredevops.OrgProjectRepoArgs{
		Org:          testOrgName,
		Project:      "fork",
		RepoNameOrID: "testproject-testrepo",
	}

	fork := azuredevops.Repository{
		ID:   "forkid",
		Name: "testproject-testrepo",
		Project: azuredevops.Project{
			ID:   "testprojectid",
			Name: "fork",
		},
		IsFork: true,
	}

	forkRespositoryInput := azuredevops.ForkRepositoryInput{
		Name: "testproject-testrepo",
		Project: azuredevops.ForkRepositoryInputProject{
			ID: fork.Project.ID,
		},
		ParentRepository: azuredevops.ForkRepositoryInputParentRepository{
			ID: "testrepoid",
			Project: azuredevops.ForkRepositoryInputProject{
				ID: fork.Project.ID,
			},
		},
	}

	t.Run("error checking for repo", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		want := errors.New("error")
		client.GetRepoFunc.SetDefaultHook(func(ctx context.Context, a azuredevops.OrgProjectRepoArgs) (azuredevops.Repository, error) {
			assert.Equal(t, args, a)
			return azuredevops.Repository{}, want
		})

		repo, err := s.GetFork(ctx, upstreamRepo, pointers.Ptr("fork"), nil)
		assert.Nil(t, repo)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("forked repo already exists", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefaultHook(func(ctx context.Context, a azuredevops.OrgProjectRepoArgs) (azuredevops.Repository, error) {
			assert.Equal(t, args, a)
			return fork, nil
		})

		forkRepo, err := s.GetFork(ctx, upstreamRepo, pointers.Ptr("fork"), nil)
		assert.Nil(t, err)
		assert.NotNil(t, forkRepo)
		assert.NotEqual(t, forkRepo, upstreamRepo)
		assert.Equal(t, &fork, forkRepo.Metadata)
		assert.Equal(t, "https://dev.azure.com/testorg/fork/_git/testproject-testrepo", forkRepo.Sources[urn].CloneURL)
	})

	t.Run("get project error", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefaultHook(func(ctx context.Context, a azuredevops.OrgProjectRepoArgs) (azuredevops.Repository, error) {
			assert.Equal(t, args, a)
			return azuredevops.Repository{}, &notFoundError{}
		})
		want := errors.New("error")
		client.GetProjectFunc.SetDefaultHook(func(ctx context.Context, org string, project string) (azuredevops.Project, error) {
			assert.Equal(t, testOrgName, org)
			assert.Equal(t, fork.Project.Name, project)
			return azuredevops.Project{}, want
		})

		repo, err := s.GetFork(ctx, upstreamRepo, pointers.Ptr("fork"), nil)
		assert.Nil(t, repo)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("fork error", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefaultHook(func(ctx context.Context, a azuredevops.OrgProjectRepoArgs) (azuredevops.Repository, error) {
			assert.Equal(t, args, a)
			return azuredevops.Repository{}, &notFoundError{}
		})

		client.GetProjectFunc.SetDefaultHook(func(ctx context.Context, org string, project string) (azuredevops.Project, error) {
			assert.Equal(t, testOrgName, org)
			assert.Equal(t, fork.Project.Name, project)
			return fork.Project, nil
		})

		want := errors.New("error")
		client.ForkRepositoryFunc.SetDefaultHook(func(ctx context.Context, org string, fi azuredevops.ForkRepositoryInput) (azuredevops.Repository, error) {
			assert.Equal(t, testOrgName, org)
			assert.Equal(t, forkRespositoryInput, fi)
			return azuredevops.Repository{}, want
		})

		repo, err := s.GetFork(ctx, upstreamRepo, pointers.Ptr("fork"), nil)
		assert.Nil(t, repo)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success with default namespace, name", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefaultHook(func(ctx context.Context, a azuredevops.OrgProjectRepoArgs) (azuredevops.Repository, error) {
			argsNew := args
			argsNew.Project = testProjectName
			assert.Equal(t, argsNew, a)
			return fork, nil
		})

		repo, err := s.GetFork(ctx, upstreamRepo, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, repo)
		assert.Equal(t, &fork, repo.Metadata)
	})

	t.Run("success with default name", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefaultHook(func(ctx context.Context, a azuredevops.OrgProjectRepoArgs) (azuredevops.Repository, error) {
			assert.Equal(t, args, a)
			return azuredevops.Repository{}, &notFoundError{}
		})

		client.GetProjectFunc.SetDefaultHook(func(ctx context.Context, org string, project string) (azuredevops.Project, error) {
			assert.Equal(t, testOrgName, org)
			assert.Equal(t, fork.Project.Name, project)
			return fork.Project, nil
		})

		client.ForkRepositoryFunc.SetDefaultHook(func(ctx context.Context, org string, fi azuredevops.ForkRepositoryInput) (azuredevops.Repository, error) {
			assert.Equal(t, testOrgName, org)
			assert.Equal(t, forkRespositoryInput, fi)
			return fork, nil
		})

		forkRepo, err := s.GetFork(ctx, upstreamRepo, pointers.Ptr("fork"), nil)
		assert.Nil(t, err)
		assert.NotNil(t, forkRepo)
		assert.NotEqual(t, forkRepo, upstreamRepo)
		assert.Equal(t, &fork, forkRepo.Metadata)
		assert.Equal(t, "https://dev.azure.com/testorg/fork/_git/testproject-testrepo", forkRepo.Sources[urn].CloneURL)
	})

	t.Run("success with set namespace, name", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefaultHook(func(ctx context.Context, a azuredevops.OrgProjectRepoArgs) (azuredevops.Repository, error) {
			newArgs := args
			newArgs.RepoNameOrID = "special-fork-name"
			assert.Equal(t, newArgs, a)
			return azuredevops.Repository{}, &notFoundError{}
		})

		client.GetProjectFunc.SetDefaultHook(func(ctx context.Context, org string, project string) (azuredevops.Project, error) {
			assert.Equal(t, testOrgName, org)
			assert.Equal(t, fork.Project.Name, project)
			return fork.Project, nil
		})

		client.ForkRepositoryFunc.SetDefaultHook(func(ctx context.Context, org string, fi azuredevops.ForkRepositoryInput) (azuredevops.Repository, error) {
			assert.Equal(t, testOrgName, org)
			newFRI := forkRespositoryInput
			newFRI.Name = "special-fork-name"
			assert.Equal(t, newFRI, fi)
			return fork, nil
		})

		forkRepo, err := s.GetFork(ctx, upstreamRepo, pointers.Ptr("fork"), pointers.Ptr("special-fork-name"))
		assert.Nil(t, err)
		assert.NotNil(t, forkRepo)
		assert.NotEqual(t, forkRepo, upstreamRepo)
		assert.Equal(t, &fork, forkRepo.Metadata)
		assert.Equal(t, "https://dev.azure.com/testorg/fork/_git/testproject-testrepo", forkRepo.Sources[urn].CloneURL)
	})
}

func TestAzureDevOpsSource_annotatePullRequest(t *testing.T) {
	// The case where GetPullRequestStatuses errors and where it returns an
	// empty result set are thoroughly covered in other tests, so we'll just
	// handle the other branches of annotatePullRequest.

	ctx := context.Background()

	t.Run("error getting all statuses", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()
		pr := mockAzureDevOpsPullRequest(&testRepository)

		want := errors.New("error")
		client.GetPullRequestStatusesFunc.SetDefaultHook(func(ctx context.Context, args azuredevops.PullRequestCommonArgs) ([]azuredevops.PullRequestBuildStatus, error) {
			assert.Equal(t, testCommonPullRequestArgs, args)
			return nil, want
		})

		apr, err := s.annotatePullRequest(ctx, &testRepository, pr)
		assert.Nil(t, apr)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()
		pr := mockAzureDevOpsPullRequest(&testRepository)

		want := []*azuredevops.PullRequestBuildStatus{
			{ID: 1},
		}
		client.GetPullRequestStatusesFunc.SetDefaultHook(func(ctx context.Context, args azuredevops.PullRequestCommonArgs) ([]azuredevops.PullRequestBuildStatus, error) {
			assert.Equal(t, args, testCommonPullRequestArgs)
			return []azuredevops.PullRequestBuildStatus{
				{
					ID: 1,
				},
			}, nil
		})

		apr, err := s.annotatePullRequest(ctx, &testRepository, pr)
		assert.Nil(t, err)
		assert.NotNil(t, apr)
		assert.Same(t, pr, apr.PullRequest)

		for index, w := range want {
			assert.Equal(t, w, apr.Statuses[index])
		}
	})
}

func assertChangesetMatchesPullRequest(t *testing.T, cs *Changeset, pr *azuredevops.PullRequest) {
	t.Helper()

	// We're not thoroughly testing setChangesetMetadata() et al in this
	// assertion, but we do want to ensure that the PR was used to populate
	// fields on the Changeset.
	assert.EqualValues(t, strconv.Itoa(pr.ID), cs.ExternalID)
	assert.Equal(t, pr.SourceRefName, cs.ExternalBranch)

	if pr.ForkSource != nil {
		assert.Equal(t, pr.ForkSource.Repository.Namespace(), cs.ExternalForkNamespace)
	} else {
		assert.Empty(t, cs.ExternalForkNamespace)
	}
}

// mockAzureDevOpsChangeset creates a plausible non-forked changeset, repo,
// and AzureDevOps specific repo.
func mockAzureDevOpsChangeset() (*Changeset, *types.Repo) {
	repo := &types.Repo{Metadata: &testRepository}
	cs := &Changeset{
		Title: "title",
		Body:  "description",
		Changeset: &btypes.Changeset{
			ExternalID: testPRID,
		},
		RemoteRepo: repo,
		TargetRepo: repo,
		BaseRef:    "refs/heads/targetbranch",
	}

	return cs, repo
}

// mockAzureDevOpsPullRequest returns a plausible pull request that would be
// returned from Bitbucket Cloud for a non-forked changeset.
func mockAzureDevOpsPullRequest(repo *azuredevops.Repository) *azuredevops.PullRequest {
	return &azuredevops.PullRequest{
		ID:            42,
		SourceRefName: "refs/heads/sourcebranch",
		TargetRefName: "refs/heads/targetbranch",
		Repository:    *repo,
		Title:         "TestPR",
	}
}

func annotateChangesetWithPullRequest(cs *Changeset, pr *azuredevops.PullRequest) {
	cs.Metadata = &adobatches.AnnotatedPullRequest{
		PullRequest: pr,
		Statuses:    []*azuredevops.PullRequestBuildStatus{},
	}
}

func mockAzureDevOpsSource() (*AzureDevOpsSource, *MockAzureDevOpsClient) {
	client := NewStrictMockAzureDevOpsClient()
	s := &AzureDevOpsSource{client: client}

	return s, client
}

// mockAzureDevOpsAnnotatePullRequestError configures the mock client to return an error
// when GetPullRequestStatuses is invoked by annotatePullRequest.
func mockAzureDevOpsAnnotatePullRequestError(client *MockAzureDevOpsClient) error {
	err := errors.New("error")
	client.GetPullRequestStatusesFunc.SetDefaultReturn(nil, err)

	return err
}

// mockAzureDevOpsAnnotatePullRequestSuccess configures the mock client to be able to
// return a valid, empty set of statuses.
func mockAzureDevOpsAnnotatePullRequestSuccess(client *MockAzureDevOpsClient) {
	client.GetPullRequestStatusesFunc.SetDefaultReturn([]azuredevops.PullRequestBuildStatus{}, nil)
}

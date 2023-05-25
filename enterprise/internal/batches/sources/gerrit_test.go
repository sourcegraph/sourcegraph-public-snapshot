package sources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	testGerritProjectName = "testrepo"
	testGerritChangeID    = "42"
	testProject           = gerrit.Project{ID: "testrepoid", Name: testGerritProjectName}
)

func TestGerritSource_GitserverPushConfig(t *testing.T) {
	// This isn't a full blown test of all the possibilities of
	// gitserverPushConfig(), but we do need to validate that the authenticator
	// on the client affects the eventual URL in the correct way, and that
	// requires a bunch of boilerplate to make it look like we have a valid
	// external service and repo.
	//
	// So, cue the boilerplate:
	au := auth.BasicAuth{Username: "user", Password: "pass"}
	s, client := mockGerritSource()
	client.AuthenticatorFunc.SetDefaultReturn(&au)

	repo := &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: extsvc.TypeGerrit,
		},
		Metadata: &gerrit.Project{
			ID:   "testrepoid",
			Name: "testrepo",
		},
		Sources: map[string]*types.SourceInfo{
			"1": {
				ID:       "extsvc:gerrit:1",
				CloneURL: "https://gerrit.sgdev.org/testrepo",
			},
		},
	}

	pushConfig, err := s.GitserverPushConfig(repo)
	assert.Nil(t, err)
	assert.NotNil(t, pushConfig)
	assert.Equal(t, "https://user:pass@gerrit.sgdev.org/testrepo", pushConfig.RemoteURL)
}

func TestGerritSource_WithAuthenticator(t *testing.T) {
	t.Run("supports BasicAuth", func(t *testing.T) {
		newClient := NewStrictMockGerritClient()
		au := &auth.BasicAuth{}
		s, client := mockGerritSource()
		client.WithAuthenticatorFunc.SetDefaultHook(func(a auth.Authenticator) (gerrit.Client, error) {
			assert.Same(t, au, a)
			return newClient, nil
		})

		newSource, err := s.WithAuthenticator(au)
		assert.Nil(t, err)
		assert.Same(t, newClient, newSource.(*GerritSource).client)
	})
}

func TestGerritSource_ValidateAuthenticator(t *testing.T) {
	ctx := context.Background()

	for name, want := range map[string]error{
		"nil":   nil,
		"error": errors.New("error"),
	} {
		t.Run(name, func(t *testing.T) {
			s, client := mockGerritSource()
			client.GetAuthenticatedUserAccountFunc.SetDefaultReturn(&gerrit.Account{}, want)

			assert.Equal(t, want, s.ValidateAuthenticator(ctx))
		})
	}
}

func TestGerritSource_LoadChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()
		want := errors.New("error")
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, testGerritChangeID)
			return &gerrit.Change{}, want
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, testGerritChangeID)
			return &gerrit.Change{}, &notFoundError{}
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		target := ChangesetNotFoundError{}
		assert.ErrorAs(t, err, &target)
		assert.Same(t, target.Changeset, cs)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		change := mockGerritChange(&testProject)
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, testGerritChangeID)
			return change, nil
		})

		err := s.LoadChangeset(ctx, cs)
		assert.Nil(t, err)
	})
}

func TestGerritSource_CloseChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error declining pull request", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		want := errors.New("error")
		client.AbandonChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, testGerritChangeID)
			return &gerrit.Change{}, want
		})

		err := s.CloseChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		pr := mockGerritChange(&testProject)
		client.AbandonChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, testGerritChangeID)
			return pr, nil
		})

		err := s.CloseChangeset(ctx, cs)
		assert.Nil(t, err)
		assertGerritChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestGerritSource_CreateComment(t *testing.T) {
	ctx := context.Background()

	t.Run("error creating comment", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		want := errors.New("error")
		client.WriteReviewCommentFunc.SetDefaultHook(func(ctx context.Context, changeID string, ci gerrit.ChangeReviewComment) error {
			assert.Equal(t, "comment", ci.Message)
			return want
		})

		err := s.CreateComment(ctx, cs, "comment")
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		client.WriteReviewCommentFunc.SetDefaultHook(func(ctx context.Context, changeID string, ci gerrit.ChangeReviewComment) error {
			assert.Equal(t, "comment", ci.Message)
			return nil
		})

		err := s.CreateComment(ctx, cs, "comment")
		assert.Nil(t, err)
	})
}

func TestGerritSource_MergeChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error merging pull request", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		want := errors.New("error")
		client.SubmitChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, testGerritChangeID, changeID)
			return &gerrit.Change{}, want
		})

		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		target := ChangesetNotMergeableError{}
		assert.ErrorAs(t, err, &target)
		assert.Equal(t, want.Error(), target.ErrorMsg)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		want := &notFoundError{}
		client.SubmitChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, testGerritChangeID, changeID)
			return &gerrit.Change{}, want
		})

		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success with squash", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		pr := mockGerritChange(&testProject)
		client.SubmitChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, testGerritChangeID, changeID)
			return pr, nil
		})

		err := s.MergeChangeset(ctx, cs, true)
		assert.Nil(t, err)
		assertGerritChangesetMatchesPullRequest(t, cs, pr)

	})

	t.Run("success with no squash", func(t *testing.T) {
		cs, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		pr := mockGerritChange(&testProject)
		client.SubmitChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, testGerritChangeID, changeID)
			return pr, nil
		})

		err := s.MergeChangeset(ctx, cs, false)
		assert.Nil(t, err)
		assertGerritChangesetMatchesPullRequest(t, cs, pr)

	})
}

func assertGerritChangesetMatchesPullRequest(t *testing.T, cs *Changeset, pr *gerrit.Change) {
	t.Helper()

	// We're not thoroughly testing setChangesetMetadata() et al in this
	// assertion, but we do want to ensure that the PR was used to populate
	// fields on the Changeset.
	assert.EqualValues(t, pr.ID, cs.ExternalID)
}

// mockGerritChangeset creates a plausible non-forked changeset, repo,
// and Gerrit specific repo.
func mockGerritChangeset() (*Changeset, *types.Repo) {
	repo := &types.Repo{Metadata: &testProject}
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

// mockGerritChange returns a plausible pull request that would be
// returned from Bitbucket Cloud for a non-forked changeset.
func mockGerritChange(project *gerrit.Project) *gerrit.Change {
	return &gerrit.Change{
		ID:      "42",
		Project: project.Name,
	}
}

func mockGerritSource() (*GerritSource, *MockGerritClient) {
	client := NewStrictMockGerritClient()
	s := &GerritSource{client: client}

	return s, client
}

package sources

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	gerritbatches "github.com/sourcegraph/sourcegraph/internal/batches/sources/gerrit"

	"github.com/sourcegraph/sourcegraph/internal/api"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	testGerritProjectName = "testrepo"
	testProject           = gerrit.Project{ID: "testrepoid", Name: testGerritProjectName}
	testChangeIDPrefix    = "ivarsano~targetbranch~"
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
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()
		want := errors.New("error")
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return &gerrit.Change{}, want
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return &gerrit.Change{}, &notFoundError{}
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		target := ChangesetNotFoundError{}
		assert.ErrorAs(t, err, &target)
		assert.Same(t, target.Changeset, cs)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		change := mockGerritChange(&testProject, id)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return change, nil
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

		err := s.LoadChangeset(ctx, cs)
		assert.Nil(t, err)
	})
}

func TestGerritSource_CreateChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, _, _ := mockGerritChangeset()
		s, client := mockGerritSource()
		testChangeID := GenerateGerritChangeID(*cs.Changeset)
		want := errors.New("error")
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, testChangeID)
			return &gerrit.Change{}, want
		})

		b, err := s.CreateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
		assert.False(t, b)
	})

	t.Run("change not found", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		s, client := mockGerritSource()
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return &gerrit.Change{}, &notFoundError{}
		})

		b, err := s.CreateChangeset(ctx, cs)
		assert.NotNil(t, err)
		target := ChangesetNotFoundError{}
		assert.ErrorAs(t, err, &target)
		assert.Same(t, target.Changeset, cs)
		assert.False(t, b)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		s, client := mockGerritSource()

		change := mockGerritChange(&testProject, id)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return change, nil
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

		b, err := s.CreateChangeset(ctx, cs)
		assert.Nil(t, err)
		assert.False(t, b)
	})
}

func TestGerritSource_CreateDraftChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error setting WIP", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		s, client := mockGerritSource()
		want := errors.New("error")
		client.SetWIPFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, changeID, id)
			return want
		})

		b, err := s.CreateDraftChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
		assert.False(t, b)
	})

	t.Run("change not found", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		s, client := mockGerritSource()
		client.SetWIPFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, changeID, id)
			return &notFoundError{}
		})

		b, err := s.CreateDraftChangeset(ctx, cs)
		assert.NotNil(t, err)
		target := ChangesetNotFoundError{}
		assert.ErrorAs(t, err, &target)
		assert.Same(t, target.Changeset, cs)
		assert.False(t, b)
	})

	t.Run("GetChange error", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		s, client := mockGerritSource()
		want := errors.New("error")
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.SetWIPFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, changeID, id)
			return nil
		})
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return &gerrit.Change{}, want
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

		b, err := s.CreateDraftChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
		assert.False(t, b)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		s, client := mockGerritSource()
		change := mockGerritChange(&testProject, id)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.SetWIPFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, changeID, id)
			return nil
		})
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return change, nil
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

		b, err := s.CreateDraftChangeset(ctx, cs)
		assert.Nil(t, err)
		assert.False(t, b)
	})
}

func TestGerritSource_UpdateChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("regular error getting change", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		cs.Metadata = &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				ID: testChangeIDPrefix + id,
			},
		}
		s, client := mockGerritSource()
		want := errors.New("error")
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return nil, want
		})

		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})
	t.Run("multiple changes, error when deleting change", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		cs.Metadata = &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				ID: testChangeIDPrefix + id,
			},
		}
		s, client := mockGerritSource()
		want := errors.New("error")
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return nil, gerrit.MultipleChangesError{}
		})
		client.DeleteChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, testChangeIDPrefix+id, changeID)
			return want
		})

		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})
	t.Run("multiple changes, error when setting change WIP", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		cs.Metadata = &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				ID:             testChangeIDPrefix + id,
				WorkInProgress: true,
			},
		}
		s, client := mockGerritSource()
		want := errors.New("error")
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return nil, gerrit.MultipleChangesError{}
		})
		client.DeleteChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, testChangeIDPrefix+id, changeID)
			return nil
		})
		client.SetWIPFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, id, changeID)
			return want
		})

		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})
	t.Run("multiple changes, error when loading change", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		cs.Metadata = &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				ID:             testChangeIDPrefix + id,
				WorkInProgress: true,
			},
		}
		s, client := mockGerritSource()
		want := errors.New("error")
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return nil, gerrit.MultipleChangesError{}
		})
		client.DeleteChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, testChangeIDPrefix+id, changeID)
			return nil
		})
		client.SetWIPFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, id, changeID)
			return nil
		})
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return nil, want
		})

		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})
	t.Run("multiple changes, success", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		cs.Metadata = &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				ID:             testChangeIDPrefix + id,
				WorkInProgress: true,
			},
		}
		change := mockGerritChange(&testProject, id)
		s, client := mockGerritSource()
		client.DeleteChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, testChangeIDPrefix+id, changeID)
			return nil
		})
		client.SetWIPFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, id, changeID)
			return nil
		})
		hook1 := func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return nil, gerrit.MultipleChangesError{}
		}
		hook2 := func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return change, nil
		}
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.GetChangeFunc.hooks = []func(ctx context.Context, changeID string) (*gerrit.Change, error){
			hook1,
			hook2,
		}

		err := s.UpdateChangeset(ctx, cs)
		assert.Nil(t, err)
	})
	t.Run("move target branch error", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		cs.Metadata = &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				ID: testChangeIDPrefix + id,
			},
		}
		change := mockGerritChange(&testProject, id)
		change.Branch = "diffbranch"
		s, client := mockGerritSource()
		want := errors.New("error")
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return change, nil
		})
		client.MoveChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string, payload gerrit.MoveChangePayload) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			assert.Equal(t, cs.BaseRef, payload.DestinationBranch)
			return nil, want
		})

		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})
	t.Run("set commit message error", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		cs.Metadata = &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				ID: testChangeIDPrefix + id,
			},
		}
		change := mockGerritChange(&testProject, id)
		ogChange := *change
		change.Branch = "diffbranch"
		change.Subject = "diffsubject"
		s, client := mockGerritSource()
		want := errors.New("error")
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return change, nil
		})
		client.MoveChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string, payload gerrit.MoveChangePayload) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			assert.Equal(t, cs.BaseRef, payload.DestinationBranch)
			return &ogChange, nil
		})
		client.SetCommitMessageFunc.SetDefaultHook(func(ctx context.Context, changeID string, payload gerrit.SetCommitMessagePayload) error {
			assert.Equal(t, id, changeID)
			assert.Equal(t, fmt.Sprintf("%s\n\nChange-Id: %s\n", cs.Title, cs.ExternalID), payload.Message)
			return want
		})

		err := s.UpdateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})
	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		cs.Metadata = &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				ID: testChangeIDPrefix + id,
			},
		}
		change := mockGerritChange(&testProject, id)
		ogChange := *change
		change.Branch = "diffbranch"
		change.Subject = "diffsubject"
		s, client := mockGerritSource()
		client.GetChangeFunc.SetDefaultReturn(change, nil)
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.MoveChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string, payload gerrit.MoveChangePayload) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			assert.Equal(t, cs.BaseRef, payload.DestinationBranch)
			return &ogChange, nil
		})
		client.SetCommitMessageFunc.SetDefaultHook(func(ctx context.Context, changeID string, payload gerrit.SetCommitMessagePayload) error {
			assert.Equal(t, id, changeID)
			assert.Equal(t, fmt.Sprintf("%s\n\nChange-Id: %s\n", cs.Title, cs.ExternalID), payload.Message)
			return nil
		})

		err := s.UpdateChangeset(ctx, cs)
		assert.Nil(t, err)
	})
}

func TestGerritSource_UndraftChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error setting ReadyForReview", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()
		want := errors.New("error")
		client.SetReadyForReviewFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, changeID, id)
			return want
		})

		err := s.UndraftChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("change not found", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()
		client.SetReadyForReviewFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, changeID, id)
			return &notFoundError{}
		})

		err := s.UndraftChangeset(ctx, cs)
		assert.NotNil(t, err)
		target := ChangesetNotFoundError{}
		assert.ErrorAs(t, err, &target)
		assert.Same(t, target.Changeset, cs)
	})

	t.Run("GetChange error", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()
		want := errors.New("error")

		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.SetReadyForReviewFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, changeID, id)
			return nil
		})
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return &gerrit.Change{}, want
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

		err := s.UndraftChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		change := mockGerritChange(&testProject, id)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.SetReadyForReviewFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, changeID, id)
			return nil
		})
		client.GetChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return change, nil
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

		err := s.UndraftChangeset(ctx, cs)
		assert.Nil(t, err)
	})
}

func TestGerritSource_CloseChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error declining pull request", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		want := errors.New("error")
		client.AbandonChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return &gerrit.Change{}, want
		})

		err := s.CloseChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		pr := mockGerritChange(&testProject, id)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.AbandonChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return pr, nil
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

		assert.Len(t, client.DeleteChangeFunc.History(), 0)

		err := s.CloseChangeset(ctx, cs)

		assert.Nil(t, err)
		assert.Len(t, client.DeleteChangeFunc.History(), 0)
		assertGerritChangesetMatchesPullRequest(t, cs, pr)
	})

	t.Run("with auto-delete branch enabled, failure", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				BatchChangesAutoDeleteBranch: true,
			},
		})
		defer conf.Mock(nil)

		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		want := errors.New("error")
		client.AbandonChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return &gerrit.Change{}, want
		})

		assert.Len(t, client.DeleteChangeFunc.History(), 0)

		err := s.CloseChangeset(ctx, cs)

		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
		assert.Len(t, client.DeleteChangeFunc.History(), 0)
	})

	t.Run("with auto-delete branch enabled, success", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				BatchChangesAutoDeleteBranch: true,
			},
		})
		defer conf.Mock(nil)

		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		pr := mockGerritChange(&testProject, id)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.AbandonChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, changeID, id)
			return pr, nil
		})
		client.DeleteChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) error {
			assert.Equal(t, changeID, id)
			return nil
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

		assert.Len(t, client.DeleteChangeFunc.History(), 0)

		err := s.CloseChangeset(ctx, cs)

		assert.Nil(t, err)
		assert.Len(t, client.DeleteChangeFunc.History(), 1)
		assertGerritChangesetMatchesPullRequest(t, cs, pr)
	})
}

func TestGerritSource_CreateComment(t *testing.T) {
	ctx := context.Background()

	t.Run("error creating comment", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		want := errors.New("error")
		client.WriteReviewCommentFunc.SetDefaultHook(func(ctx context.Context, changeID string, ci gerrit.ChangeReviewComment) error {
			assert.Equal(t, changeID, id)
			assert.Equal(t, "comment", ci.Message)
			return want
		})

		err := s.CreateComment(ctx, cs, "comment")
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		client.WriteReviewCommentFunc.SetDefaultHook(func(ctx context.Context, changeID string, ci gerrit.ChangeReviewComment) error {
			assert.Equal(t, changeID, id)
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
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		want := errors.New("error")
		client.SubmitChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return &gerrit.Change{}, want
		})

		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		target := ChangesetNotMergeableError{}
		assert.ErrorAs(t, err, &target)
		assert.Equal(t, want.Error(), target.ErrorMsg)
	})

	t.Run("change not found", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		want := &notFoundError{}
		client.SubmitChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return &gerrit.Change{}, want
		})

		err := s.MergeChangeset(ctx, cs, false)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success with squash", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		pr := mockGerritChange(&testProject, id)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.SubmitChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return pr, nil
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

		err := s.MergeChangeset(ctx, cs, true)
		assert.Nil(t, err)
		assertGerritChangesetMatchesPullRequest(t, cs, pr)

	})

	t.Run("success with no squash", func(t *testing.T) {
		cs, id, _ := mockGerritChangeset()
		cs.ExternalID = id
		s, client := mockGerritSource()

		pr := mockGerritChange(&testProject, id)
		client.GetURLFunc.SetDefaultReturn(&url.URL{})
		client.SubmitChangeFunc.SetDefaultHook(func(ctx context.Context, changeID string) (*gerrit.Change, error) {
			assert.Equal(t, id, changeID)
			return pr, nil
		})
		client.GetChangeReviewsFunc.SetDefaultReturn(&[]gerrit.Reviewer{}, nil)

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
	assert.EqualValues(t, pr.ChangeID, cs.ExternalID)
}

// mockGerritChangeset creates a plausible non-forked changeset, repo,
// and Gerrit specific repo.
func mockGerritChangeset() (cs *Changeset, id string, repo *types.Repo) {
	repo = &types.Repo{Metadata: &testProject}
	cs = &Changeset{
		Title:      "title",
		Body:       "description",
		Changeset:  &btypes.Changeset{},
		RemoteRepo: repo,
		TargetRepo: repo,
		BaseRef:    "refs/heads/targetbranch",
	}
	id = GenerateGerritChangeID(*cs.Changeset)

	return cs, id, repo
}

// mockGerritChange returns a plausible pull request that would be
// returned from Bitbucket Cloud for a non-forked changeset.
func mockGerritChange(project *gerrit.Project, id string) *gerrit.Change {
	return &gerrit.Change{
		ChangeID: id,
		Project:  project.Name,
	}
}

func mockGerritSource() (*GerritSource, *MockGerritClient) {
	client := NewStrictMockGerritClient()
	s := &GerritSource{client: client}

	return s, client
}

package sources

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/stretchr/testify/assert"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	testPerforceProjectName = "testdepot"
	testPerforceChangeID    = "111"
)

func TestPerforceSource_ValidateAuthenticator(t *testing.T) {
	ctx := context.Background()

	for name, want := range map[string]error{
		"nil":   nil,
		"error": errors.New("error"),
	} {
		t.Run(name, func(t *testing.T) {
			s, client := mockPerforceSource()
			client.P4ExecFunc.SetDefaultReturn(fakeCloser{}, http.Header{}, want)
			assert.Equal(t, want, s.ValidateAuthenticator(ctx))
		})
	}
}

func TestPerforceSource_LoadChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error getting changelist", func(t *testing.T) {
		cs, _ := mockPerforceChangeset()
		s, client := mockPerforceSource()
		want := errors.New("error")
		client.P4GetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string) (protocol.PerforceChangelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			return protocol.PerforceChangelist{}, want
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockPerforceChangeset()
		s, client := mockPerforceSource()

		change := mockPerforceChange()
		client.P4GetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string) (protocol.PerforceChangelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			return *change, nil
		})

		err := s.LoadChangeset(ctx, cs)
		assert.Nil(t, err)
	})
}

func TestPerforceSource_CreateChangeset(t *testing.T) {
	ctx := context.Background()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, _ := mockPerforceChangeset()
		s, client := mockPerforceSource()
		want := errors.New("error")
		client.P4GetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string) (protocol.PerforceChangelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			return protocol.PerforceChangelist{}, want
		})

		b, err := s.CreateChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
		assert.False(t, b)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockPerforceChangeset()
		s, client := mockPerforceSource()

		change := mockPerforceChange()
		client.P4GetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string) (protocol.PerforceChangelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			return *change, nil
		})

		b, err := s.CreateChangeset(ctx, cs)
		assert.Nil(t, err)
		assert.False(t, b)
	})
}

func assertPerforceChangesetMatchesPullRequest(t *testing.T, cs *Changeset, pr *protocol.PerforceChangelist) {
	t.Helper()

	// We're not thoroughly testing setChangesetMetadata() et al in this
	// assertion, but we do want to ensure that the PR was used to populate
	// fields on the Changeset.
	assert.EqualValues(t, pr.ID, cs.ExternalID)
}

// mockPerforceChangeset creates a plausible non-forked changeset, repo,
// and Perforce specific repo.
func mockPerforceChangeset() (*Changeset, *types.Repo) {
	repo := &types.Repo{Metadata: &testProject}
	cs := &Changeset{
		Title:      "title",
		Body:       "description",
		Changeset:  &btypes.Changeset{},
		RemoteRepo: repo,
		TargetRepo: repo,
		BaseRef:    "refs/heads/targetbranch",
	}

	cs.Changeset.ExternalID = testPerforceChangeID

	return cs, repo
}

// mockPerforceChange returns a plausible changelist that would be
// returned from Perforce.
func mockPerforceChange() *protocol.PerforceChangelist {
	return &protocol.PerforceChangelist{
		ID:     testPerforceChangeID,
		Author: "Peter Guy",
		State:  protocol.PerforceChangelistStatePending,
	}
}

func mockPerforceSource() (*PerforceSource, *MockGitserverClient) {
	client := NewStrictMockGitserverClient()
	auther := auth.BasicAuth{Username: "user", Password: "pass"}
	s := &PerforceSource{gitServerClient: client, auther: &auther}
	return s, client
}

type fakeCloser struct {
	io.Reader
}

func (fakeCloser) Close() error { return nil }

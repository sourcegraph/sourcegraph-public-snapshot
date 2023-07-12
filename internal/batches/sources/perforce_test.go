package sources

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/schema"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	testPerforceProjectName = "testdepot"
	testPerforceChangeID    = "111"
	testPerforceCredentials = gitserver.PerforceCredentials{Username: "user", Password: "pass", Host: "https://perforce.sgdev.org:1666"}
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
		client.P4GetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string, credentials gitserver.PerforceCredentials) (*protocol.PerforceChangelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			assert.Equal(t, testPerforceCredentials, credentials)
			return new(protocol.PerforceChangelist), want
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockPerforceChangeset()
		s, client := mockPerforceSource()

		change := mockPerforceChange()
		client.P4GetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string, credentials gitserver.PerforceCredentials) (*protocol.PerforceChangelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			assert.Equal(t, testPerforceCredentials, credentials)
			return change, nil
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
		client.P4GetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string, credentials gitserver.PerforceCredentials) (*protocol.PerforceChangelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			assert.Equal(t, testPerforceCredentials, credentials)
			return new(protocol.PerforceChangelist), want
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
		client.P4GetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string, credentials gitserver.PerforceCredentials) (*protocol.PerforceChangelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			assert.Equal(t, testPerforceCredentials, credentials)
			return change, nil
		})

		b, err := s.CreateChangeset(ctx, cs)
		assert.Nil(t, err)
		assert.False(t, b)
	})
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
	s := &PerforceSource{gitServerClient: client, perforceCreds: &testPerforceCredentials, server: schema.PerforceConnection{P4Port: "https://perforce.sgdev.org:1666"}}
	return s, client
}

type fakeCloser struct {
	io.Reader
}

func (fakeCloser) Close() error { return nil }

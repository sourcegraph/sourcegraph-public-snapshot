package sources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	v1 "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/schema"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	testPerforceChangeID    = "111"
	testPerforceCredentials = gitserver.PerforceCredentials{Username: "user", Password: "pass", Host: "perforce.sgdev.org:1666"}
)

func TestPerforceSource_ValidateAuthenticator(t *testing.T) {
	ctx := context.Background()

	for name, want := range map[string]error{
		"nil":   nil,
		"error": errors.New("error"),
	} {
		t.Run(name, func(t *testing.T) {
			s, client := mockPerforceSource()
			client.CheckPerforceCredentialsFunc.PushReturn(want)
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
		client.PerforceGetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string, credentials gitserver.PerforceCredentials) (*perforce.Changelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			assert.Equal(t, testPerforceCredentials, credentials)
			return new(perforce.Changelist), want
		})

		err := s.LoadChangeset(ctx, cs)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, want)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockPerforceChangeset()
		s, client := mockPerforceSource()

		change := mockPerforceChange()
		client.P4GetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string, credentials gitserver.PerforceCredentials) (*perforce.Changelist, error) {
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
		client.PerforceGetChangelistFunc.SetDefaultHook(func(ctx context.Context, conn *v1.PerforceConnectionDetails, changeID string) (*perforce.Changelist, error) {
			assert.Equal(t, changeID, testPerforceChangeID)
			assert.Equal(t, testPerforceCredentials, credentials)
			return new(perforce.Changelist), want
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
		client.PerforceGetChangelistFunc.SetDefaultHook(func(ctx context.Context, changeID string, credentials gitserver.PerforceCredentials) (*perforce.Changelist, error) {
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
func mockPerforceChange() *perforce.Changelist {
	return &perforce.Changelist{
		ID:     testPerforceChangeID,
		Author: "Peter Guy",
		State:  perforce.ChangelistStatePending,
	}
}

func mockPerforceSource() (*PerforceSource, *MockGitserverClient) {
	client := NewStrictMockGitserverClient()
	// Cred checks should pass by default.
	client.CheckPerforceCredentialsFunc.SetDefaultReturn(nil)
	s := &PerforceSource{gitServerClient: client, perforceCreds: &testPerforceCredentials, conn: schema.PerforceConnection{P4Port: "perforce.sgdev.org:1666"}}
	return s, client
}

package gitlaboauth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSessionIssuerHelper_CreateCodeHostConnection(t *testing.T) {
	createCodeHostConnectionHelper(t, false)
}

func TestSessionIssuerHelper_CreateCodeHostConnectionHandlesExistingService(t *testing.T) {
	createCodeHostConnectionHelper(t, true)
}

func createCodeHostConnectionHelper(t *testing.T, serviceExists bool) {
	t.Helper()

	ctx := context.Background()
	db := database.NewMockDB()
	s := &sessionIssuerHelper{db: db}
	t.Run("Unauthenticated request", func(t *testing.T) {
		_, _, err := s.CreateCodeHostConnection(ctx, nil, "")
		assert.Error(t, err)
	})

	mockGitLabCom := newMockProvider(t, db, "gitlabcomclient", "gitlabcomsecret", "https://gitlab.com/")
	providers.MockProviders = []providers.Provider{mockGitLabCom.Provider}
	defer func() { providers.MockProviders = nil }()

	tok := &oauth2.Token{
		AccessToken: "dummy-value-that-isnt-relevant-to-unit-correctness",
	}
	act := &actor.Actor{UID: 1}
	glUser := &gitlab.User{
		ID:       101,
		Username: "alice",
	}

	ctx = actor.WithActor(ctx, act)
	ctx = WithUser(ctx, glUser)
	now := time.Now()

	externalServices := database.NewMockExternalServiceStore()
	externalServices.TransactFunc.SetDefaultReturn(externalServices, nil)
	externalServices.ListFunc.SetDefaultHook(func(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		if !serviceExists {
			return nil, nil
		}
		return []*types.ExternalService{
			{
				Kind:        extsvc.KindGitLab,
				DisplayName: fmt.Sprintf("GitLab (%s)", glUser.Username),
				Config: fmt.Sprintf(`
{
  "url": "%s",
  "token": "%s",
  "token.type": "oauth",
  "projectQuery": ["projects?id_before=0"]
}
`, mockGitLabCom.ServiceID, "a-token-that-should-be-replaced"),
				NamespaceUserID: act.UID,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		}, nil
	})
	var got *types.ExternalService
	externalServices.CreateFunc.SetDefaultHook(func(ctx context.Context, confGet func() *conf.Unified, es *types.ExternalService) error {
		got = es
		return nil
	})
	externalServices.UpsertFunc.SetDefaultHook(func(ctx context.Context, svcs ...*types.ExternalService) error {
		require.Len(t, svcs, 1)

		// Tweak timestamps
		svcs[0].CreatedAt = now
		svcs[0].UpdatedAt = now
		got = svcs[0]
		return nil
	})
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	fromCreation, _, err := s.CreateCodeHostConnection(ctx, tok, mockGitLabCom.ConfigID().ID)
	require.NoError(t, err)

	want := &types.ExternalService{
		Kind:        extsvc.KindGitLab,
		DisplayName: fmt.Sprintf("GitLab (%s)", glUser.Username),
		Config: fmt.Sprintf(`
{
  "url": "%s",
  "token": "%s",
  "token.type": "oauth",
  "projectQuery": ["projects?id_before=0"]
}
`, mockGitLabCom.ServiceID, tok.AccessToken),
		NamespaceUserID: act.UID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	assert.Equal(t, want, got)
	assert.Equal(t, want, fromCreation)
}

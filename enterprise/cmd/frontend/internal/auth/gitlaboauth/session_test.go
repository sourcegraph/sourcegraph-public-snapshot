package gitlaboauth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_CreateCodeHostConnection(t *testing.T) {
	createCodeHostConnectionHelper(t, false)
}

func Test_CreateCodeHostConnectionHandlesExistingService(t *testing.T) {
	createCodeHostConnectionHelper(t, true)
}

func createCodeHostConnectionHelper(t *testing.T, serviceExists bool) {
	t.Helper()

	ctx := context.Background()
	s := &sessionIssuerHelper{}
	t.Run("Unauthenticated request", func(t *testing.T) {
		_, err := s.CreateCodeHostConnection(ctx, nil, "")
		if err == nil {
			t.Fatal("Want error but got nil")
		}
	})

	mockGitLabCom := newMockProvider(t, "gitlabcomclient", "gitlabcomsecret", "https://gitlab.com/")
	providers.MockProviders = []providers.Provider{mockGitLabCom.Provider}
	defer func() { providers.MockProviders = nil }()

	var got *types.ExternalService
	database.Mocks.ExternalServices.Create = func(ctx context.Context, confGet func() *conf.Unified, externalService *types.ExternalService) error {
		got = externalService
		return nil
	}
	defer func() { database.Mocks.ExternalServices.Create = nil }()

	act := &actor.Actor{UID: 1}
	glUser := &gitlab.User{
		ID:       101,
		Username: "alice",
	}

	now := time.Now()
	ctx = actor.WithActor(ctx, act)
	ctx = WithUser(ctx, glUser)
	tok := &oauth2.Token{
		AccessToken: "dummy-value-that-isnt-relevant-to-unit-correctness",
	}

	database.Mocks.ExternalServices.Transact = func(ctx context.Context) (*database.ExternalServiceStore, error) {
		return database.GlobalExternalServices, nil
	}
	database.Mocks.ExternalServices.Done = func(err error) error {
		return nil
	}
	database.Mocks.ExternalServices.List = func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
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
	}
	database.Mocks.ExternalServices.Upsert = func(ctx context.Context, services ...*types.ExternalService) error {
		if len(services) != 1 {
			t.Fatalf("Expected 1 service in Upsert, got %d", len(services))
		}
		// Tweak timestamps
		services[0].CreatedAt = now
		services[0].UpdatedAt = now
		got = services[0]
		return nil
	}
	t.Cleanup(func() {
		database.Mocks.ExternalServices = database.MockExternalServices{}
	})

	_, err := s.CreateCodeHostConnection(ctx, tok, mockGitLabCom.ConfigID().ID)
	if err != nil {
		t.Fatal(err)
	}

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
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

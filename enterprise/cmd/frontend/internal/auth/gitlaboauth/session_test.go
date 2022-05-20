package gitlaboauth

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSessionIssuerHelper_CreateCodeHostConnection(t *testing.T) {
	createCodeHostConnectionHelper(t, false)
}

func TestSessionIssuerHelper_GetOrCreateUser(t *testing.T) {
	glURL, _ := url.Parse("https://gitlab.com")
	codeHost := extsvc.NewCodeHost(glURL, extsvc.TypeGitLab)
	clientID := "client-id"

	var expErr bool
	authSaveableUsers := map[string]int32{
		"alice": 1,
	}

	glUser := &gitlab.User{ID: int32(101), Username: string("alice"), Email: string("alice@example.com")}
	expActor := &actor.Actor{UID: 1}
	expAuthUserOp := &auth.GetAndSaveUserOp{
		UserProps:        u("alice", "alice@example.com", true),
		ExternalAccount:  acct(extsvc.TypeGitLab, "https://gitlab.com/", clientID, "101"),
		CreateIfNotExist: true,
	}

	t.Run("gitlab signin", func(t *testing.T) {
		var gotAuthUserOp *auth.GetAndSaveUserOp
		auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (userID int32, safeErrMsg string, err error) {
			if gotAuthUserOp != nil {
				t.Fatal("GetAndSaveUser called more than once")
			}

			op.ExternalAccountData = extsvc.AccountData{}
			gotAuthUserOp = &op

			if uid, ok := authSaveableUsers[op.UserProps.Username]; ok {
				return uid, "", nil
			}

			return 0, "safeErr", errors.New("error mocking get and save user")
		}

		defer func() {
			auth.MockGetAndSaveUser = nil
		}()

		ctx := context.Background()
		ctx = actor.WithActor(ctx, expActor)
		ctx = WithUser(ctx, glUser)

		s := &sessionIssuerHelper{
			CodeHost: codeHost,
			clientID: clientID,
		}

		tok := &oauth2.Token{AccessToken: "dummy-value-that-isnt-relevant-to-unit-correctness"}
		actr, _, err := s.GetOrCreateUser(ctx, tok, "", "", "")

		if got, exp := actr, expActor; !reflect.DeepEqual(got, exp) {
			t.Errorf("expected actor %v, got %v", exp, got)
		}

		if expErr && err == nil {
			t.Errorf("expected err %v, but was nil", expErr)
		} else if !expErr && err != nil {
		}

		if got, exp := gotAuthUserOp, expAuthUserOp; !reflect.DeepEqual(got, exp) {
			t.Error(cmp.Diff(got, exp))
		}
	})

}

func u(username, email string, emailIsVerified bool) database.NewUser {
	return database.NewUser{
		Username:        username,
		Email:           email,
		EmailIsVerified: emailIsVerified,
	}
}

func acct(serviceType, serviceID, clientID, accountID string) extsvc.AccountSpec {
	return extsvc.AccountSpec{
		ServiceType: serviceType,
		ServiceID:   serviceID,
		ClientID:    clientID,
		AccountID:   accountID,
	}
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

package graphqlbackend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAddExternalService(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
		users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
		users.TagsFunc.SetDefaultReturn(map[string]bool{}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := newSchemaResolver(db, gitserver.NewClient(db)).AddExternalService(ctx, &addExternalServiceArgs{})
		if want := auth.ErrMustBeSiteAdmin; err != want {
			t.Errorf("err: want %q but got %q", want, err)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.CreateFunc.SetDefaultReturn(nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
			mutation {
				addExternalService(input: {
					kind: GITHUB,
					displayName: "GITHUB #1",
					config: "{\"url\": \"https://github.com\", \"repositoryQuery\": [\"none\"], \"token\": \"abc\"}"
				}) {
					kind
					displayName
					config
				}
			}
		`,
			ExpectedResult: `
			{
				"addExternalService": {
					"kind": "GITHUB",
					"displayName": "GITHUB #1",
					"config":"{\n  \"url\": \"https://github.com\",\n  \"repositoryQuery\": [\n    \"none\"\n  ],\n  \"token\": \"` + types.RedactedSecret + `\"\n}"
				}
			}
		`,
		},
	})
}

func TestUpdateExternalService(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		t.Run("cannot update external services", func(t *testing.T) {
			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := newSchemaResolver(db, nil).UpdateExternalService(ctx, &updateExternalServiceArgs{
				Input: updateExternalServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				},
			})
			if want := auth.ErrMustBeSiteAdmin; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})
	})

	t.Run("empty config", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

		externalServices := database.NewMockExternalServiceStore()
		externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
			return &types.ExternalService{
				ID:     id,
				Config: extsvc.NewEmptyConfig(),
			}, nil
		})

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.ExternalServicesFunc.SetDefaultReturn(externalServices)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		result, err := newSchemaResolver(db, gitserver.NewClient(db)).UpdateExternalService(ctx, &updateExternalServiceArgs{
			Input: updateExternalServiceInput{
				ID:     "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
				Config: strptr(""),
			},
		})
		gotErr := fmt.Sprintf("%v", err)
		wantErr := "blank external service configuration is invalid (must be valid JSONC)"
		if gotErr != wantErr {
			t.Errorf("err: want %q but got %q", wantErr, gotErr)
		}
		if result != nil {
			t.Errorf("result: want nil but got %v", result)
		}
	})

	var cachedUpdate *database.ExternalServiceUpdate

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.UpdateFunc.SetDefaultHook(func(ctx context.Context, ps []schema.AuthProviders, id int64, update *database.ExternalServiceUpdate) error {
		cachedUpdate = update
		return nil
	})
	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		if cachedUpdate == nil {
			return &types.ExternalService{
				ID:     id,
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewEmptyConfig(),
			}, nil
		}
		return &types.ExternalService{
			ID:          id,
			Kind:        extsvc.KindGitHub,
			DisplayName: *cachedUpdate.DisplayName,
			Config:      extsvc.NewUnencryptedConfig(*cachedUpdate.Config),
		}, nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
			mutation {
				updateExternalService(input: {
					id: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
					displayName: "GITHUB #2",
					config: "{\"url\": \"https://github.com\", \"repositoryQuery\": [\"none\"], \"token\": \"def\"}"
				}) {
					displayName
					config
				}
			}
		`,
		ExpectedResult: `
			{
				"updateExternalService": {
				  "displayName": "GITHUB #2",
				  "config":"{\n  \"url\": \"https://github.com\",\n  \"repositoryQuery\": [\n    \"none\"\n  ],\n  \"token\": \"` + types.RedactedSecret + `\"\n}"

				}
			}
		`,
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
	})
}

func TestExcludeRepoFromExternalService_ExternalServiceDoesntSupportRepoExclusion(t *testing.T) {
	var cachedUpdate *database.ExternalServiceUpdate
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.UpdateFunc.SetDefaultHook(func(ctx context.Context, ps []schema.AuthProviders, id int64, update *database.ExternalServiceUpdate) error {
		cachedUpdate = update
		return nil
	})
	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:     id,
			Kind:   extsvc.KindGerrit,
			Config: extsvc.NewEmptyConfig(),
		}, nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
			mutation {
				excludeRepoFromExternalService(
					externalService: "RXh0ZXJuYWxTZXJ2aWNlOjI=",
					repo: "UmVwb3NpdG9yeTox"
				) {
					alwaysNil
				}
			}
		`,
		ExpectedResult: `
			{
				"excludeRepoFromExternalService": {
					"alwaysNil": null
				}
			}
		`,
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
	})

	assert.Nil(t, cachedUpdate)
}

func TestExcludeRepoFromExternalService_NoExistingExcludedRepos_NewExcludedRepoAdded(t *testing.T) {
	var cachedUpdate *database.ExternalServiceUpdate

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.UpdateFunc.SetDefaultHook(func(ctx context.Context, ps []schema.AuthProviders, id int64, update *database.ExternalServiceUpdate) error {
		cachedUpdate = update
		return nil
	})
	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:     id,
			Kind:   extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(`{"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`),
		}, nil
	})

	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: api.RepoID(1), Name: "sourcegraph/sourcegraph"}, nil
	})
	mockSyncExternalService = func(_ context.Context, _ *syncExternalServiceArgs) (*EmptyResponse, error) {
		return &EmptyResponse{}, nil
	}
	t.Cleanup(func() { mockSyncExternalService = nil })

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.ReposFunc.SetDefaultReturn(repos)

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Label:  "ExcludeRepoFromExternalService. Empty exclude. Repo exclusion added.",
		Query: `
			mutation {
				excludeRepoFromExternalService(
					externalService: "RXh0ZXJuYWxTZXJ2aWNlOjE=",
					repo: "UmVwb3NpdG9yeTox"
				) {
					alwaysNil
				}
			}
		`,
		ExpectedResult: `
			{
				"excludeRepoFromExternalService": {
					"alwaysNil": null
				}
			}
		`,
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
	})

	expectedConfig := `{"exclude":[{"id":"1","name":"sourcegraph/sourcegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`
	assert.Equal(t, expectedConfig, *cachedUpdate.Config)
}

func TestExcludeRepoFromExternalService_ExcludedRepoExists_AnotherExcludedRepoAdded(t *testing.T) {
	var cachedUpdate *database.ExternalServiceUpdate

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.UpdateFunc.SetDefaultHook(func(ctx context.Context, ps []schema.AuthProviders, id int64, update *database.ExternalServiceUpdate) error {
		cachedUpdate = update
		return nil
	})
	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:     id,
			Kind:   extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(`{"exclude":[{"id":"1","name":"sourcegraph/sourcegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`),
		}, nil
	})

	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: api.RepoID(2), Name: "sourcegraph/horsegraph"}, nil
	})
	mockSyncExternalService = func(_ context.Context, _ *syncExternalServiceArgs) (*EmptyResponse, error) {
		return &EmptyResponse{}, nil
	}
	t.Cleanup(func() { mockSyncExternalService = nil })

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.ReposFunc.SetDefaultReturn(repos)

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
			mutation {
				excludeRepoFromExternalService(
					externalService: "RXh0ZXJuYWxTZXJ2aWNlOjE=",
					repo: "UmVwb3NpdG9yeToy"
				) {
					alwaysNil
				}
			}
		`,
		ExpectedResult: `
			{
				"excludeRepoFromExternalService": {
					"alwaysNil": null
				}
			}
		`,
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
	})

	expectedConfig := `{"exclude":[{"id":"1","name":"sourcegraph/sourcegraph"},{"id":"2","name":"sourcegraph/horsegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`
	assert.Equal(t, expectedConfig, *cachedUpdate.Config)
}

func TestExcludeRepoFromExternalService_ExcludedRepoExists_SameRepoIsNotExcludedAgain(t *testing.T) {
	var cachedUpdate *database.ExternalServiceUpdate

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.UpdateFunc.SetDefaultHook(func(ctx context.Context, ps []schema.AuthProviders, id int64, update *database.ExternalServiceUpdate) error {
		cachedUpdate = update
		return nil
	})
	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:     id,
			Kind:   extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(`{"exclude":[{"id":"1","name":"sourcegraph/sourcegraph"},{"id":"2","name":"sourcegraph/horsegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`),
		}, nil
	})

	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(_ context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: api.RepoID(2), Name: "sourcegraph/horsegraph"}, nil
	})
	mockSyncExternalService = func(_ context.Context, _ *syncExternalServiceArgs) (*EmptyResponse, error) {
		return &EmptyResponse{}, nil
	}
	t.Cleanup(func() { mockSyncExternalService = nil })

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.ReposFunc.SetDefaultReturn(repos)

	RunTest(t, &Test{
		Schema: mustParseGraphQLSchema(t, db),
		Query: `
			mutation {
				excludeRepoFromExternalService(
					externalService: "RXh0ZXJuYWxTZXJ2aWNlOjE=",
					repo: "UmVwb3NpdG9yeToy"
				) {
					alwaysNil
				}
			}
		`,
		ExpectedResult: `
			{
				"excludeRepoFromExternalService": {
					"alwaysNil": null
				}
			}
		`,
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
	})

	expectedConfig := `{"exclude":[{"id":"1","name":"sourcegraph/sourcegraph"},{"id":"2","name":"sourcegraph/horsegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`
	assert.Equal(t, expectedConfig, *cachedUpdate.Config)
}

func TestAddRepoToExclude(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		kind           string
		initialConfig  string
		expectedConfig string
	}{
		{
			name:           "second attempt of excluding same repo is ignored for AWSCodeCommit schema",
			kind:           extsvc.KindAWSCodeCommit,
			initialConfig:  `{"accessKeyID":"accessKeyID","gitCredentials":{"password":"","username":""},"region":"","secretAccessKey":""}`,
			expectedConfig: `{"accessKeyID":"accessKeyID","exclude":[{"id":"1","name":"sourcegraph/sourcegraph"}],"gitCredentials":{"password":"","username":""},"region":"","secretAccessKey":""}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for BitbucketCloud schema",
			kind:           extsvc.KindBitbucketCloud,
			initialConfig:  `{"appPassword":"","url":"https://bitbucket.org","username":""}`,
			expectedConfig: `{"appPassword":"","exclude":[{"name":"sourcegraph/sourcegraph"}],"url":"https://bitbucket.org","username":""}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for BitbucketServer schema",
			kind:           extsvc.KindBitbucketServer,
			initialConfig:  `{"repositoryQuery":["none"],"token":"abc","url":"https://bitbucket.sg.org","username":""}`,
			expectedConfig: `{"exclude":[{"id":1,"name":"sourcegraph/sourcegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://bitbucket.sg.org","username":""}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for GitHub schema",
			kind:           extsvc.KindGitHub,
			initialConfig:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			expectedConfig: `{"exclude":[{"id":"1","name":"sourcegraph/sourcegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for GitLab schema",
			kind:           extsvc.KindGitLab,
			initialConfig:  `{"projectQuery":null,"token":"abc","url":"https://gitlab.com"}`,
			expectedConfig: `{"exclude":[{"name":"sourcegraph/sourcegraph"}],"projectQuery":null,"token":"abc","url":"https://gitlab.com"}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for Gitolite schema",
			kind:           extsvc.KindGitolite,
			initialConfig:  `{"host":"gitolite.com","prefix":""}`,
			expectedConfig: `{"exclude":[{"name":"sourcegraph/sourcegraph"}],"host":"gitolite.com","prefix":""}`,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			extSvc := &types.ExternalService{
				Kind:        test.kind,
				DisplayName: fmt.Sprintf("%s #1", test.kind),
				Config:      extsvc.NewUnencryptedConfig(test.initialConfig),
			}
			actualConfig, err := addRepoToExclude(ctx, extSvc, &types.Repo{ID: api.RepoID(1), Name: "sourcegraph/sourcegraph"})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, test.expectedConfig, actualConfig)

			actualConfig, err = addRepoToExclude(ctx, extSvc, &types.Repo{ID: api.RepoID(1), Name: "sourcegraph/sourcegraph"})
			if err != nil {
				t.Fatal(err)
			}
			// Config shouldn't have been changed.
			assert.Equal(t, test.expectedConfig, actualConfig)
		})
	}
}

func TestDeleteExternalService(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		t.Run("cannot delete external services", func(t *testing.T) {
			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
			result, err := newSchemaResolver(db, gitserver.NewClient(db)).DeleteExternalService(ctx, &deleteExternalServiceArgs{
				ExternalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
			})
			if want := auth.ErrMustBeSiteAdmin; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})
	})

	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	externalServices := database.NewMockExternalServiceStore()
	externalServices.DeleteFunc.SetDefaultReturn(nil)
	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:     id,
			Config: extsvc.NewEmptyConfig(),
		}, nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
			mutation {
				deleteExternalService(externalService: "RXh0ZXJuYWxTZXJ2aWNlOjQ=") {
					alwaysNil
				}
			}
		`,
			ExpectedResult: `
			{
				"deleteExternalService": {
					"alwaysNil": null
				}
			}
		`,
			Context: actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
		},
	})
}

func TestExternalServicesResolver(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		t.Run("cannot read site-level external services", func(t *testing.T) {
			users := database.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
			users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			result, err := newSchemaResolver(db, gitserver.NewClient(db)).ExternalServices(context.Background(), &ExternalServicesArgs{})
			if want := auth.ErrMustBeSiteAdmin; err != want {
				t.Errorf("err: want %q but got %v", want, err)
			}
			if result != nil {
				t.Errorf("result: want nil but got %v", result)
			}
		})
	})

	t.Run("authenticated as admin", func(t *testing.T) {
		t.Run("can read site-level external service", func(t *testing.T) {
			users := database.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id, SiteAdmin: true}, nil
			})

			db := database.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)

			_, err := newSchemaResolver(db, gitserver.NewClient(db)).ExternalServices(context.Background(), &ExternalServicesArgs{})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}

func TestExternalServices(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultHook(func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		if opt.AfterID > 0 {
			return []*types.ExternalService{
				{ID: 4, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindAWSCodeCommit},
				{ID: 5, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindGerrit},
			}, nil
		}

		ess := []*types.ExternalService{
			{ID: 1, Config: extsvc.NewEmptyConfig()},
			{ID: 2, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindGitHub},
			{ID: 3, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindGitHub},
			{ID: 4, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindAWSCodeCommit},
			{ID: 5, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindGerrit},
		}
		if opt.LimitOffset != nil {
			return ess[:opt.LimitOffset.Limit], nil
		}
		return ess, nil
	})
	externalServices.CountFunc.SetDefaultHook(func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error) {
		if opt.AfterID > 0 {
			return 1, nil
		}

		return 2, nil
	})
	externalServices.GetLastSyncErrorFunc.SetDefaultReturn("Oops", nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	mockLastCheckedAt := time.Now()
	mockCheckConnection = func(ctx context.Context, r *externalServiceResolver) (*externalServiceResolver, error) {
		switch r.externalService.ID {
		case 2:
			r.availability.unavailable = &externalServiceUnavailable{
				suspectedReason: "failed to connect",
			}
		case 3:
			r.availability.available = &externalServiceAvailable{
				lastCheckedAt: mockLastCheckedAt,
			}
		default:
			r.availability.unknown = &externalServiceUnknown{}
		}

		return r, nil
	}

	// NOTE: all these tests are run as site admin
	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Label:  "Read all external services",
			Query: `
			{
				externalServices() {
					nodes {
						id
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes": [
						{"id":"RXh0ZXJuYWxTZXJ2aWNlOjE="},
						{"id":"RXh0ZXJuYWxTZXJ2aWNlOjI="},
						{"id":"RXh0ZXJuYWxTZXJ2aWNlOjM="},
						{"id":"RXh0ZXJuYWxTZXJ2aWNlOjQ="},
						{"id":"RXh0ZXJuYWxTZXJ2aWNlOjU="}
                    ]
                }
			}
		`,
		},
		{
			Schema: mustParseGraphQLSchema(t, db),
			Label:  "LastSyncError included",
			Query: `
			{
				externalServices() {
					nodes {
						id
						lastSyncError
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes": [
                        {"id":"RXh0ZXJuYWxTZXJ2aWNlOjE=","lastSyncError":"Oops"},
                        {"id":"RXh0ZXJuYWxTZXJ2aWNlOjI=","lastSyncError":"Oops"},
						{"id":"RXh0ZXJuYWxTZXJ2aWNlOjM=","lastSyncError":"Oops"},
						{"id":"RXh0ZXJuYWxTZXJ2aWNlOjQ=","lastSyncError":"Oops"},
						{"id":"RXh0ZXJuYWxTZXJ2aWNlOjU=","lastSyncError":"Oops"}
                    ]
				}
			}
		`,
		},
		{
			Schema: mustParseGraphQLSchema(t, db),
			Label:  "Check connection",
			Query: `
				{
					externalServices() {
						nodes {
							id
							checkConnection {
								... on ExternalServiceAvailable {
									lastCheckedAt
								}
								... on ExternalServiceUnavailable {
									suspectedReason
								}
								... on ExternalServiceAvailabilityUnknown {
									implementationNote
								}
							}
							hasConnectionCheck
						}
					}
				}
			`,
			ExpectedResult: fmt.Sprintf(`
			{
				"externalServices": {
					"nodes": [
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjE=",
							"checkConnection": {
								"implementationNote": "not implemented"
							},
							"hasConnectionCheck": false
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjI=",
							"checkConnection": {
								"suspectedReason": "failed to connect"
							},
							"hasConnectionCheck": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjM=",
							"checkConnection": {
								"lastCheckedAt": %q
							},
							"hasConnectionCheck": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
							"checkConnection": {
								"implementationNote": "not implemented"
							},
							"hasConnectionCheck": false
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjU=",
							"checkConnection": {
								"implementationNote": "not implemented"
							},
							"hasConnectionCheck": false
						}
					]
				}
			}
			`, mockLastCheckedAt.Format("2006-01-02T15:04:05Z")),
		},
		{
			Schema: mustParseGraphQLSchema(t, db),
			Label:  "PageInfo included, using first",
			Query: `
			{
				externalServices(first: 1) {
					nodes {
						id
					}
					pageInfo {
						endCursor
						hasNextPage
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes":[{"id":"RXh0ZXJuYWxTZXJ2aWNlOjE="}],
					"pageInfo":{"endCursor":"RXh0ZXJuYWxTZXJ2aWNlOjE=","hasNextPage":true}
				}
			}
		`,
		},
		{
			Schema: mustParseGraphQLSchema(t, db),
			Label:  "PageInfo included, using after",
			Query: `
			{
				externalServices(after: "RXh0ZXJuYWxTZXJ2aWNlOjM=") {
					nodes {
						id
					}
					pageInfo {
						endCursor
						hasNextPage
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes":[{"id":"RXh0ZXJuYWxTZXJ2aWNlOjQ="},{"id":"RXh0ZXJuYWxTZXJ2aWNlOjU="}],
					"pageInfo":{"endCursor":null,"hasNextPage":false}
				}
			}
		`,
		},
		{
			Schema: mustParseGraphQLSchema(t, db),
			Label:  "SupportsRepoExclusion included",
			Query: `
			{
				externalServices() {
					nodes {
						id
						supportsRepoExclusion
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externalServices": {
					"nodes": [
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjE=",
							"supportsRepoExclusion": false
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjI=",
							"supportsRepoExclusion": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjM=",
							"supportsRepoExclusion": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjQ=",
							"supportsRepoExclusion": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2aWNlOjU=",
							"supportsRepoExclusion": false
						}
					]
				}
			}
		`,
		},
	})
}

func TestExternalServices_PageInfo(t *testing.T) {
	cmpOpts := cmp.AllowUnexported(graphqlutil.PageInfo{})
	tests := []struct {
		name         string
		opt          database.ExternalServicesListOptions
		mockList     func(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
		mockCount    func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error)
		wantPageInfo *graphqlutil.PageInfo
	}{
		{
			name: "no limit set",
			mockList: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
			},
			wantPageInfo: graphqlutil.HasNextPage(false),
		},
		{
			name: "less results than the limit",
			opt: database.ExternalServicesListOptions{
				LimitOffset: &database.LimitOffset{
					Limit: 10,
				},
			},
			mockList: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
			},
			wantPageInfo: graphqlutil.HasNextPage(false),
		},
		{
			name: "same number of results as the limit, and no more",
			opt: database.ExternalServicesListOptions{
				LimitOffset: &database.LimitOffset{
					Limit: 1,
				},
			},
			mockList: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
			},
			mockCount: func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error) {
				return 1, nil
			},
			wantPageInfo: graphqlutil.HasNextPage(false),
		},
		{
			name: "same number of results as the limit, and has more",
			opt: database.ExternalServicesListOptions{
				LimitOffset: &database.LimitOffset{
					Limit: 1,
				},
			},
			mockList: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return []*types.ExternalService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
			},
			mockCount: func(ctx context.Context, opt database.ExternalServicesListOptions) (int, error) {
				return 2, nil
			},
			wantPageInfo: graphqlutil.NextPageCursor(string(MarshalExternalServiceID(1))),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			externalServices := database.NewMockExternalServiceStore()
			externalServices.ListFunc.SetDefaultHook(test.mockList)
			externalServices.CountFunc.SetDefaultHook(test.mockCount)

			db := database.NewMockDB()
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			r := &externalServiceConnectionResolver{
				db:  db,
				opt: test.opt,
			}
			pageInfo, err := r.PageInfo(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantPageInfo, pageInfo, cmpOpts); diff != "" {
				t.Fatalf("PageInfo mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSyncExternalService_ContextTimeout(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Since the timeout in our test is set to 0ms, we do not need to sleep at all. If our code
		// is correct, this handler should timeout right away.
		w.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(func() { s.Close() })

	ctx := context.Background()
	svc := &types.ExternalService{
		Config: extsvc.NewEmptyConfig(),
	}

	err := backend.SyncExternalService(ctx, logtest.Scoped(t), svc, 0*time.Millisecond, repoupdater.NewClient(s.URL))

	if err == nil {
		t.Error("Expected error but got nil")
	}

	expected := "context deadline exceeded"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error: %q, but got %v", expected, err)
	}
}

func TestCancelExternalServiceSync(t *testing.T) {
	externalServiceID := int64(1234)
	syncJobID := int64(99)

	newExternalServices := func() *database.MockExternalServiceStore {
		externalServices := database.NewMockExternalServiceStore()
		externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
			return &types.ExternalService{
				ID:          externalServiceID,
				Kind:        extsvc.KindGitHub,
				DisplayName: "my external service",
				Config:      extsvc.NewUnencryptedConfig(`{}`),
			}, nil
		})

		externalServices.GetSyncJobByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalServiceSyncJob, error) {
			return &types.ExternalServiceSyncJob{
				ID:                id,
				State:             "processing",
				QueuedAt:          timeutil.Now().Add(-5 * time.Minute),
				StartedAt:         timeutil.Now(),
				ExternalServiceID: externalServiceID,
			}, nil
		})
		return externalServices
	}

	t.Run("as an admin with access to the external service", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

		externalServices := newExternalServices()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.ExternalServicesFunc.SetDefaultReturn(externalServices)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		syncJobIDGraphQL := marshalExternalServiceSyncJobID(syncJobID)

		RunTest(t, &Test{
			Schema:         mustParseGraphQLSchema(t, db),
			Query:          fmt.Sprintf(`mutation { cancelExternalServiceSync(id: %q) { alwaysNil } } `, syncJobIDGraphQL),
			ExpectedResult: `{ "cancelExternalServiceSync": { "alwaysNil": null } }`,
			Context:        ctx,
		})

		if callCount := len(externalServices.CancelSyncJobFunc.History()); callCount != 1 {
			t.Errorf("unexpected handle call count. want=%d have=%d", 1, callCount)
		} else if arg := externalServices.CancelSyncJobFunc.History()[0].Arg1; arg.ID != syncJobID {
			t.Errorf("unexpected sync job ID. want=%d have=%d", syncJobID, arg)
		}
	})

	t.Run("as a user without access to the external service", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		syncJobIDGraphQL := marshalExternalServiceSyncJobID(syncJobID)

		RunTest(t, &Test{
			Schema:         mustParseGraphQLSchema(t, db),
			Query:          fmt.Sprintf(`mutation { cancelExternalServiceSync(id: %q) { alwaysNil } } `, syncJobIDGraphQL),
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []any{"cancelExternalServiceSync"},
					Message:       auth.ErrMustBeSiteAdmin.Error(),
					ResolverError: auth.ErrMustBeSiteAdmin,
				},
			},
			Context: ctx,
		})
	})
}

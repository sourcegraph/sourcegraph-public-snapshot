package database

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExternalServicesListOptions_sqlConditions(t *testing.T) {
	tests := []struct {
		name                 string
		noNamespace          bool
		excludeNamespaceUser bool
		namespaceUserID      int32
		namespaceOrgID       int32
		kinds                []string
		afterID              int64
		updatedAfter         time.Time
		wantQuery            string
		onlyCloudDefault     bool
		includeDeleted       bool
		wantArgs             []any
	}{
		{
			name:      "no condition",
			wantQuery: "deleted_at IS NULL",
		},
		{
			name:      "only one kind: GitHub",
			kinds:     []string{extsvc.KindGitHub},
			wantQuery: "deleted_at IS NULL AND kind = ANY($1)",
			wantArgs:  []any{pq.Array([]string{extsvc.KindGitHub})},
		},
		{
			name:      "two kinds: GitHub and GitLab",
			kinds:     []string{extsvc.KindGitHub, extsvc.KindGitLab},
			wantQuery: "deleted_at IS NULL AND kind = ANY($1)",
			wantArgs:  []any{pq.Array([]string{extsvc.KindGitHub, extsvc.KindGitLab})},
		},
		{
			name:            "has namespace user ID",
			namespaceUserID: 1,
			wantQuery:       "deleted_at IS NULL AND namespace_user_id = $1",
			wantArgs:        []any{int32(1)},
		},
		{
			name:           "has namespace org ID",
			namespaceOrgID: 1,
			wantQuery:      "deleted_at IS NULL AND namespace_org_id = $1",
			wantArgs:       []any{int32(1)},
		},
		{
			name:            "want no namespace",
			noNamespace:     true,
			namespaceUserID: 1,
			namespaceOrgID:  42,
			wantQuery:       "deleted_at IS NULL AND namespace_user_id IS NULL AND namespace_org_id IS NULL",
		},
		{
			name:                 "want exclude namespace user",
			excludeNamespaceUser: true,
			wantQuery:            "deleted_at IS NULL AND namespace_user_id IS NULL",
		},
		{
			name:      "has after ID",
			afterID:   10,
			wantQuery: "deleted_at IS NULL AND id < $1",
			wantArgs:  []any{int64(10)},
		},
		{
			name:         "has after updated_at",
			updatedAfter: time.Date(2013, 04, 19, 0, 0, 0, 0, time.UTC),
			wantQuery:    "deleted_at IS NULL AND updated_at > $1",
			wantArgs:     []any{time.Date(2013, 04, 19, 0, 0, 0, 0, time.UTC)},
		},
		{
			name:             "has OnlyCloudDefault",
			onlyCloudDefault: true,
			wantQuery:        "deleted_at IS NULL AND cloud_default = true",
		},
		{
			name:           "includeDeleted",
			includeDeleted: true,
			wantQuery:      "TRUE",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := ExternalServicesListOptions{
				NoNamespace:          test.noNamespace,
				ExcludeNamespaceUser: test.excludeNamespaceUser,
				NamespaceUserID:      test.namespaceUserID,
				NamespaceOrgID:       test.namespaceOrgID,
				Kinds:                test.kinds,
				AfterID:              test.afterID,
				UpdatedAfter:         test.updatedAfter,
				OnlyCloudDefault:     test.onlyCloudDefault,
				IncludeDeleted:       test.includeDeleted,
			}
			q := sqlf.Join(opts.sqlConditions(), "AND")
			if diff := cmp.Diff(test.wantQuery, q.Query(sqlf.PostgresBindVar)); diff != "" {
				t.Fatalf("query mismatch (-want +got):\n%s", diff)
			} else if diff = cmp.Diff(test.wantArgs, q.Args()); diff != "" {
				t.Fatalf("args mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExternalServicesStore_ValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		kind            string
		config          string
		namespaceUserID int32
		namespaceOrgID  int32
		listFunc        func(ctx context.Context, opt ExternalServicesListOptions) ([]*types.ExternalService, error)
		wantErr         string
	}{
		{
			name:    "0 errors - GitHub.com",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			wantErr: "<nil>",
		},
		{
			name:    "0 errors - GitLab.com",
			kind:    extsvc.KindGitLab,
			config:  `{"url": "https://github.com", "projectQuery": ["none"], "token": "abc"}`,
			wantErr: "<nil>",
		},
		{
			name:    "0 errors - Bitbucket.org",
			kind:    extsvc.KindBitbucketCloud,
			config:  `{"url": "https://bitbucket.org", "username": "ceo", "appPassword": "abc"}`,
			wantErr: "<nil>",
		},
		{
			name:    "1 error",
			kind:    extsvc.KindGitHub,
			config:  `{"repositoryQuery": ["none"], "token": "fake"}`,
			wantErr: "url is required",
		},
		{
			name:    "2 errors",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": ""}`,
			wantErr: "2 errors occurred:\n\t* token: String length must be greater than or equal to 1\n\t* at least one of token or githubAppInstallationID must be set",
		},
		{
			name:   "no conflicting rate limit",
			kind:   extsvc.KindGitHub,
			config: `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "rateLimit": {"enabled": true, "requestsPerHour": 5000}}`,
			listFunc: func(ctx context.Context, opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return nil, nil
			},
			wantErr: "<nil>",
		},
		{
			name:            "prevent code hosts that are not allowed",
			kind:            extsvc.KindGitHub,
			config:          `{"url": "https://github.example.com", "repositoryQuery": ["none"], "token": "abc"}`,
			namespaceUserID: 1,
			wantErr:         `external service only allowed for https://github.com/ and https://gitlab.com/`,
		},
		{
			name:           "prevent code hosts that are not allowed for organizations",
			kind:           extsvc.KindGitHub,
			config:         `{"url": "https://github.example.com", "repositoryQuery": ["none"], "token": "abc"}`,
			namespaceOrgID: 1,
			wantErr:        `external service only allowed for https://github.com/ and https://gitlab.com/`,
		},
		{
			name:            "gjson handles comments",
			kind:            extsvc.KindGitHub,
			config:          `{"url": "https://github.com", "token": "abc", "repositoryQuery": ["affiliated"]} // comment`,
			namespaceUserID: 1,
			wantErr:         "<nil>",
		},
		{
			name:            "prevent disallowed repositoryPathPattern field",
			kind:            extsvc.KindGitHub,
			config:          `{"url": "https://github.com", "repositoryPathPattern": "github/{nameWithOwner}"}`,
			namespaceUserID: 1,
			wantErr:         `field "repositoryPathPattern" is not allowed in a user-added external service`,
		},
		{
			name:            "prevent disallowed nameTransformations field",
			kind:            extsvc.KindGitHub,
			config:          `{"url": "https://github.com", "nameTransformations": [{"regex": "\\.d/","replacement": "/"},{"regex": "-git$","replacement": ""}]}`,
			namespaceUserID: 1,
			wantErr:         `field "nameTransformations" is not allowed in a user-added external service`,
		},
		{
			name:            "prevent disallowed rateLimit field",
			kind:            extsvc.KindGitHub,
			config:          `{"url": "https://github.com", "rateLimit": {}}`,
			namespaceUserID: 1,
			wantErr:         `field "rateLimit" is not allowed in a user-added external service`,
		},
		{
			name:    "1 errors - GitHub.com",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "` + types.RedactedSecret + `"}`,
			wantErr: "unable to write external service config as it contains redacted fields, this is likely a bug rather than a problem with your config",
		},
		{
			name:    "1 errors - GitLab.com",
			kind:    extsvc.KindGitLab,
			config:  `{"url": "https://github.com", "projectQuery": ["none"], "token": "` + types.RedactedSecret + `"}`,
			wantErr: "unable to write external service config as it contains redacted fields, this is likely a bug rather than a problem with your config",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ess := NewMockExternalServiceStore()
			if test.listFunc != nil {
				ess.ListFunc.SetDefaultHook(test.listFunc)
			}
			_, err := ValidateExternalServiceConfig(context.Background(), ess, ValidateExternalServiceConfigOptions{
				Kind:            test.kind,
				Config:          test.config,
				NamespaceUserID: test.namespaceUserID,
				NamespaceOrgID:  test.namespaceOrgID,
			})
			gotErr := fmt.Sprintf("%v", err)
			if gotErr != test.wantErr {
				t.Errorf("error: want %q but got %q", test.wantErr, gotErr)
			}
		})
	}
}

func TestExternalServicesStore_Create(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(false)

	user, err := db.Users().Create(ctx,
		NewUser{
			Email:           "alice@example.com",
			Username:        "alice",
			Password:        "password",
			EmailIsVerified: true,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	displayName := "Acme org"
	org, err := db.Orgs().Create(ctx, "acme", &displayName)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	tests := []struct {
		name             string
		externalService  *types.ExternalService
		wantUnrestricted bool
		wantHasWebhooks  bool
	}{
		{
			name: "with webhooks",
			externalService: &types.ExternalService{
				Kind:            extsvc.KindGitHub,
				DisplayName:     "GITHUB #1",
				Config:          extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "webhooks": [{"org": "org", "secret": "secret"}]}`),
				NamespaceUserID: user.ID,
			},
			wantUnrestricted: false,
			wantHasWebhooks:  true,
		},
		{
			name: "without authorization",
			externalService: &types.ExternalService{
				Kind:            extsvc.KindGitHub,
				DisplayName:     "GITHUB #1",
				Config:          extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
				NamespaceUserID: user.ID,
			},
			wantUnrestricted: false,
			wantHasWebhooks:  false,
		},
		{
			name: "with authorization",
			externalService: &types.ExternalService{
				Kind:            extsvc.KindGitHub,
				DisplayName:     "GITHUB #2",
				Config:          extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`),
				NamespaceUserID: user.ID,
			},
			wantUnrestricted: false,
			wantHasWebhooks:  false,
		},
		{
			name: "with authorization in comments",
			externalService: &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "GITHUB #3",
				Config: extsvc.NewUnencryptedConfig(`
{
	"url": "https://github.com",
	"repositoryQuery": ["none"],
	"token": "abc",
	// "authorization": {}
}`),
				NamespaceUserID: user.ID,
			},
			wantUnrestricted: false,
		},

		{
			name: "Cloud: auto-add authorization to code host connections for GitHub",
			externalService: &types.ExternalService{
				Kind:            extsvc.KindGitHub,
				DisplayName:     "GITHUB #4",
				Config:          extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
				NamespaceUserID: user.ID,
			},
			wantUnrestricted: false,
			wantHasWebhooks:  false,
		},
		{
			name: "Cloud: auto-add authorization to code host connections for GitLab",
			externalService: &types.ExternalService{
				Kind:            extsvc.KindGitLab,
				DisplayName:     "GITLAB #1",
				Config:          extsvc.NewUnencryptedConfig(`{"url": "https://gitlab.com", "projectQuery": ["none"], "token": "abc"}`),
				NamespaceUserID: user.ID,
			},
			wantUnrestricted: false,
			wantHasWebhooks:  false,
		},
		{
			name: "Cloud: support org namespace on code host connections for GitHub",
			externalService: &types.ExternalService{
				Kind:           extsvc.KindGitHub,
				DisplayName:    "GITHUB #4",
				Config:         extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
				NamespaceOrgID: org.ID,
			},
			wantUnrestricted: false,
			wantHasWebhooks:  false,
		},
		{
			name: "Cloud: support org namespace on code host connections for GitLab",
			externalService: &types.ExternalService{
				Kind:           extsvc.KindGitLab,
				DisplayName:    "GITLAB #1",
				Config:         extsvc.NewUnencryptedConfig(`{"url": "https://gitlab.com", "projectQuery": ["none"], "token": "abc"}`),
				NamespaceOrgID: org.ID,
			},
			wantUnrestricted: false,
			wantHasWebhooks:  false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := db.ExternalServices().Create(ctx, confGet, test.externalService)
			if err != nil {
				t.Fatal(err)
			}

			// Should get back the same one
			got, err := db.ExternalServices().GetByID(ctx, test.externalService.ID)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.externalService, got, et.CompareEncryptable); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}

			if test.wantUnrestricted != got.Unrestricted {
				t.Fatalf("Want unrestricted = %v, but got %v", test.wantUnrestricted, got.Unrestricted)
			}

			if got.HasWebhooks == nil {
				t.Fatal("has_webhooks must not be null")
			} else if *got.HasWebhooks != test.wantHasWebhooks {
				t.Fatalf("Wanted has_webhooks = %v, but got %v", test.wantHasWebhooks, *got.HasWebhooks)
			}

			// Adding it another service with the same kind and owner should fail
			if test.externalService.NamespaceUserID != 0 || test.externalService.NamespaceOrgID != 0 {
				err := db.ExternalServices().Create(ctx, confGet, test.externalService)
				if err == nil {
					t.Fatal("Should not be able to create two services of same kind with same owner")
				}
			}

			err = db.ExternalServices().Delete(ctx, test.externalService.ID)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExternalServicesStore_CreateWithTierEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Background()
	confGet := func() *conf.Unified { return &conf.Unified{} }
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	store := db.ExternalServices()
	BeforeCreateExternalService = func(context.Context, ExternalServiceStore, *types.ExternalService) error {
		return errcode.NewPresentationError("test plan limit exceeded")
	}
	t.Cleanup(func() { BeforeCreateExternalService = nil })
	if err := store.Create(ctx, confGet, es); err == nil {
		t.Fatal("expected an error, got none")
	}
}

func TestExternalServicesStore_Update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(false)

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: The order of tests matters
	tests := []struct {
		name               string
		update             *ExternalServiceUpdate
		wantUnrestricted   bool
		wantCloudDefault   bool
		wantHasWebhooks    bool
		wantTokenExpiresAt bool
	}{
		{
			name: "update with authorization",
			update: &ExternalServiceUpdate{
				DisplayName: strptr("GITHUB (updated) #1"),
				Config:      strptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def", "authorization": {}, "webhooks": [{"org": "org", "secret": "secret"}]}`),
			},
			wantUnrestricted: false,
			wantCloudDefault: false,
			wantHasWebhooks:  true,
		},
		{
			name: "update without authorization",
			update: &ExternalServiceUpdate{
				DisplayName: strptr("GITHUB (updated) #2"),
				Config:      strptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
			},
			wantUnrestricted: false,
			wantCloudDefault: false,
			wantHasWebhooks:  false,
		},
		{
			name: "update with authorization in comments",
			update: &ExternalServiceUpdate{
				DisplayName: strptr("GITHUB (updated) #3"),
				Config: strptr(`
{
	"url": "https://github.com",
	"repositoryQuery": ["none"],
	"token": "def",
	// "authorization": {}
}`),
			},
			wantUnrestricted: false,
			wantCloudDefault: false,
			wantHasWebhooks:  false,
		},
		{
			name: "set cloud_default true",
			update: &ExternalServiceUpdate{
				DisplayName:  strptr("GITHUB (updated) #4"),
				CloudDefault: boolptr(true),
				Config: strptr(`
{
	"url": "https://github.com",
	"repositoryQuery": ["none"],
	"token": "def",
	"authorization": {},
	"webhooks": [{"org": "org", "secret": "secret"}]
}`),
			},
			wantUnrestricted: false,
			wantCloudDefault: true,
			wantHasWebhooks:  true,
		},
		{
			name: "update token_expires_at",
			update: &ExternalServiceUpdate{
				DisplayName:    strptr("GITHUB (updated) #5"),
				Config:         strptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
				TokenExpiresAt: timePtr(time.Now()),
			},
			wantCloudDefault:   true,
			wantTokenExpiresAt: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err = db.ExternalServices().Update(ctx, nil, es.ID, test.update)
			if err != nil {
				t.Fatal(err)
			}

			// Get and verify update
			got, err := db.ExternalServices().GetByID(ctx, es.ID)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(*test.update.DisplayName, got.DisplayName); diff != "" {
				t.Fatalf("DisplayName mismatch (-want +got):\n%s", diff)
			} else {
				cmpJSON := func(a, b string) string {
					normalize := func(s string) string {
						values := map[string]any{}
						_ = json.Unmarshal([]byte(s), &values)
						delete(values, "authorization")
						serialized, _ := json.Marshal(values)
						return string(serialized)
					}

					return cmp.Diff(normalize(a), normalize(b))
				}

				cfg, err := got.Config.Decrypt(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if diff = cmpJSON(*test.update.Config, cfg); diff != "" {
					t.Fatalf("Config mismatch (-want +got):\n%s", diff)
				} else if got.UpdatedAt.Equal(es.UpdatedAt) {
					t.Fatalf("UpdateAt: want to be updated but not")
				}
			}

			if test.wantUnrestricted != got.Unrestricted {
				t.Fatalf("Want unrestricted = %v, but got %v", test.wantUnrestricted, got.Unrestricted)
			}

			if test.wantCloudDefault != got.CloudDefault {
				t.Fatalf("Want cloud_default = %v, but got %v", test.wantCloudDefault, got.CloudDefault)
			}

			if got.HasWebhooks == nil {
				t.Fatal("has_webhooks is unexpectedly null")
			} else if test.wantHasWebhooks != *got.HasWebhooks {
				t.Fatalf("Want has_webhooks = %v, but got %v", test.wantHasWebhooks, *got.HasWebhooks)
			}

			if (got.TokenExpiresAt != nil) != test.wantTokenExpiresAt {
				t.Fatalf("Want token_expires_at = %v, but got %v", test.wantTokenExpiresAt, got.TokenExpiresAt)
			}
		})
	}
}

func TestUpsertAuthorizationToExternalService(t *testing.T) {
	tests := []struct {
		name   string
		kind   string
		config string
		want   string
	}{
		{
			name: "github with authorization",
			kind: extsvc.KindGitHub,
			config: `
{
  // Useful comments
  "url": "https://github.com",
  "repositoryQuery": ["none"],
  "token": "def",
  "authorization": {}
}`,
			want: `
{
  // Useful comments
  "url": "https://github.com",
  "repositoryQuery": ["none"],
  "token": "def",
  "authorization": {}
}`,
		},
		{
			name: "github without authorization",
			kind: extsvc.KindGitHub,
			config: `
{
  // Useful comments
  "url": "https://github.com",
  "repositoryQuery": ["none"],
  "token": "def"
}`,
			want: `
{
  // Useful comments
  "url": "https://github.com",
  "repositoryQuery": ["none"],
  "token": "def",
  "authorization": {}
}`,
		},
		{
			name: "gitlab with authorization",
			kind: extsvc.KindGitLab,
			config: `
{
  // Useful comments
  "url": "https://gitlab.com",
  "projectQuery": ["none"],
  "token": "abc",
  "authorization": {}
}`,
			want: `
{
  // Useful comments
  "url": "https://gitlab.com",
  "projectQuery": ["none"],
  "token": "abc",
  "authorization": {
    "identityProvider": {
      "type": "oauth"
    }
  }
}`,
		},
		{
			name: "gitlab without authorization",
			kind: extsvc.KindGitLab,
			config: `
{
  // Useful comments
  "url": "https://gitlab.com",
  "projectQuery": ["none"],
  "token": "abc"
}`,
			want: `
{
  // Useful comments
  "url": "https://gitlab.com",
  "projectQuery": ["none"],
  "token": "abc",
  "authorization": {
    "identityProvider": {
      "type": "oauth"
    }
  }
}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := upsertAuthorizationToExternalService(test.kind, test.config)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// This test ensures under Sourcegraph.com mode, every call of `Create`,
// `Upsert` and `Update` has the "authorization" field presented in the external
// service config automatically.
func TestExternalServicesStore_upsertAuthorizationToExternalService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(false)

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	externalServices := db.ExternalServices()

	// Test Create method
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := externalServices.Create(ctx, confGet, es)
	require.NoError(t, err)

	got, err := externalServices.GetByID(ctx, es.ID)
	require.NoError(t, err)
	cfg, err := got.Config.Decrypt(ctx)
	if err != nil {
		t.Fatal(err)
	}
	exists := gjson.Get(cfg, "authorization").Exists()
	assert.True(t, exists, `"authorization" field exists`)

	// Reset Config field and test Upsert method
	es.Config.Set(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`)
	err = externalServices.Upsert(ctx, es)
	require.NoError(t, err)

	got, err = externalServices.GetByID(ctx, es.ID)
	require.NoError(t, err)
	cfg, err = got.Config.Decrypt(ctx)
	if err != nil {
		t.Fatal(err)
	}
	exists = gjson.Get(cfg, "authorization").Exists()
	assert.True(t, exists, `"authorization" field exists`)

	// Reset Config field and test Update method
	es.Config.Set(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`)
	err = externalServices.Update(ctx,
		conf.Get().AuthProviders,
		es.ID,
		&ExternalServiceUpdate{
			Config: &cfg,
		},
	)
	require.NoError(t, err)

	got, err = externalServices.GetByID(ctx, es.ID)
	require.NoError(t, err)
	cfg, err = got.Config.Decrypt(ctx)
	if err != nil {
		t.Fatal(err)
	}
	exists = gjson.Get(cfg, "authorization").Exists()
	assert.True(t, exists, `"authorization" field exists`)
}

func TestCountRepoCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es1 := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.ExecContext(ctx, `
INSERT INTO repo (id, name, description, fork)
VALUES (1, 'github.com/user/repo', '', FALSE);
`)
	if err != nil {
		t.Fatal(err)
	}

	// Insert rows to `external_service_repos` table to test the trigger.
	q := sqlf.Sprintf(`
INSERT INTO external_service_repos (external_service_id, repo_id, clone_url)
VALUES (%d, 1, '')
`, es1.ID)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	count, err := db.ExternalServices().RepoCount(ctx, es1.ID)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("Expected 1, got %d", count)
	}
}

func TestExternalServicesStore_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es1 := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es1)
	if err != nil {
		t.Fatal(err)
	}

	es2 := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #2",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
	}
	err = db.ExternalServices().Create(ctx, confGet, es2)
	if err != nil {
		t.Fatal(err)
	}

	// Create two repositories to test trigger of soft-deleting external service:
	//  - ID=1 is expected to be deleted along with deletion of the external service.
	//  - ID=2 remains untouched because it is not associated with the external service.
	_, err = db.ExecContext(ctx, `
INSERT INTO repo (id, name, description, fork)
VALUES (1, 'github.com/user/repo', '', FALSE);
INSERT INTO repo (id, name, description, fork)
VALUES (2, 'github.com/user/repo2', '', FALSE);
`)
	if err != nil {
		t.Fatal(err)
	}

	// Insert rows to `external_service_repos` table to test the trigger.
	q := sqlf.Sprintf(`
INSERT INTO external_service_repos (external_service_id, repo_id, clone_url)
VALUES (%d, 1, ''), (%d, 2, '')
`, es1.ID, es2.ID)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	// Delete this external service
	err = db.ExternalServices().Delete(ctx, es1.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Delete again should get externalServiceNotFoundError
	err = db.ExternalServices().Delete(ctx, es1.ID)
	gotErr := fmt.Sprintf("%v", err)
	wantErr := fmt.Sprintf("external service not found: %v", es1.ID)
	if gotErr != wantErr {
		t.Errorf("error: want %q but got %q", wantErr, gotErr)
	}

	_, err = db.ExternalServices().GetByID(ctx, es1.ID)
	if err == nil {
		t.Fatal("expected an error")
	}

	// Should only get back the repo with ID=2
	repos, err := db.Repos().GetByIDs(ctx, 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	want := []*types.Repo{
		{ID: 2, Name: "github.com/user/repo2"},
	}

	repos = types.Repos(repos).With(func(r *types.Repo) {
		r.CreatedAt = time.Time{}
		r.Sources = nil
	})
	if diff := cmp.Diff(want, repos); diff != "" {
		t.Fatalf("Repos mismatch (-want +got):\n%s", diff)
	}
}

// reposNumber is a number of repos created in one batch.
// TestExternalServicesStore_DeleteExtServiceWithManyRepos does 5 such batches
const reposNumber = 1000

// TestExternalServicesStore_DeleteExtServiceWithManyRepos can be used locally
// with increased number of repos to see how fast/slow deletion of external
// services works.
func TestExternalServicesStore_DeleteExtServiceWithManyRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	extSvc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	servicesStore := db.ExternalServices()
	err := servicesStore.Create(ctx, confGet, extSvc)
	if err != nil {
		t.Fatal(err)
	}

	createRepo := func(offset int, c chan<- int) {
		inserter := func(inserter *batch.Inserter) error {
			for i := 0 + offset; i < reposNumber+offset; i++ {
				if err := inserter.Insert(ctx, i, "repo"+strconv.Itoa(i)); err != nil {
					return err
				}
			}
			return nil
		}

		if err := batch.WithInserter(
			ctx,
			db,
			"repo",
			batch.MaxNumPostgresParameters,
			[]string{"id", "name"},
			inserter,
		); err != nil {
			t.Error(err)
			c <- 1
			return
		}
		c <- 0
	}

	ready := make(chan int, 5)
	defer close(ready)
	offsets := []int{0, reposNumber, reposNumber * 2, reposNumber * 3, reposNumber * 4}

	for _, offset := range offsets {
		go createRepo(offset, ready)
	}

	for i := 0; i < 5; i++ {
		if status := <-ready; status != 0 {
			t.Fatal("Error during repo creation")
		}
	}

	ready2 := make(chan int, 5)
	defer close(ready2)

	extSvcId := extSvc.ID

	createExtSvc := func(offset int, c chan<- int) {
		inserter := func(inserter *batch.Inserter) error {
			for i := 0 + offset; i < reposNumber+offset; i++ {
				if err := inserter.Insert(ctx, extSvcId, i, ""); err != nil {
					return err
				}
			}
			return nil
		}

		if err := batch.WithInserter(
			ctx,
			db,
			"external_service_repos",
			batch.MaxNumPostgresParameters,
			[]string{"external_service_id", "repo_id", "clone_url"},
			inserter,
		); err != nil {
			t.Error(err)
			c <- 1
			return
		}
		c <- 0
	}

	for _, offset := range offsets {
		go createExtSvc(offset, ready2)
	}

	for i := 0; i < 5; i++ {
		if status := <-ready2; status != 0 {
			t.Fatal("Error during external service repo creation")
		}
	}

	// Delete this external service
	start := time.Now()
	err = servicesStore.Delete(ctx, extSvcId)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Deleting of an external service with %d repos took %s", reposNumber*5, time.Since(start))

	count, err := servicesStore.RepoCount(ctx, extSvcId)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("External service repos are not deleted")
	}

	// Should throw not found error
	_, err = servicesStore.GetByID(ctx, extSvcId)
	if err == nil {
		t.Fatal("External service is not deleted")
	}

	rows, err := db.Handle().QueryContext(ctx, `select * from repo where deleted_at is null`)
	if err != nil {
		t.Fatal("Error during fetching repos from the DB")
	}
	if rows.Next() {
		t.Fatal("Repos of external service are not deleted")
	}
}

func TestExternalServicesStore_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// Should be able to get back by its ID
	_, err = db.ExternalServices().GetByID(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Delete this external service
	err = db.ExternalServices().Delete(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Should now get externalServiceNotFoundError
	_, err = db.ExternalServices().GetByID(ctx, es.ID)
	gotErr := fmt.Sprintf("%v", err)
	wantErr := fmt.Sprintf("external service not found: %v", es.ID)
	if gotErr != wantErr {
		t.Errorf("error: want %q but got %q", wantErr, gotErr)
	}
}

func TestExternalServicesStore_GetByID_Encrypted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}

	store := db.ExternalServices().WithEncryptionKey(et.TestKey{})

	err := store.Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// values encrypted should not be readable without the encrypting key
	noopStore := store.WithEncryptionKey(&encryption.NoopKey{FailDecrypt: true})
	svc, err := noopStore.GetByID(ctx, es.ID)
	if err != nil {
		t.Fatalf("unexpected error querying service: %s", err)
	}
	if _, err := svc.Config.Decrypt(ctx); err == nil {
		t.Fatalf("expected error decrypting with a different key")
	}

	// Should be able to get back by its ID
	_, err = store.GetByID(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Delete this external service
	err = store.Delete(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Should now get externalServiceNotFoundError
	_, err = store.GetByID(ctx, es.ID)
	gotErr := fmt.Sprintf("%v", err)
	wantErr := fmt.Sprintf("external service not found: %v", es.ID)
	if gotErr != wantErr {
		t.Errorf("error: want %q but got %q", wantErr, gotErr)
	}
}

func TestGetAffiliatedSyncErrors(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	// Initial user always gets created as an admin
	admin, err := db.Users().Create(ctx, NewUser{
		Email:                 "a1@example.com",
		Username:              "u1",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	user2, err := db.Users().Create(ctx, NewUser{
		Email:                 "u2@example.com",
		Username:              "u2",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	org1, err := db.Orgs().Create(ctx, "ACME", nil)
	if err != nil {
		t.Fatal(err)
	}

	createService := func(u *types.User, o *types.Org, name string) *types.ExternalService {
		svc := &types.ExternalService{
			Kind:        extsvc.KindGitHub,
			DisplayName: name,
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
		}

		if u != nil {
			svc.NamespaceUserID = u.ID
		}

		if o != nil {
			svc.NamespaceOrgID = o.ID
		}

		err = db.ExternalServices().Create(ctx, confGet, svc)
		if err != nil {
			t.Fatal(err)
		}
		return svc
	}

	countErrors := func(results map[int64]string) int {
		var errorCount int
		for _, v := range results {
			if len(v) > 0 {
				errorCount++
			}
		}
		return errorCount
	}

	siteLevel := createService(nil, nil, "GITHUB #1")
	adminOwned := createService(admin, nil, "GITHUB #2")
	userOwned := createService(user2, nil, "GITHUB #3")

	// Listing errors now should return an empty map as none have been added yet
	results, err := db.ExternalServices().GetAffiliatedSyncErrors(ctx, admin)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	errorCount := countErrors(results)
	if errorCount != 0 {
		t.Fatal("Expected 0 errors")
	}

	// Add two failures for the same service
	failure1 := "oops"
	_, err = db.Handle().ExecContext(ctx, `
INSERT INTO external_service_sync_jobs (external_service_id, state, finished_at, failure_message)
VALUES ($1,'errored', now(), $2)
`, siteLevel.ID, failure1)
	if err != nil {
		t.Fatal(err)
	}
	failure2 := "oops again"
	_, err = db.Handle().ExecContext(ctx, `
INSERT INTO external_service_sync_jobs (external_service_id, state, finished_at, failure_message)
VALUES ($1,'errored', now(), $2)
`, siteLevel.ID, failure2)
	if err != nil {
		t.Fatal(err)
	}

	// We should get the latest failure
	results, err = db.ExternalServices().GetAffiliatedSyncErrors(ctx, admin)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	errorCount = countErrors(results)
	if errorCount != 1 {
		t.Fatal("Expected 1 error")
	}
	failure := results[siteLevel.ID]
	if failure != failure2 {
		t.Fatalf("Want %q, got %q", failure2, failure)
	}

	// Adding a second failing service
	_, err = db.Handle().ExecContext(ctx, `
INSERT INTO external_service_sync_jobs (external_service_id, state, finished_at, failure_message)
VALUES ($1,'errored', now(), $2)
`, adminOwned.ID, failure1)
	if err != nil {
		t.Fatal(err)
	}

	results, err = db.ExternalServices().GetAffiliatedSyncErrors(ctx, admin)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatal("Expected 2 results")
	}
	errorCount = countErrors(results)
	if errorCount != 2 {
		t.Fatal("Expected 2 errors")
	}

	// User should not see any failures as they don't own any services
	results, err = db.ExternalServices().GetAffiliatedSyncErrors(ctx, user2)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatal("Expected 1 result")
	}
	errorCount = countErrors(results)
	if errorCount != 0 {
		t.Fatal("Expected 0 errors")
	}

	// Add a failure to user service
	failure3 := "user failure"
	_, err = db.Handle().ExecContext(ctx, `
INSERT INTO external_service_sync_jobs (external_service_id, state, finished_at, failure_message)
VALUES ($1,'errored', now(), $2)
`, userOwned.ID, failure3)
	if err != nil {
		t.Fatal(err)
	}

	// We should get the latest failure
	results, err = db.ExternalServices().GetAffiliatedSyncErrors(ctx, user2)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatal("Expected 1 result")
	}
	errorCount = countErrors(results)
	if errorCount != 1 {
		t.Fatal("Expected 1 error")
	}
	failure = results[userOwned.ID]
	if failure != failure3 {
		t.Fatalf("Want %q, got %q", failure3, failure)
	}

	// Add a failure to org service
	orgOwned := createService(nil, org1, "GITHUB Org owned")

	_, err = db.Handle().ExecContext(ctx, `
INSERT INTO external_service_sync_jobs (external_service_id, state, finished_at, failure_message)
VALUES ($1,'errored', now(), $2)
`, orgOwned.ID, "org failure")
	if err != nil {
		t.Fatal(err)
	}

	// Assert that site-admin should only get errors for site level external services
	// or self owned external services.
	results, err = db.ExternalServices().GetAffiliatedSyncErrors(ctx, admin)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %v", results)
	}

	if _, ok := results[siteLevel.ID]; !ok {
		t.Fatalf("expected admin to only get errors for site level external services and self-owned, got %+v", results)
	}

	if _, ok := results[adminOwned.ID]; !ok {
		t.Fatalf("expected admin to only get errors for site level external services and self-owned, got %+v", results)
	}
}

func TestGetLastSyncError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// Should be able to get back by its ID
	_, err = db.ExternalServices().GetByID(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	lastSyncError, err := db.ExternalServices().GetLastSyncError(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}
	if lastSyncError != "" {
		t.Fatalf("Expected empty error, have %q", lastSyncError)
	}

	// Could have failure message
	_, err = db.Handle().ExecContext(ctx, `
INSERT INTO external_service_sync_jobs (external_service_id, state, finished_at)
VALUES ($1,'errored', now())
`, es.ID)

	if err != nil {
		t.Fatal(err)
	}

	lastSyncError, err = db.ExternalServices().GetLastSyncError(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}
	if lastSyncError != "" {
		t.Fatalf("Expected empty error, have %q", lastSyncError)
	}

	// Add sync error
	expectedError := "oops"
	_, err = db.Handle().ExecContext(ctx, `
INSERT INTO external_service_sync_jobs (external_service_id, failure_message, state, finished_at)
VALUES ($1,$2,'errored', now())
`, es.ID, expectedError)

	if err != nil {
		t.Fatal(err)
	}

	lastSyncError, err = db.ExternalServices().GetLastSyncError(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}
	if lastSyncError != expectedError {
		t.Fatalf("Expected %q, have %q", expectedError, lastSyncError)
	}
}

func TestExternalServicesStore_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create test user
	user, err := db.Users().Create(ctx, NewUser{
		Email:           "alice@example.com",
		Username:        "alice",
		Password:        "password",
		EmailIsVerified: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create test org
	displayName := "Acme Org"
	org, err := db.Orgs().Create(ctx, "acme", &displayName)
	if err != nil {
		t.Fatal(err)
	}

	// Create new external services
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	ess := []*types.ExternalService{
		{
			Kind:            extsvc.KindGitHub,
			DisplayName:     "GITHUB #1",
			Config:          extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`),
			NamespaceUserID: user.ID,
			CloudDefault:    true,
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "GITHUB #2",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
		},
		{
			Kind:           extsvc.KindGitHub,
			DisplayName:    "GITHUB #3",
			Config:         extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def", "authorization": {}}`),
			NamespaceOrgID: org.ID,
		},
	}

	for _, es := range ess {
		err := db.ExternalServices().Create(ctx, confGet, es)
		if err != nil {
			t.Fatal(err)
		}
	}
	createdAt := time.Now()

	deletedES := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #4",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
	}
	err = db.ExternalServices().Create(ctx, confGet, deletedES)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.ExternalServices().Delete(ctx, deletedES.ID); err != nil {
		t.Fatal(err)
	}

	t.Run("list all external services", func(t *testing.T) {
		got, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })

		if diff := cmp.Diff(ess, got, et.CompareEncryptable); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list all external services in ascending order", func(t *testing.T) {
		got, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{OrderByDirection: "ASC"})
		if err != nil {
			t.Fatal(err)
		}
		want := []*types.ExternalService(types.ExternalServices(ess).Clone())
		sort.Slice(want, func(i, j int) bool { return want[i].ID < want[j].ID })

		if diff := cmp.Diff(want, got, et.CompareEncryptable); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list all external services in descending order", func(t *testing.T) {
		got, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := []*types.ExternalService(types.ExternalServices(ess).Clone())
		sort.Slice(want, func(i, j int) bool { return want[i].ID > want[j].ID })

		if diff := cmp.Diff(want, got, et.CompareEncryptable); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list external services with certain IDs", func(t *testing.T) {
		got, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			IDs: []int64{ess[1].ID},
		})
		if err != nil {
			t.Fatal(err)
		}
		sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })

		if diff := cmp.Diff(ess[1:2], got, et.CompareEncryptable); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list external services with no namespace", func(t *testing.T) {
		got, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			NoNamespace: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 1 {
			t.Fatalf("Want 1 external service but got %d", len(ess))
		} else if diff := cmp.Diff(ess[1], got[0], et.CompareEncryptable); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list only test user's external services", func(t *testing.T) {
		got, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			NamespaceUserID: user.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 1 {
			t.Fatalf("Want 1 external service but got %d", len(ess))
		} else if diff := cmp.Diff(ess[0], got[0], et.CompareEncryptable); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list non-exist user's external services", func(t *testing.T) {
		ess, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			NamespaceUserID: 404,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(ess) != 0 {
			t.Fatalf("Want 0 external service but got %d", len(ess))
		}
	})

	t.Run("list only test org's external services", func(t *testing.T) {
		got, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			NamespaceOrgID: org.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 1 {
			t.Fatalf("Want 1 external service but got %d", len(ess))
		} else if diff := cmp.Diff(ess[2], got[0], et.CompareEncryptable); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list non-existing org external services", func(t *testing.T) {
		ess, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			NamespaceOrgID: 404,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(ess) != 0 {
			t.Fatalf("Want 0 external service but got %d", len(ess))
		}
	})

	t.Run("list services updated after a certain date, expect 0", func(t *testing.T) {
		ess, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			UpdatedAfter: createdAt,
		})
		if err != nil {
			t.Fatal(err)
		}
		// We expect zero services to have been updated after they were created
		if len(ess) != 0 {
			t.Fatalf("Want 0 external service but got %d", len(ess))
		}
	})

	t.Run("list services updated after a certain date, expect 3", func(t *testing.T) {
		ess, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			UpdatedAfter: createdAt.Add(-5 * time.Minute),
		})
		if err != nil {
			t.Fatal(err)
		}
		// We should find all services were updated after a time in the past
		if len(ess) != 3 {
			t.Fatalf("Want 3 external services but got %d", len(ess))
		}
	})

	t.Run("list cloud default services", func(t *testing.T) {
		ess, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			OnlyCloudDefault: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		// We should find all services were updated after a time in the past
		if len(ess) != 1 {
			t.Fatalf("Want 0 external services but got %d", len(ess))
		}
	})

	t.Run("list including deleted", func(t *testing.T) {
		ess, err := db.ExternalServices().List(ctx, ExternalServicesListOptions{
			IncludeDeleted: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		// We should find all services were updated after a time in the past
		if len(ess) != 4 {
			t.Fatalf("Want 4 external services but got %d", len(ess))
		}
	})
}

func TestExternalServicesStore_DistinctKinds(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	t.Run("no external service won't blow up", func(t *testing.T) {
		kinds, err := db.ExternalServices().DistinctKinds(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(kinds) != 0 {
			t.Fatalf("Kinds: want 0 but got %d", len(kinds))
		}
	})

	// Create new external services in different kinds
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	ess := []*types.ExternalService{
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "GITHUB #1",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "GITHUB #2",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
		},
		{
			Kind:        extsvc.KindGitLab,
			DisplayName: "GITLAB #1",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "projectQuery": ["none"], "token": "abc"}`),
		},
		{
			Kind:        extsvc.KindOther,
			DisplayName: "OTHER #1",
			Config:      extsvc.NewUnencryptedConfig(`{"repos": []}`),
		},
	}
	for _, es := range ess {
		err := db.ExternalServices().Create(ctx, confGet, es)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Delete the last external service which should be excluded from the result
	err := db.ExternalServices().Delete(ctx, ess[3].ID)
	if err != nil {
		t.Fatal(err)
	}

	kinds, err := db.ExternalServices().DistinctKinds(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(kinds)
	wantKinds := []string{extsvc.KindGitHub, extsvc.KindGitLab}
	if diff := cmp.Diff(wantKinds, kinds); diff != "" {
		t.Fatalf("Kinds mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalServicesStore_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	count, err := db.ExternalServices().Count(ctx, ExternalServicesListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("Want 1 external service but got %d", count)
	}
}

func TestExternalServicesStore_Upsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	clock := timeutil.NewFakeClock(time.Now(), 0)

	svcs := typestest.MakeExternalServices()

	t.Run("no external services", func(t *testing.T) {
		if err := db.ExternalServices().Upsert(ctx); err != nil {
			t.Fatalf("Upsert error: %s", err)
		}
	})

	t.Run("one external service", func(t *testing.T) {
		tx, err := db.ExternalServices().Transact(ctx)
		if err != nil {
			t.Fatalf("Transact error: %s", err)
		}
		defer func() {
			err = tx.Done(err)
			if err != nil {
				t.Fatalf("Done error: %s", err)
			}
		}()

		svc := svcs[1]
		if svc.Kind != extsvc.KindGitLab {
			t.Fatalf("expected external service at [1] to be GitLab; got %s", svc.Kind)
		}

		if err := tx.Upsert(ctx, svc); err != nil {
			t.Fatalf("upsert error: %v", err)
		}
		if *svc.HasWebhooks != false {
			t.Fatalf("unexpected HasWebhooks: %v", svc.HasWebhooks)
		}

		cfg, err := svc.Config.Decrypt(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Add webhooks to the config and upsert.
		svc.Config.Set(`{"webhooks":[{"secret": "secret"}],` + cfg[1:])
		if err := tx.Upsert(ctx, svc); err != nil {
			t.Fatalf("upsert error: %v", err)
		}
		if *svc.HasWebhooks != true {
			t.Fatalf("unexpected HasWebhooks: %v", svc.HasWebhooks)
		}
	})

	t.Run("many external services", func(t *testing.T) {
		user, err := db.Users().Create(ctx, NewUser{Username: "alice"})
		if err != nil {
			t.Fatalf("Test setup error %s", err)
		}
		org, err := db.Orgs().Create(ctx, "acme", nil)
		if err != nil {
			t.Fatalf("Test setup error %s", err)
		}

		namespacedSvcs := typestest.MakeNamespacedExternalServices(user.ID, org.ID)

		tx, err := db.ExternalServices().Transact(ctx)
		if err != nil {
			t.Fatalf("Transact error: %s", err)
		}
		defer func() {
			err = tx.Done(err)
			if err != nil {
				t.Fatalf("Done error: %s", err)
			}
		}()

		services := append(svcs, namespacedSvcs...)
		want := typestest.GenerateExternalServices(11, services...)

		if err := tx.Upsert(ctx, want...); err != nil {
			t.Fatalf("Upsert error: %s", err)
		}

		for _, e := range want {
			if e.Kind != strings.ToUpper(e.Kind) {
				t.Errorf("external service kind didn't get upper-cased: %q", e.Kind)
				break
			}
		}

		sort.Sort(want)

		have, err := tx.List(ctx, ExternalServicesListOptions{
			Kinds: services.Kinds(),
		})
		if err != nil {
			t.Fatalf("List error: %s", err)
		}

		sort.Sort(types.ExternalServices(have))

		if diff := cmp.Diff(have, []*types.ExternalService(want), cmpopts.EquateEmpty(), et.CompareEncryptable); diff != "" {
			t.Fatalf("List:\n%s", diff)
		}

		// We'll update the external services, but being careful to keep the
		// config valid as we go.
		now := clock.Now()
		suffix := "-updated"
		for _, r := range want {
			cfg, err := r.Config.Decrypt(ctx)
			if err != nil {
				t.Fatal(err)
			}

			r.DisplayName += suffix
			r.Config.Set(`{"wanted":true,` + cfg[1:])
			r.UpdatedAt = now
			r.CreatedAt = now
		}

		if err = tx.Upsert(ctx, want...); err != nil {
			t.Errorf("Upsert error: %s", err)
		}
		have, err = tx.List(ctx, ExternalServicesListOptions{})
		if err != nil {
			t.Fatalf("List error: %s", err)
		}

		sort.Sort(types.ExternalServices(have))

		if diff := cmp.Diff(have, []*types.ExternalService(want), cmpopts.EquateEmpty(), et.CompareEncryptable); diff != "" {
			t.Errorf("List:\n%s", diff)
		}

		// Delete external services
		for _, es := range want {
			if err := tx.Delete(ctx, es.ID); err != nil {
				t.Fatal(err)
			}
		}

		have, err = tx.List(ctx, ExternalServicesListOptions{})
		if err != nil {
			t.Errorf("List error: %s", err)
		}

		sort.Sort(types.ExternalServices(have))

		if diff := cmp.Diff(have, []*types.ExternalService(nil), cmpopts.EquateEmpty(), et.CompareEncryptable); diff != "" {
			t.Errorf("List:\n%s", diff)
		}
	})

	t.Run("with encryption key", func(t *testing.T) {
		tx, err := db.ExternalServices().WithEncryptionKey(et.TestKey{}).Transact(ctx)
		if err != nil {
			t.Fatalf("Transact error: %s", err)
		}
		defer func() {
			err = tx.Done(err)
			if err != nil {
				t.Fatalf("Done error: %s", err)
			}
		}()

		want := typestest.GenerateExternalServices(7, svcs...)

		if err := tx.Upsert(ctx, want...); err != nil {
			t.Fatalf("Upsert error: %s", err)
		}
		for _, e := range want {
			if e.Kind != strings.ToUpper(e.Kind) {
				t.Errorf("external service kind didn't get upper-cased: %q", e.Kind)
				break
			}
		}

		// values encrypted should not be readable without the encrypting key
		noopStore := ExternalServicesWith(logger, tx).WithEncryptionKey(&encryption.NoopKey{FailDecrypt: true})

		for _, e := range want {
			svc, err := noopStore.GetByID(ctx, e.ID)
			if err != nil {
				t.Fatalf("unexpected error querying service: %s", err)
			}
			if _, err := svc.Config.Decrypt(ctx); err == nil {
				t.Fatalf("expected error decrypting with a different key")
			}
		}

		have, err := tx.List(ctx, ExternalServicesListOptions{
			Kinds: svcs.Kinds(),
		})
		if err != nil {
			t.Fatalf("List error: %s", err)
		}

		sort.Sort(types.ExternalServices(have))
		sort.Sort(want)

		if diff := cmp.Diff(have, []*types.ExternalService(want), cmpopts.EquateEmpty(), et.CompareEncryptable); diff != "" {
			t.Fatalf("List:\n%s", diff)
		}

		// We'll update the external services, but being careful to keep the
		// config valid as we go.
		now := clock.Now()
		suffix := "-updated"
		for _, r := range want {
			cfg, err := r.Config.Decrypt(ctx)
			if err != nil {
				t.Fatal(err)
			}

			r.DisplayName += suffix
			r.Config.Set(`{"wanted":true,` + cfg[1:])
			r.UpdatedAt = now
			r.CreatedAt = now
		}

		if err = tx.Upsert(ctx, want...); err != nil {
			t.Errorf("Upsert error: %s", err)
		}
		have, err = tx.List(ctx, ExternalServicesListOptions{})
		if err != nil {
			t.Fatalf("List error: %s", err)
		}

		sort.Sort(types.ExternalServices(have))

		if diff := cmp.Diff(have, []*types.ExternalService(want), cmpopts.EquateEmpty(), et.CompareEncryptable); diff != "" {
			t.Errorf("List:\n%s", diff)
		}

		// Delete external services
		for _, es := range want {
			if err := tx.Delete(ctx, es.ID); err != nil {
				t.Fatal(err)
			}
		}

		have, err = tx.List(ctx, ExternalServicesListOptions{})
		if err != nil {
			t.Errorf("List error: %s", err)
		}

		sort.Sort(types.ExternalServices(have))

		if diff := cmp.Diff(have, []*types.ExternalService(nil), cmpopts.EquateEmpty(), et.CompareEncryptable); diff != "" {
			t.Errorf("List:\n%s", diff)
		}
	})
}

func TestExternalServiceStore_GetSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Handle().ExecContext(ctx, "INSERT INTO external_service_sync_jobs (external_service_id) VALUES ($1)", es.ID)
	if err != nil {
		t.Fatal(err)
	}

	have, err := db.ExternalServices().GetSyncJobs(ctx, ExternalServicesGetSyncJobsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(have) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(have))
	}

	want := &types.ExternalServiceSyncJob{
		ID:                1,
		State:             "queued",
		ExternalServiceID: es.ID,
	}
	if diff := cmp.Diff(want, have[0], cmpopts.IgnoreFields(types.ExternalServiceSyncJob{}, "ID", "QueuedAt")); diff != "" {
		t.Fatal(diff)
	}
}

func TestExternalServiceStore_CountSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Handle().ExecContext(ctx, "INSERT INTO external_service_sync_jobs (external_service_id) VALUES ($1)", es.ID)
	if err != nil {
		t.Fatal(err)
	}

	have, err := db.ExternalServices().CountSyncJobs(ctx, ExternalServicesGetSyncJobsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if have != 1 {
		t.Fatalf("Expected 1 job, got %d", have)
	}

	require.Exactly(t, int64(1), have, "total count is incorrect")

	have, err = db.ExternalServices().CountSyncJobs(ctx, ExternalServicesGetSyncJobsOptions{ExternalServiceID: es.ID + 1})
	if err != nil {
		t.Fatal(err)
	}
	if have != 0 {
		t.Fatalf("Expected 0 jobs, got %d", have)
	}

	require.Exactly(t, int64(0), have, "total count is incorrect")
}

func TestExternalServiceStore_GetSyncJobByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Handle().ExecContext(ctx, "INSERT INTO external_service_sync_jobs (id, external_service_id) VALUES (1, $1)", es.ID)
	if err != nil {
		t.Fatal(err)
	}

	have, err := db.ExternalServices().GetSyncJobByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	want := &types.ExternalServiceSyncJob{
		ID:                1,
		State:             "queued",
		ExternalServiceID: es.ID,
	}
	if diff := cmp.Diff(want, have, cmpopts.IgnoreFields(types.ExternalServiceSyncJob{}, "ID", "QueuedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Test not found:
	_, err = db.ExternalServices().GetSyncJobByID(ctx, 2)
	if err == nil {
		t.Fatal("no error for not found")
	}
	if !errcode.IsNotFound(err) {
		t.Fatal("wrong err code for not found")
	}
}

func TestExternalServicesStore_OneCloudDefaultPerKind(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	now := time.Now()

	makeService := func(cloudDefault bool) *types.ExternalService {
		cfg := `{"url": "https://github.com", "token": "abc", "repositoryQuery": ["none"]}`
		svc := &types.ExternalService{
			Kind:         extsvc.KindGitHub,
			DisplayName:  "Github - Test",
			Config:       extsvc.NewUnencryptedConfig(cfg),
			CreatedAt:    now,
			UpdatedAt:    now,
			CloudDefault: cloudDefault,
		}
		return svc
	}

	t.Run("non default", func(t *testing.T) {
		gh := makeService(false)
		if err := db.ExternalServices().Upsert(ctx, gh); err != nil {
			t.Fatalf("Upsert error: %s", err)
		}
	})

	t.Run("first default", func(t *testing.T) {
		gh := makeService(true)
		if err := db.ExternalServices().Upsert(ctx, gh); err != nil {
			t.Fatalf("Upsert error: %s", err)
		}
	})

	t.Run("second default", func(t *testing.T) {
		gh := makeService(true)
		if err := db.ExternalServices().Upsert(ctx, gh); err == nil {
			t.Fatal("Expected an error")
		}
	})
}

func TestExternalServiceStore_SyncDue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	now := time.Now()

	makeService := func() *types.ExternalService {
		cfg := `{"url": "https://github.com", "token": "abc", "repositoryQuery": ["none"]}`
		svc := &types.ExternalService{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test",
			Config:      extsvc.NewUnencryptedConfig(cfg),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		return svc
	}
	svc1 := makeService()
	svc2 := makeService()
	err := db.ExternalServices().Upsert(ctx, svc1, svc2)
	if err != nil {
		t.Fatal(err)
	}

	assertDue := func(d time.Duration, want bool) {
		t.Helper()
		ids := []int64{svc1.ID, svc2.ID}
		due, err := db.ExternalServices().SyncDue(ctx, ids, d)
		if err != nil {
			t.Error(err)
		}
		if due != want {
			t.Errorf("Want %v, got %v", want, due)
		}
	}

	makeSyncJob := func(svcID int64, state string) {
		_, err = db.Handle().ExecContext(ctx, `
INSERT INTO external_service_sync_jobs (external_service_id, state)
VALUES ($1,$2)
`, svcID, state)
		if err != nil {
			t.Fatal(err)
		}
	}

	// next_sync_at is null, so we expect a sync soon
	assertDue(1*time.Second, true)

	// next_sync_at in the future
	_, err = db.Handle().ExecContext(ctx, "UPDATE external_services SET next_sync_at = $1", now.Add(10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	assertDue(1*time.Second, false)
	assertDue(11*time.Minute, true)

	// With sync jobs
	makeSyncJob(svc1.ID, "queued")
	makeSyncJob(svc2.ID, "completed")
	assertDue(1*time.Minute, true)

	// Sync jobs not running
	_, err = db.Handle().ExecContext(ctx, "UPDATE external_service_sync_jobs SET state = 'completed'")
	if err != nil {
		t.Fatal(err)
	}
	assertDue(1*time.Minute, false)
}

func TestConfigurationHasWebhooks(t *testing.T) {
	t.Run("supported kinds with webhooks", func(t *testing.T) {
		for _, cfg := range []any{
			&schema.GitHubConnection{
				Webhooks: []*schema.GitHubWebhook{
					{Org: "org", Secret: "super secret"},
				},
			},
			&schema.GitLabConnection{
				Webhooks: []*schema.GitLabWebhook{
					{Secret: "super secret"},
				},
			},
			&schema.BitbucketServerConnection{
				Plugin: &schema.BitbucketServerPlugin{
					Webhooks: &schema.BitbucketServerPluginWebhooks{
						Secret: "super secret",
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("%T", cfg), func(t *testing.T) {
				assert.True(t, configurationHasWebhooks(cfg))
			})
		}
	})

	t.Run("supported kinds without webhooks", func(t *testing.T) {
		for _, cfg := range []any{
			&schema.GitHubConnection{},
			&schema.GitLabConnection{},
			&schema.BitbucketServerConnection{},
		} {
			t.Run(fmt.Sprintf("%T", cfg), func(t *testing.T) {
				assert.False(t, configurationHasWebhooks(cfg))
			})
		}
	})

	t.Run("unsupported kinds", func(t *testing.T) {
		for _, cfg := range []any{
			&schema.AWSCodeCommitConnection{},
			&schema.BitbucketCloudConnection{},
			&schema.GitoliteConnection{},
			&schema.PerforceConnection{},
			&schema.PhabricatorConnection{},
			&schema.JVMPackagesConnection{},
			&schema.OtherExternalServiceConnection{},
			nil,
		} {
			t.Run(fmt.Sprintf("%T", cfg), func(t *testing.T) {
				assert.False(t, configurationHasWebhooks(cfg))
			})
		}
	})
}

func TestExternalServiceStore_ListRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
	}
	err := db.ExternalServices().Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// create new user
	user, err := db.Users().Create(ctx,
		NewUser{
			Email:           "alice@example.com",
			Username:        "alice",
			Password:        "password",
			EmailIsVerified: true,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// create new org
	displayName := "Acme org"
	org, err := db.Orgs().Create(ctx, "acme", &displayName)
	if err != nil {
		t.Fatal(err)
	}

	const repoId = 1
	err = db.Repos().Create(ctx, &types.Repo{ID: repoId, Name: "test1"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Handle().ExecContext(ctx, "INSERT INTO external_service_repos (external_service_id, repo_id, clone_url, user_id, org_id) VALUES ($1, $2, $3, $4, $5)",
		es.ID,
		repoId,
		"cloneUrl",
		user.ID,
		org.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	// check that repos are found with empty ExternalServiceReposListOptions
	haveRepos, err := db.ExternalServices().ListRepos(ctx, ExternalServiceReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(haveRepos) != 1 {
		t.Fatalf("Expected 1 external service repo, got %d", len(haveRepos))
	}

	have := haveRepos[0]

	require.Exactly(t, es.ID, have.ExternalServiceID, "externalServiceID is incorrect")
	require.Exactly(t, api.RepoID(repoId), have.RepoID, "repoID is incorrect")
	require.Exactly(t, "cloneUrl", have.CloneURL, "cloneURL is incorrect")
	require.Exactly(t, user.ID, have.UserID, "userID is incorrect")
	require.Exactly(t, org.ID, have.OrgID, "orgID is incorrect")

	// check that repos are found with given externalServiceID
	haveRepos, err = db.ExternalServices().ListRepos(ctx, ExternalServiceReposListOptions{ExternalServiceID: 1, LimitOffset: &LimitOffset{Limit: 1}})
	if err != nil {
		t.Fatal(err)
	}

	if len(haveRepos) != 1 {
		t.Fatalf("Expected 1 external service repo, got %d", len(haveRepos))
	}

	// check that repos are limited
	haveRepos, err = db.ExternalServices().ListRepos(ctx, ExternalServiceReposListOptions{ExternalServiceID: 1, LimitOffset: &LimitOffset{Limit: 0}})
	if err != nil {
		t.Fatal(err)
	}

	if len(haveRepos) != 0 {
		t.Fatalf("Expected 0 external service repos, got %d", len(haveRepos))
	}
}

package database_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestExternalServicesStore_ValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		kind     string
		config   string
		listFunc func(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
		wantErr  string
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
			name: "1 error - Bitbucket.org",
			kind: extsvc.KindBitbucketCloud,
			// Invalid UUID, using + instead of -
			config:  `{"url": "https://bitbucket.org", "username": "ceo", "appPassword": "abc", "exclude": [{"uuid":"{fceb73c7+cef6-4abe-956d-e471281126bd}"}]}`,
			wantErr: `exclude.0.uuid: Does not match pattern '^\{[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\}$'`,
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
			wantErr: "2 errors occurred:\n\t* token: String length must be greater than or equal to 1\n\t* either token or GitHub App Details must be set",
		},
		{
			name:   "no conflicting rate limit",
			kind:   extsvc.KindGitHub,
			config: `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "rateLimit": {"enabled": true, "requestsPerHour": 5000}}`,
			listFunc: func(ctx context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return nil, nil
			},
			wantErr: "<nil>",
		},
		{
			name:    "gjson handles comments",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "token": "abc", "repositoryQuery": ["affiliated"]} // comment`,
			wantErr: "<nil>",
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
		{
			name:    "1 errors - dev.azure.com",
			kind:    extsvc.KindAzureDevOps,
			config:  `{"url": "https://dev.azure.com", "token": "token", "username": "username"}`,
			wantErr: "either 'projects' or 'orgs' must be set",
		},
		{
			name:    "0 errors - dev.azure.com",
			kind:    extsvc.KindAzureDevOps,
			config:  `{"url": "https://dev.azure.com", "token": "token", "username": "username", "projects":[]}`,
			wantErr: "<nil>",
		},
		{
			name:    "0 errors - dev.azure.com",
			kind:    extsvc.KindAzureDevOps,
			config:  `{"url": "https://dev.azure.com", "token": "token", "username": "username", "orgs":[]}`,
			wantErr: "<nil>",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dbm := dbmocks.NewMockDB()
			ess := dbmocks.NewMockExternalServiceStore()
			if test.listFunc != nil {
				ess.ListFunc.SetDefaultHook(test.listFunc)
			}
			dbm.ExternalServicesFunc.SetDefaultReturn(ess)
			_, err := database.ValidateExternalServiceConfig(context.Background(), dbm, database.ValidateExternalServiceConfigOptions{
				Kind:   test.kind,
				Config: test.config,
			})
			gotErr := fmt.Sprintf("%v", err)
			if gotErr != test.wantErr {
				t.Errorf("error: want %q but got %q", test.wantErr, gotErr)
			}
		})
	}
}

package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExternalServicesListOptions_sqlConditions(t *testing.T) {
	tests := []struct {
		name      string
		kinds     []string
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name:      "no kind",
			wantQuery: "deleted_at IS NULL",
		},
		{
			name:      "only one kind: GitHub",
			kinds:     []string{extsvc.KindGitHub},
			wantQuery: "deleted_at IS NULL AND kind IN ($1)",
			wantArgs:  []interface{}{extsvc.KindGitHub},
		},
		{
			name:      "two kinds: GitHub and GitLab",
			kinds:     []string{extsvc.KindGitHub, extsvc.KindGitLab},
			wantQuery: "deleted_at IS NULL AND kind IN ($1 , $2)",
			wantArgs:  []interface{}{extsvc.KindGitHub, extsvc.KindGitLab},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := ExternalServicesListOptions{
				Kinds: test.kinds,
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
	tests := []struct {
		name    string
		kind    string
		config  string
		setup   func(t *testing.T)
		wantErr string
	}{
		{
			name:    "0 errors",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			wantErr: "<nil>",
		},
		{
			name:    "1 error",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": ""}`,
			wantErr: "1 error occurred:\n\t* token: String length must be greater than or equal to 1\n\n",
		},
		{
			name:    "2 errors",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "", "x": 123}`,
			wantErr: "2 errors occurred:\n\t* Additional property x is not allowed\n\t* token: String length must be greater than or equal to 1\n\n",
		},
		{
			name:   "no conflicting rate limit",
			kind:   extsvc.KindGitHub,
			config: `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "rateLimit": {"enabled": true, "requestsPerHour": 5000}}`,
			setup: func(t *testing.T) {
				t.Cleanup(func() {
					Mocks.ExternalServices.List = nil
				})
				Mocks.ExternalServices.List = func(opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
					return nil, nil
				}
			},
			wantErr: "<nil>",
		},
		{
			name:   "conflicting rate limit",
			kind:   extsvc.KindGitHub,
			config: `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "rateLimit": {"enabled": true, "requestsPerHour": 5000}}`,
			setup: func(t *testing.T) {
				t.Cleanup(func() {
					Mocks.ExternalServices.List = nil
				})
				Mocks.ExternalServices.List = func(opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
					return []*types.ExternalService{
						{
							ID:          1,
							Kind:        extsvc.KindGitHub,
							DisplayName: "GITHUB 1",
							Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "rateLimit": {"enabled": true, "requestsPerHour": 5000}}`,
						},
					}, nil
				}
			},
			wantErr: "1 error occurred:\n\t* existing external service, \"GITHUB 1\", already has a rate limit set\n\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(t)
			}

			err := (&ExternalServicesStore{}).ValidateConfig(context.Background(), 0, test.kind, test.config, nil)
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
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
	}
	err := (&ExternalServicesStore{}).Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// Should get back the same one
	got, err := (&ExternalServicesStore{}).GetByID(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(es, got); diff != "" {
		t.Fatalf("(-want +got):\n%s", diff)
	}
}

func TestExternalServicesStore_Update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
	}
	err := (&ExternalServicesStore{}).Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// Update its name and config
	esUpdate := &ExternalServiceUpdate{
		DisplayName: strptr("GITHUB (updated) #1"),
		Config:      strptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
	}
	err = (&ExternalServicesStore{}).Update(ctx, nil, es.ID, esUpdate)
	if err != nil {
		t.Fatal(err)
	}

	// Get and verify update
	got, err := (&ExternalServicesStore{}).GetByID(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(*esUpdate.DisplayName, got.DisplayName); diff != "" {
		t.Fatalf("DisplayName mismatch (-want +got):\n%s", diff)
	} else if diff = cmp.Diff(*esUpdate.Config, got.Config); diff != "" {
		t.Fatalf("Config mismatch (-want +got):\n%s", diff)
	} else if got.UpdatedAt.Equal(es.UpdatedAt) {
		t.Fatalf("UpdateAt: want to be updated but not")
	}
}

func TestExternalServicesStore_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
	}
	err := (&ExternalServicesStore{}).Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// Delete this external service
	err = (&ExternalServicesStore{}).Delete(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Delete again should get externalServiceNotFoundError
	err = (&ExternalServicesStore{}).Delete(ctx, es.ID)
	gotErr := fmt.Sprintf("%v", err)
	wantErr := fmt.Sprintf("external service not found: %v", es.ID)
	if gotErr != wantErr {
		t.Errorf("error: want %q but got %q", wantErr, gotErr)
	}
}

func TestExternalServicesStore_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
	}
	err := (&ExternalServicesStore{}).Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// Should be able to get back by its ID
	_, err = (&ExternalServicesStore{}).GetByID(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Delete this external service
	err = (&ExternalServicesStore{}).Delete(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Should now get externalServiceNotFoundError
	_, err = (&ExternalServicesStore{}).GetByID(ctx, es.ID)
	gotErr := fmt.Sprintf("%v", err)
	wantErr := fmt.Sprintf("external service not found: %v", es.ID)
	if gotErr != wantErr {
		t.Errorf("error: want %q but got %q", wantErr, gotErr)
	}
}

func TestExternalServicesStore_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
	}
	err := (&ExternalServicesStore{}).Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	ess, err := (&ExternalServicesStore{}).List(ctx, ExternalServicesListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(ess) != 1 {
		t.Fatalf("Want 1 external service but got %d", len(ess))
	} else if diff := cmp.Diff(es, ess[0]); diff != "" {
		t.Fatalf("(-want +got):\n%s", diff)
	}
}

func TestExternalServicesStore_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
	}
	err := (&ExternalServicesStore{}).Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	count, err := (&ExternalServicesStore{}).Count(ctx, ExternalServicesListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("Want 1 external service but got %d", count)
	}
}

func TestListConfigs(t *testing.T) {
	t.Run("call SetURN method", func(t *testing.T) {
		Mocks.ExternalServices.List = func(opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
			return []*types.ExternalService{
				{
					ID:          1,
					Kind:        extsvc.KindGitHub,
					DisplayName: "GITHUB #1",
					Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
				},
			}, nil
		}
		defer func() { Mocks.ExternalServices = MockExternalServices{} }()

		var connections []*types.GitHubConnection
		err := ExternalServices.listConfigs(context.Background(), extsvc.KindGitHub, &connections)
		if err != nil {
			t.Fatal(err)
		}

		urns := make([]string, 0, len(connections))
		for _, conn := range connections {
			urns = append(urns, conn.URN)
		}

		wantURNs := []string{"extsvc:github:1"}
		if diff := cmp.Diff(wantURNs, urns); diff != "" {
			t.Fatalf("URNs mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("should not fail when no SetURN method", func(t *testing.T) {
		Mocks.ExternalServices.List = func(opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
			return []*types.ExternalService{
				{
					ID:          1,
					Kind:        extsvc.KindGitHub,
					DisplayName: "GITHUB #1",
					Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
				},
			}, nil
		}
		defer func() { Mocks.ExternalServices = MockExternalServices{} }()

		var connections []*schema.GitHubConnection
		err := ExternalServices.listConfigs(context.Background(), extsvc.KindGitHub, &connections)
		if err != nil {
			t.Fatal(err)
		}
	})
}

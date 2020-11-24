package db

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestExternalServicesListOptions_sqlConditions(t *testing.T) {
	tests := []struct {
		name            string
		noNamespace     bool
		namespaceUserID int32
		kinds           []string
		afterID         int64
		wantQuery       string
		wantArgs        []interface{}
	}{
		{
			name:      "no condition",
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
		{
			name:            "has namespace user ID",
			namespaceUserID: 1,
			wantQuery:       "deleted_at IS NULL AND namespace_user_id = $1",
			wantArgs:        []interface{}{int32(1)},
		},
		{
			name:            "want no namespace",
			noNamespace:     true,
			namespaceUserID: 1,
			wantQuery:       "deleted_at IS NULL AND namespace_user_id IS NULL",
		},
		{
			name:      "has after ID",
			afterID:   10,
			wantQuery: "deleted_at IS NULL AND id < $1",
			wantArgs:  []interface{}{int64(10)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := ExternalServicesListOptions{
				NoNamespace:     test.noNamespace,
				NamespaceUserID: test.namespaceUserID,
				Kinds:           test.kinds,
				AfterID:         test.afterID,
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
		name         string
		kind         string
		config       string
		hasNamespace bool
		setup        func(t *testing.T)
		wantErr      string
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
		{
			name:         "prevent code hosts that are not allowed",
			kind:         extsvc.KindGitHub,
			config:       `{"url": "https://github.example.com", "repositoryQuery": ["none"], "token": "abc"}`,
			hasNamespace: true,
			wantErr:      `users are only allowed to add external service for https://github.com/, https://gitlab.com/ and https://bitbucket.org/`,
		},
		{
			name:         "gjson handles comments",
			kind:         extsvc.KindGitHub,
			config:       `{"url": "https://github.com", "token": "abc", "repositoryQuery": ["affiliated"]} // comment`,
			hasNamespace: true,
			wantErr:      "<nil>",
		},
		{
			name:         "prevent disallowed repositoryPathPattern field",
			kind:         extsvc.KindGitHub,
			config:       `{"url": "https://github.com", "repositoryPathPattern": "github/{nameWithOwner}"}`,
			hasNamespace: true,
			wantErr:      `field "repositoryPathPattern" is not allowed in a user-added external service`,
		},
		{
			name:         "prevent disallowed nameTransformations field",
			kind:         extsvc.KindGitHub,
			config:       `{"url": "https://github.com", "nameTransformations": [{"regex": "\\.d/","replacement": "/"},{"regex": "-git$","replacement": ""}]}`,
			hasNamespace: true,
			wantErr:      `field "nameTransformations" is not allowed in a user-added external service`,
		},
		{
			name:         "prevent disallowed rateLimit field",
			kind:         extsvc.KindGitHub,
			config:       `{"url": "https://github.com", "rateLimit": {}}`,
			hasNamespace: true,
			wantErr:      `field "rateLimit" is not allowed in a user-added external service`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(t)
			}

			err := ExternalServices.ValidateConfig(context.Background(), ValidateExternalServiceConfigOptions{
				Kind:         test.kind,
				Config:       test.config,
				HasNamespace: test.hasNamespace,
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
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	tests := []struct {
		name             string
		externalService  *types.ExternalService
		wantUnrestricted bool
	}{
		{
			name: "without authorization",
			externalService: &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "GITHUB #1",
				Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			},
			wantUnrestricted: true,
		},
		{
			name: "with authorization",
			externalService: &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "GITHUB #2",
				Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
			},
			wantUnrestricted: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ExternalServices.Create(ctx, confGet, test.externalService)
			if err != nil {
				t.Fatal(err)
			}

			// Should get back the same one
			got, err := ExternalServices.GetByID(ctx, test.externalService.ID)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.externalService, got); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}

			if test.wantUnrestricted != got.Unrestricted {
				t.Fatalf("Want unrestricted = %v, but got %v", test.wantUnrestricted, got.Unrestricted)
			}
		})
	}
}

func TestExternalServicesStore_CreateWithTierEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)

	ctx := context.Background()
	confGet := func() *conf.Unified { return &conf.Unified{} }
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
	}
	store := &ExternalServiceStore{
		PreCreateExternalService: func(ctx context.Context) error {
			return errcode.NewPresentationError("test plan limit exceeded")
		},
	}
	if err := store.Create(ctx, confGet, es); err == nil {
		t.Fatal("expected an error, got none")
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
	err := (&ExternalServiceStore{}).Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: The order of tests matters
	tests := []struct {
		name             string
		update           *ExternalServiceUpdate
		wantUnrestricted bool
	}{
		{
			name: "update with authorization",
			update: &ExternalServiceUpdate{
				DisplayName: strptr("GITHUB (updated) #1"),
				Config:      strptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def", "authorization": {}}`),
			},
			wantUnrestricted: false,
		},
		{
			name: "update without authorization",
			update: &ExternalServiceUpdate{
				DisplayName: strptr("GITHUB (updated) #2"),
				Config:      strptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
			},
			wantUnrestricted: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err = ExternalServices.Update(ctx, nil, es.ID, test.update)
			if err != nil {
				t.Fatal(err)
			}

			// Get and verify update
			got, err := ExternalServices.GetByID(ctx, es.ID)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(*test.update.DisplayName, got.DisplayName); diff != "" {
				t.Fatalf("DisplayName mismatch (-want +got):\n%s", diff)
			} else if diff = cmp.Diff(*test.update.Config, got.Config); diff != "" {
				t.Fatalf("Config mismatch (-want +got):\n%s", diff)
			} else if got.UpdatedAt.Equal(es.UpdatedAt) {
				t.Fatalf("UpdateAt: want to be updated but not")
			}

			if test.wantUnrestricted != got.Unrestricted {
				t.Fatalf("Want unrestricted = %v, but got %v", test.wantUnrestricted, got.Unrestricted)
			}
		})
	}
}

func TestExternalServicesStore_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := actor.WithInternalActor(context.Background())

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
	}
	err := ExternalServices.Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// Create two repositories to test trigger of soft-deleting external service:
	//  - ID=1 is expected to be deleted along with deletion of the external service.
	//  - ID=2 remains untouched because it is not associated with the external service.
	_, err = dbconn.Global.ExecContext(ctx, `
INSERT INTO repo (id, name, description, fork)
VALUES (1, 'github.com/user/repo', '', FALSE);
INSERT INTO repo (id, name, description, fork)
VALUES (2, 'github.com/user/repo2', '', FALSE);
`)
	if err != nil {
		t.Fatal(err)
	}

	// Insert a row to `external_service_repos` table to test the trigger.
	q := sqlf.Sprintf(`
INSERT INTO external_service_repos (external_service_id, repo_id, clone_url)
VALUES (%d, 1, '')
`, es.ID)
	_, err = dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}

	// Delete this external service
	err = ExternalServices.Delete(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Delete again should get externalServiceNotFoundError
	err = ExternalServices.Delete(ctx, es.ID)
	gotErr := fmt.Sprintf("%v", err)
	wantErr := fmt.Sprintf("external service not found: %v", es.ID)
	if gotErr != wantErr {
		t.Errorf("error: want %q but got %q", wantErr, gotErr)
	}

	// Should only get back the repo with ID=2
	repos, err := Repos.GetByIDs(ctx, 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	want := []*types.Repo{
		{ID: 2, Name: "github.com/user/repo2"},
	}
	if diff := cmp.Diff(want, repos); diff != "" {
		t.Fatalf("Repos mismatch (-want +got):\n%s", diff)
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
	err := (&ExternalServiceStore{}).Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	// Should be able to get back by its ID
	_, err = (&ExternalServiceStore{}).GetByID(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Delete this external service
	err = (&ExternalServiceStore{}).Delete(ctx, es.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Should now get externalServiceNotFoundError
	_, err = (&ExternalServiceStore{}).GetByID(ctx, es.ID)
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

	// Create test user
	user, err := Users.Create(ctx, NewUser{
		Email:           "alice@example.com",
		Username:        "alice",
		Password:        "password",
		EmailIsVerified: true,
	})
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
			Config:          `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			NamespaceUserID: user.ID,
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "GITHUB #2",
			Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`,
		},
	}
	for _, es := range ess {
		err := ExternalServices.Create(ctx, confGet, es)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("list all external services", func(t *testing.T) {
		got, err := (&ExternalServiceStore{}).List(ctx, ExternalServicesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })

		if diff := cmp.Diff(ess, got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list external services with certain IDs", func(t *testing.T) {
		got, err := (&ExternalServiceStore{}).List(ctx, ExternalServicesListOptions{
			IDs: []int64{ess[1].ID},
		})
		if err != nil {
			t.Fatal(err)
		}
		sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })

		if diff := cmp.Diff(ess[1:], got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list external services with no namespace", func(t *testing.T) {
		got, err := (&ExternalServiceStore{}).List(ctx, ExternalServicesListOptions{
			NoNamespace: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 1 {
			t.Fatalf("Want 1 external service but got %d", len(ess))
		} else if diff := cmp.Diff(ess[1], got[0]); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list only test user's external services", func(t *testing.T) {
		got, err := (&ExternalServiceStore{}).List(ctx, ExternalServicesListOptions{
			NamespaceUserID: user.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 1 {
			t.Fatalf("Want 1 external service but got %d", len(ess))
		} else if diff := cmp.Diff(ess[0], got[0]); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list non-exist user's external services", func(t *testing.T) {
		ess, err := (&ExternalServiceStore{}).List(ctx, ExternalServicesListOptions{
			NamespaceUserID: 404,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(ess) != 0 {
			t.Fatalf("Want 0 external service but got %d", len(ess))
		}
	})
}

func TestExternalServicesStore_DistinctKinds(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	t.Run("no external service won't blow up", func(t *testing.T) {
		kinds, err := ExternalServices.DistinctKinds(ctx)
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
			Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "GITHUB #2",
			Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`,
		},
		{
			Kind:        extsvc.KindGitLab,
			DisplayName: "GITLAB #1",
			Config:      `{"url": "https://github.com", "projectQuery": ["none"], "token": "abc"}`,
		},
		{
			Kind:        extsvc.KindOther,
			DisplayName: "OTHER #1",
			Config:      `{"repos": []}`,
		},
	}
	for _, es := range ess {
		err := ExternalServices.Create(ctx, confGet, es)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Delete the last external service which should be excluded from the result
	err := ExternalServices.Delete(ctx, ess[3].ID)
	if err != nil {
		t.Fatal(err)
	}

	kinds, err := ExternalServices.DistinctKinds(ctx)
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
	err := (&ExternalServiceStore{}).Create(ctx, confGet, es)
	if err != nil {
		t.Fatal(err)
	}

	count, err := (&ExternalServiceStore{}).Count(ctx, ExternalServicesListOptions{})
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
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	clock := dbtesting.NewFakeClock(time.Now(), 0)

	svcs := types.MakeExternalServices()

	t.Run("no external services", func(t *testing.T) {
		if err := ExternalServices.Upsert(ctx); err != nil {
			t.Fatalf("Upsert error: %s", err)
		}
	})

	t.Run("many external services", func(t *testing.T) {
		tx, err := ExternalServices.Transact(ctx)
		if err != nil {
			t.Fatalf("Transact error: %s", err)
		}
		defer func() {
			err = tx.Done(err)
			if err != nil {
				t.Fatalf("Done error: %s", err)
			}
		}()

		want := types.GenerateExternalServices(7, svcs...)

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
			Kinds: svcs.Kinds(),
		})
		if err != nil {
			t.Fatalf("List error: %s", err)
		}

		sort.Sort(types.ExternalServices(have))

		if diff := cmp.Diff(have, []*types.ExternalService(want), cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("List:\n%s", diff)
		}

		now := clock.Now()
		suffix := "-updated"
		for _, r := range want {
			r.DisplayName += suffix
			r.Kind += suffix
			r.Config += suffix
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

		if diff := cmp.Diff(have, []*types.ExternalService(want), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("List:\n%s", diff)
		}

		want.Apply(func(e *types.ExternalService) {
			e.UpdatedAt = now
			e.DeletedAt = now
		})

		if err = tx.Upsert(ctx, want.Clone()...); err != nil {
			t.Errorf("Upsert error: %s", err)
		}
		have, err = tx.List(ctx, ExternalServicesListOptions{})
		if err != nil {
			t.Errorf("List error: %s", err)
		}

		sort.Sort(types.ExternalServices(have))

		if diff := cmp.Diff(have, []*types.ExternalService(nil), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("List:\n%s", diff)
		}
	})
}

func TestExternalServicesStore_ValidateSingleGlobalConnection(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	now := time.Now()

	makeService := func(global bool) *types.ExternalService {
		cfg := `{"url": "https://github.com", "token": "abc", "repositoryQuery": ["none"]}`
		if global {
			cfg = `{"url": "https://github.com", "token": "abc", "repositoryQuery": ["none"], "cloudGlobal": true}`
		}
		svc := &types.ExternalService{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test",
			Config:      cfg,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		return svc
	}

	t.Run("non global", func(t *testing.T) {
		gh := makeService(false)
		if err := ExternalServices.Upsert(ctx, gh); err != nil {
			t.Fatalf("Upsert error: %s", err)
		}
		if err := ExternalServices.validateSingleGlobalConnection(ctx, gh.ID, extsvc.KindGitHub); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("first global", func(t *testing.T) {
		gh := makeService(true)
		if err := ExternalServices.Upsert(ctx, gh); err != nil {
			t.Fatalf("Upsert error: %s", err)
		}
		if err := ExternalServices.validateSingleGlobalConnection(ctx, gh.ID, extsvc.KindGitHub); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("second global", func(t *testing.T) {
		gh := makeService(true)
		if err := ExternalServices.Upsert(ctx, gh); err != nil {
			t.Fatalf("Upsert error: %s", err)
		}
		if err := ExternalServices.validateSingleGlobalConnection(ctx, gh.ID, extsvc.KindGitHub); err == nil {
			t.Fatal("Expected validation error")
		}
	})
}

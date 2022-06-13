package database

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

/*
 * Helpers
 */

func sortedRepoNames(repos []*types.Repo) []api.RepoName {
	names := repoNames(repos)
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names
}

func repoNames(repos []*types.Repo) []api.RepoName {
	var names []api.RepoName
	for _, repo := range repos {
		names = append(names, repo.Name)
	}
	return names
}

func createRepo(ctx context.Context, t *testing.T, db DB, repo *types.Repo) {
	t.Helper()

	op := InsertRepoOp{
		Name:         repo.Name,
		Private:      repo.Private,
		ExternalRepo: repo.ExternalRepo,
		Description:  repo.Description,
		Fork:         repo.Fork,
		Archived:     repo.Archived,
	}

	if err := upsertRepo(ctx, db, op); err != nil {
		t.Fatal(err)
	}
}

func mustCreate(ctx context.Context, t *testing.T, db DB, repo *types.Repo) []*types.Repo {
	t.Helper()

	return mustCreateGitserverRepo(ctx, t, db, repo, types.GitserverRepo{
		CloneStatus: types.CloneStatusNotCloned,
	})
}

func mustCreateGitserverRepo(ctx context.Context, t *testing.T, db DB, repo *types.Repo, gitserver types.GitserverRepo) []*types.Repo {
	t.Helper()

	var createdRepos []*types.Repo
	createRepo(ctx, t, db, repo)
	repo, err := db.Repos().GetByName(ctx, repo.Name)
	if err != nil {
		t.Fatal(err)
	}
	createdRepos = append(createdRepos, repo)

	gitserver.RepoID = repo.ID
	if gitserver.ShardID == "" {
		gitserver.ShardID = "test"
	}

	// Add a row in gitserver_repos
	if err := db.GitserverRepos().Upsert(ctx, &gitserver); err != nil {
		t.Fatal(err)
	}

	return createdRepos
}

func repoNamesFromRepos(repos []*types.Repo) []types.MinimalRepo {
	rnames := make([]types.MinimalRepo, 0, len(repos))
	for _, repo := range repos {
		rnames = append(rnames, types.MinimalRepo{ID: repo.ID, Name: repo.Name})
	}

	return rnames
}

func reposFromRepoNames(names []types.MinimalRepo) []*types.Repo {
	repos := make([]*types.Repo, 0, len(names))
	for _, name := range names {
		repos = append(repos, &types.Repo{ID: name.ID, Name: name.Name})
	}

	return repos
}

// InsertRepoOp represents an operation to insert a repository.
type InsertRepoOp struct {
	Name         api.RepoName
	Description  string
	Fork         bool
	Archived     bool
	Private      bool
	ExternalRepo api.ExternalRepoSpec
}

const upsertSQL = `
WITH upsert AS (
  UPDATE repo
  SET
    name                  = $1,
    description           = $2,
    fork                  = $3,
    external_id           = NULLIF(BTRIM($4), ''),
    external_service_type = NULLIF(BTRIM($5), ''),
    external_service_id   = NULLIF(BTRIM($6), ''),
    archived              = $7,
    private               = $8
  WHERE name = $1 OR (
    external_id IS NOT NULL
    AND external_service_type IS NOT NULL
    AND external_service_id IS NOT NULL
    AND NULLIF(BTRIM($4), '') IS NOT NULL
    AND NULLIF(BTRIM($5), '') IS NOT NULL
    AND NULLIF(BTRIM($6), '') IS NOT NULL
    AND external_id = NULLIF(BTRIM($4), '')
    AND external_service_type = NULLIF(BTRIM($5), '')
    AND external_service_id = NULLIF(BTRIM($6), '')
  )
  RETURNING repo.name
)

INSERT INTO repo (
  name,
  description,
  fork,
  external_id,
  external_service_type,
  external_service_id,
  archived,
  private
) (
  SELECT
    $1 AS name,
    $2 AS description,
    $3 AS fork,
    NULLIF(BTRIM($4), '') AS external_id,
    NULLIF(BTRIM($5), '') AS external_service_type,
    NULLIF(BTRIM($6), '') AS external_service_id,
    $7 AS archived,
    $8 AS private
  WHERE NOT EXISTS (SELECT 1 FROM upsert)
)`

// upsertRepo updates the repository if it already exists (keyed on name) and
// inserts it if it does not.
func upsertRepo(ctx context.Context, db DB, op InsertRepoOp) error {
	s := db.Repos()
	insert := false

	// We optimistically assume the repo is already in the table, so first
	// check if it is. We then fallback to the upsert functionality. The
	// upsert is logged as a modification to the DB, even if it is a no-op. So
	// we do this check to avoid log spam if postgres is configured with
	// log_statement='mod'.
	r, err := s.GetByName(ctx, op.Name)
	if err != nil {
		if !errors.HasType(err, &RepoNotFoundErr{}) {
			return err
		}
		insert = true // missing
	} else {
		insert = (op.Description != r.Description) ||
			(op.Fork != r.Fork) ||
			(!op.ExternalRepo.Equal(&r.ExternalRepo))
	}

	if !insert {
		return nil
	}

	_, err = s.Handle().DBUtilDB().ExecContext(
		ctx,
		upsertSQL,
		op.Name,
		op.Description,
		op.Fork,
		op.ExternalRepo.ID,
		op.ExternalRepo.ServiceType,
		op.ExternalRepo.ServiceID,
		op.Archived,
		op.Private,
	)

	return err
}

/*
 * Tests
 */

func TestRepos_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	now := time.Now()

	service := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternalServices().Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}

	want := mustCreate(ctx, t, db, &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "r",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "name",
		Private:     true,
		URI:         "uri",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			service.URN(): {
				ID:       service.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	})

	repo, err := db.Repos().Get(ctx, want[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(t, repo, want[0]) {
		t.Errorf("got %v, want %v", repo, want[0])
	}
}

func TestRepos_GetByIDs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	want := mustCreate(ctx, t, db, &types.Repo{
		Name: "r",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "a",
			ServiceType: "b",
			ServiceID:   "c",
		},
	})

	repos, err := db.Repos().GetByIDs(ctx, want[0].ID, 404)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("got %d repos, but want 1", len(repos))
	}

	if !jsonEqual(t, repos[0], want[0]) {
		t.Errorf("got %v, want %v", repos[0], want[0])
	}
}

func TestRepos_GetByIDs_EmptyIDs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	repos, err := db.Repos().GetByIDs(ctx, []api.RepoID{}...)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 0 {
		t.Fatalf("got %d repos, but want 0", len(repos))
	}

}

func TestRepos_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	now := time.Now()

	service := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternalServices().Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}

	want := mustCreate(ctx, t, db, &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "r",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "name",
		Private:     true,
		URI:         "uri",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			service.URN(): {
				ID:       service.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	})

	repos, err := db.Repos().List(ctx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(t, repos, want) {
		t.Errorf("got %v, want %v", repos, want)
	}
}

func TestRepos_ListMinimalRepos_userID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	// Create a user
	user, err := db.Users().Create(ctx, NewUser{
		Email:                 "a1@example.com",
		Username:              "u1",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	require.NoError(t, err)

	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: user.ID,
	})

	now := time.Now()
	confGet := func() *conf.Unified { return &conf.Unified{} }
	externalServices := db.ExternalServices()
	repos := db.Repos()

	// Create a user-owned external service and its repository
	userExternalService := types.ExternalService{
		Kind:            extsvc.KindGitHub,
		DisplayName:     "Github - User-owned",
		Config:          `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
		CreatedAt:       now,
		UpdatedAt:       now,
		NamespaceUserID: user.ID,
	}
	err = externalServices.Create(ctx, confGet, &userExternalService)
	require.NoError(t, err)

	userRepo := &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "github.com/sourcegraph/user",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "github.com/sourcegraph/user",
		Private:     true,
		URI:         "github.com/sourcegraph/user",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			userExternalService.URN(): {
				ID:       userExternalService.URN(),
				CloneURL: "git@github.com:sourcegraph/user.git",
			},
		},
	}
	err = repos.Create(ctx, userRepo)
	require.NoError(t, err)

	// Create a site-owned external service and its repository
	siteExternalService := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Site-owned",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err = externalServices.Create(ctx, confGet, &siteExternalService)
	require.NoError(t, err)

	siteRepo := &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "github.com/sourcegraph/site",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "github.com/sourcegraph/site",
		Private:     true,
		URI:         "github.com/sourcegraph/site",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			siteExternalService.URN(): {
				ID:       siteExternalService.URN(),
				CloneURL: "git@github.com:sourcegraph/site.git",
			},
		},
	}
	err = repos.Create(ctx, siteRepo)
	require.NoError(t, err)

	have, err := repos.ListMinimalRepos(ctx, ReposListOptions{UserID: user.ID})
	require.NoError(t, err)

	want := []types.MinimalRepo{
		{ID: userRepo.ID, Name: userRepo.Name},
	}
	assert.Equal(t, want, have)
}

func TestRepos_ListMinimalRepos_orgID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	// Create an org
	displayName := "Acme Corp"
	org, err := db.Orgs().Create(ctx, "acme", &displayName)
	require.NoError(t, err)

	now := time.Now()
	confGet := func() *conf.Unified { return &conf.Unified{} }
	externalServices := db.ExternalServices()
	repos := db.Repos()

	// Create an org-owned external service and its repository
	orgExternalService := types.ExternalService{
		Kind:           extsvc.KindGitHub,
		DisplayName:    "Github - Org-owned",
		Config:         `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
		CreatedAt:      now,
		UpdatedAt:      now,
		NamespaceOrgID: org.ID,
	}
	err = externalServices.Create(ctx, confGet, &orgExternalService)
	require.NoError(t, err)

	repo := &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "github.com/sourcegraph/org",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "github.com/sourcegraph/org",
		Private:     true,
		URI:         "github.com/sourcegraph/org",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			orgExternalService.URN(): {
				ID:       orgExternalService.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	}
	err = repos.Create(ctx, repo)
	require.NoError(t, err)

	// Create a site-owned external service and its repository
	siteExternalService := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Site-owned",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err = externalServices.Create(ctx, confGet, &siteExternalService)
	require.NoError(t, err)

	siteRepo := &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "github.com/sourcegraph/site",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "github.com/sourcegraph/site",
		Private:     true,
		URI:         "github.com/sourcegraph/site",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			siteExternalService.URN(): {
				ID:       siteExternalService.URN(),
				CloneURL: "git@github.com:sourcegraph/site.git",
			},
		},
	}
	err = repos.Create(ctx, siteRepo)
	require.NoError(t, err)

	have, err := repos.ListMinimalRepos(ctx, ReposListOptions{OrgID: org.ID})
	require.NoError(t, err)

	want := []types.MinimalRepo{
		{ID: repo.ID, Name: repo.Name},
	}
	assert.Equal(t, want, have)
}

func TestRepos_List_fork(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	mine := mustCreate(ctx, t, db, &types.Repo{Name: "a/r", Fork: false})
	yours := mustCreate(ctx, t, db, &types.Repo{Name: "b/r", Fork: true})

	{
		repos, err := db.Repos().List(ctx, ReposListOptions{OnlyForks: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, yours, repos)
	}
	{
		repos, err := db.Repos().List(ctx, ReposListOptions{NoForks: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, mine, repos)
	}
	{
		repos, err := db.Repos().List(ctx, ReposListOptions{NoForks: true, OnlyForks: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, nil, repos)
	}
	{
		repos, err := db.Repos().List(ctx, ReposListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, append(append([]*types.Repo(nil), mine...), yours...), repos)
	}
}

func TestRepos_List_FailedSync(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	created := mustCreateGitserverRepo(ctx, t, db, &types.Repo{Name: "repo1"}, types.GitserverRepo{CloneStatus: types.CloneStatusCloned})
	assertCount := func(t *testing.T, opts ReposListOptions, want int) {
		t.Helper()
		count, err := db.Repos().Count(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Fatalf("Expected %d repos, got %d", want, count)
		}
	}
	assertCount(t, ReposListOptions{}, 1)
	assertCount(t, ReposListOptions{FailedFetch: true}, 0)

	repo := created[0]
	if err := db.GitserverRepos().SetLastError(ctx, repo.Name, "Oops", "test"); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ReposListOptions{}, 1)
}

func TestRepos_List_cloned(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	mine := mustCreateGitserverRepo(ctx, t, db, &types.Repo{Name: "a/r"}, types.GitserverRepo{CloneStatus: types.CloneStatusNotCloned})
	yours := mustCreateGitserverRepo(ctx, t, db, &types.Repo{Name: "b/r"}, types.GitserverRepo{CloneStatus: types.CloneStatusCloned})

	tests := []struct {
		name string
		opt  ReposListOptions
		want []*types.Repo
	}{
		{"OnlyCloned", ReposListOptions{OnlyCloned: true}, yours},
		{"NoCloned", ReposListOptions{NoCloned: true}, mine},
		{"NoCloned && OnlyCloned", ReposListOptions{NoCloned: true, OnlyCloned: true}, nil},
		{"Default", ReposListOptions{}, append(append([]*types.Repo(nil), mine...), yours...)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, test.want, repos)
		})
	}
}

func TestRepos_List_LastChanged(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	repos := db.Repos()

	// Insert a repo which should never be returned since we always specify
	// OnlyCloned.
	if err := upsertRepo(ctx, db, InsertRepoOp{Name: "not-on-gitserver"}); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	mustCreateGitserverRepo(ctx, t, db, &types.Repo{Name: "old"}, types.GitserverRepo{
		CloneStatus: types.CloneStatusCloned,
		LastChanged: now.Add(-time.Hour),
	})
	mustCreateGitserverRepo(ctx, t, db, &types.Repo{Name: "new"}, types.GitserverRepo{
		CloneStatus: types.CloneStatusCloned,
		LastChanged: now,
	})

	// Our test helpers don't do updated_at, so manually doing it.
	_, err := db.Handle().DBUtilDB().ExecContext(ctx, "update repo set updated_at = $1", now.Add(-24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// will have update_at set to now, so should be included as often as new.
	mustCreateGitserverRepo(ctx, t, db, &types.Repo{Name: "newMeta"}, types.GitserverRepo{
		CloneStatus: types.CloneStatusCloned,
		LastChanged: now.Add(-24 * time.Hour),
	})
	_, err = db.Handle().DBUtilDB().ExecContext(ctx, "update repo set updated_at = $1 where name = 'newMeta'", now)
	if err != nil {
		t.Fatal(err)
	}

	// we create two search contexts, with one being updated recently only
	// including "newSearchContext".
	mustCreateGitserverRepo(ctx, t, db, &types.Repo{Name: "newSearchContext"}, types.GitserverRepo{
		CloneStatus: types.CloneStatusCloned,
		LastChanged: now.Add(-24 * time.Hour),
	})
	{
		mkSearchContext := func(name string, opts ReposListOptions) {
			t.Helper()
			var revs []*types.SearchContextRepositoryRevisions
			err := repos.StreamMinimalRepos(ctx, opts, func(repo *types.MinimalRepo) {
				revs = append(revs, &types.SearchContextRepositoryRevisions{
					Repo:      *repo,
					Revisions: []string{"HEAD"},
				})
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.SearchContexts().CreateSearchContextWithRepositoryRevisions(ctx, &types.SearchContext{Name: name}, revs)
			if err != nil {
				t.Fatal(err)
			}
		}
		mkSearchContext("old", ReposListOptions{})
		_, err = db.Handle().DBUtilDB().ExecContext(ctx, "update search_contexts set updated_at = $1", now.Add(-24*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		mkSearchContext("new", ReposListOptions{
			Names: []string{"newSearchContext"},
		})
	}

	tests := []struct {
		Name           string
		MinLastChanged time.Time
		Want           []string
	}{{
		Name: "not specified",
		Want: []string{"old", "new", "newMeta", "newSearchContext"},
	}, {
		Name:           "old",
		MinLastChanged: now.Add(-24 * time.Hour),
		Want:           []string{"old", "new", "newMeta", "newSearchContext"},
	}, {
		Name:           "new",
		MinLastChanged: now.Add(-time.Minute),
		Want:           []string{"new", "newMeta", "newSearchContext"},
	}, {
		Name:           "none",
		MinLastChanged: now.Add(time.Minute),
	}}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			repos, err := repos.List(ctx, ReposListOptions{
				OnlyCloned:     true,
				MinLastChanged: test.MinLastChanged,
			})
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, r := range repos {
				got = append(got, string(r.Name))
			}
			if d := cmp.Diff(test.Want, got); d != "" {
				t.Fatalf("mismatch (-want, +got):\n%s", d)
			}
		})
	}
}

func TestRepos_List_ids(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	mine := types.Repos(mustCreate(ctx, t, db, typestest.MakeGithubRepo()))
	mine = append(mine, mustCreate(ctx, t, db, typestest.MakeGitlabRepo())...)

	yours := types.Repos(mustCreate(ctx, t, db, typestest.MakeGitoliteRepo()))
	all := append(mine, yours...)

	tests := []struct {
		name string
		opt  ReposListOptions
		want []*types.Repo
	}{
		{"Subset", ReposListOptions{IDs: mine.IDs()}, mine},
		{"All", ReposListOptions{IDs: all.IDs()}, all},
		{"Default", ReposListOptions{}, all},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, test.want, repos)
		})
	}
}

func TestRepos_List_pagination(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "r1"},
		{Name: "r2"},
		{Name: "r3"},
	}
	for _, repo := range createdRepos {
		mustCreate(ctx, t, db, repo)
	}

	type testcase struct {
		limit  int
		offset int
		exp    []api.RepoName
	}
	tests := []testcase{
		{limit: 1, offset: 0, exp: []api.RepoName{"r1"}},
		{limit: 1, offset: 1, exp: []api.RepoName{"r2"}},
		{limit: 1, offset: 2, exp: []api.RepoName{"r3"}},
		{limit: 2, offset: 0, exp: []api.RepoName{"r1", "r2"}},
		{limit: 2, offset: 2, exp: []api.RepoName{"r3"}},
		{limit: 3, offset: 0, exp: []api.RepoName{"r1", "r2", "r3"}},
		{limit: 3, offset: 3, exp: nil},
		{limit: 4, offset: 0, exp: []api.RepoName{"r1", "r2", "r3"}},
		{limit: 4, offset: 4, exp: nil},
	}
	for _, test := range tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{LimitOffset: &LimitOffset{Limit: test.limit, Offset: test.offset}})
		if err != nil {
			t.Fatal(err)
		}
		if got := sortedRepoNames(repos); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("for test case %v, got %v (want %v)", test, got, test.exp)
		}
	}
}

// TestRepos_List_query tests the behavior of Repos.List when called with
// a query.
// Test batch 1 (correct filtering)
func TestRepos_List_query1(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "abc/def"},
		{Name: "def/ghi"},
		{Name: "jkl/mno/pqr"},
		{Name: "github.com/abc/xyz"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query string
		want  []api.RepoName
	}{
		{"def", []api.RepoName{"abc/def", "def/ghi"}},
		{"ABC/DEF", []api.RepoName{"abc/def"}},
		{"xyz", []api.RepoName{"github.com/abc/xyz"}},
		{"mno/p", []api.RepoName{"jkl/mno/pqr"}},
	}
	for _, test := range tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{Query: test.query})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%q: got repos %q, want %q", test.query, got, test.want)
		}
	}
}

// Test batch 2 (correct ranking)
func TestRepos_List_correct_ranking(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "a/def"},
		{Name: "b/def"},
		{Name: "c/def"},
		{Name: "def/ghi"},
		{Name: "def/jkl"},
		{Name: "def/mno"},
		{Name: "abc/m"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query string
		want  []api.RepoName
	}{
		{"def", []api.RepoName{"a/def", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"}},
		{"b/def", []api.RepoName{"b/def"}},
		{"def/", []api.RepoName{"def/ghi", "def/jkl", "def/mno"}},
		{"def/m", []api.RepoName{"def/mno"}},
	}
	for _, test := range tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{Query: test.query})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("Unexpected repo result for query %q:\ngot:  %q\nwant: %q", test.query, got, test.want)
		}
	}
}

// Test sort
func TestRepos_List_sort(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "c/def"},
		{Name: "def/mno"},
		{Name: "b/def"},
		{Name: "abc/m"},
		{Name: "abc/def"},
		{Name: "def/jkl"},
		{Name: "def/ghi"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query   string
		orderBy RepoListOrderBy
		want    []api.RepoName
	}{
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListName,
			}},
			want: []api.RepoName{"abc/def", "abc/m", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListCreatedAt,
			}},
			want: []api.RepoName{"c/def", "def/mno", "b/def", "abc/m", "abc/def", "def/jkl", "def/ghi"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCreatedAt,
				Descending: true,
			}},
			want: []api.RepoName{"def/ghi", "def/jkl", "abc/def", "abc/m", "b/def", "def/mno", "c/def"},
		},
		{
			query: "def",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCreatedAt,
				Descending: true,
			}},
			want: []api.RepoName{"def/ghi", "def/jkl", "abc/def", "b/def", "def/mno", "c/def"},
		},
	}
	for _, test := range tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{Query: test.query, OrderBy: test.orderBy})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("Unexpected repo result for query %q, orderBy %v:\ngot:  %q\nwant: %q", test.query, test.orderBy, got, test.want)
		}
	}
}

// TestRepos_List_patterns tests the behavior of Repos.List when called with
// IncludePatterns and ExcludePattern.
func TestRepos_List_patterns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "a/b"},
		{Name: "c/d"},
		{Name: "e/f"},
		{Name: "g/h"},
		{Name: "I/J"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, db, repo)
	}
	tests := []struct {
		includePatterns []string
		excludePattern  string
		caseSensitive   bool
		want            []api.RepoName
	}{
		{
			includePatterns: []string{"(a|c)"},
			want:            []api.RepoName{"a/b", "c/d"},
		},
		{
			includePatterns: []string{"(a|c)", "b"},
			want:            []api.RepoName{"a/b"},
		},
		{
			includePatterns: []string{"(a|c)"},
			excludePattern:  "d",
			want:            []api.RepoName{"a/b"},
		},
		{
			excludePattern: "(d|e)",
			want:           []api.RepoName{"a/b", "g/h", "I/J"},
		},
		{
			includePatterns: []string{"(A|c|I)"},
			want:            []api.RepoName{"a/b", "c/d", "I/J"},
		},
		{
			includePatterns: []string{"I", "J"},
			caseSensitive:   true,
			want:            []api.RepoName{"I/J"},
		},
	}
	for _, test := range tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{
			IncludePatterns:       test.includePatterns,
			ExcludePattern:        test.excludePattern,
			CaseSensitivePatterns: test.caseSensitive,
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("include %q exclude %q: got repos %q, want %q", test.includePatterns, test.excludePattern, got, test.want)
		}
	}
}

func TestRepos_List_queryAndPatternsMutuallyExclusive(t *testing.T) {
	ctx := actor.WithInternalActor(context.Background())
	wantErr := "Query and IncludePatterns/ExcludePattern options are mutually exclusive"

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))

	t.Run("Query and IncludePatterns", func(t *testing.T) {
		_, err := db.Repos().List(ctx, ReposListOptions{Query: "x", IncludePatterns: []string{"y"}})
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("got error %v, want it to contain %q", err, wantErr)
		}
	})

	t.Run("Query and ExcludePattern", func(t *testing.T) {
		_, err := db.Repos().List(ctx, ReposListOptions{Query: "x", ExcludePattern: "y"})
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("got error %v, want it to contain %q", err, wantErr)
		}
	})
}

func TestRepos_createRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	// Add a repo.
	createRepo(ctx, t, db, &types.Repo{
		Name:        "a/b",
		Description: "test",
	})

	repo, err := db.Repos().GetByName(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := repo.Name, api.RepoName("a/b"); got != want {
		t.Fatalf("got Name %q, want %q", got, want)
	}
	if got, want := repo.Description, "test"; got != want {
		t.Fatalf("got Description %q, want %q", got, want)
	}
}

func TestRepos_List_useOr(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	archived := types.Repos{typestest.MakeGitlabRepo()}.With(func(r *types.Repo) { r.Archived = true })
	archived = mustCreate(ctx, t, db, archived[0])
	forks := types.Repos{typestest.MakeGitoliteRepo()}.With(func(r *types.Repo) { r.Fork = true })
	forks = mustCreate(ctx, t, db, forks[0])
	cloned := types.Repos{typestest.MakeGithubRepo()}
	cloned = mustCreateGitserverRepo(ctx, t, db, cloned[0], types.GitserverRepo{CloneStatus: types.CloneStatusCloned})

	archivedAndForks := append(archived, forks...)
	sort.Sort(archivedAndForks)
	all := append(archivedAndForks, cloned...)
	sort.Sort(all)

	tests := []struct {
		name string
		opt  ReposListOptions
		want []*types.Repo
	}{
		{"Archived or Forks", ReposListOptions{OnlyArchived: true, OnlyForks: true, UseOr: true}, archivedAndForks},
		{"Archived or Forks Or Cloned", ReposListOptions{OnlyArchived: true, OnlyForks: true, OnlyCloned: true, UseOr: true}, all},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, test.want, repos)
		})
	}
}

func TestRepos_List_externalServiceID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	services := typestest.MakeExternalServices()
	service1 := services[0]
	service2 := services[1]
	if err := db.ExternalServices().Create(ctx, confGet, service1); err != nil {
		t.Fatal(err)
	}
	if err := db.ExternalServices().Create(ctx, confGet, service2); err != nil {
		t.Fatal(err)
	}

	mine := types.Repos{typestest.MakeGithubRepo(service1)}
	if err := db.Repos().Create(ctx, mine...); err != nil {
		t.Fatal(err)
	}

	yours := types.Repos{typestest.MakeGitlabRepo(service2)}
	if err := db.Repos().Create(ctx, yours...); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		opt  ReposListOptions
		want []*types.Repo
	}{
		{"Some", ReposListOptions{ExternalServiceIDs: []int64{service1.ID}}, mine},
		{"Default", ReposListOptions{}, append(mine, yours...)},
		{"NonExistant", ReposListOptions{ExternalServiceIDs: []int64{1000}}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, test.want, repos)
		})
	}
}

func TestRepos_ListMinimalRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	now := time.Now()

	service := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternalServices().Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}

	repo := mustCreate(ctx, t, db, &types.Repo{
		Name: "name",
	})
	want := []types.MinimalRepo{{ID: repo[0].ID, Name: repo[0].Name}}

	repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(t, repos, want) {
		t.Errorf("got %v, want %v", repos, want)
	}
}

func TestRepos_ListMinimalRepos_fork(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	mine := repoNamesFromRepos(mustCreate(ctx, t, db, &types.Repo{Name: "a/r", Fork: false}))
	yours := repoNamesFromRepos(mustCreate(ctx, t, db, &types.Repo{Name: "b/r", Fork: true}))

	{
		repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{OnlyForks: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, yours, repos)
	}
	{
		repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{NoForks: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, mine, repos)
	}
	{
		repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{NoForks: true, OnlyForks: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, nil, repos)
	}
	{
		repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, append(append([]types.MinimalRepo(nil), mine...), yours...), repos)
	}
}

func TestRepos_ListMinimalRepos_cloned(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	mine := repoNamesFromRepos(mustCreate(ctx, t, db, &types.Repo{Name: "a/r"}))
	yours := repoNamesFromRepos(mustCreateGitserverRepo(ctx, t, db, &types.Repo{Name: "b/r"}, types.GitserverRepo{CloneStatus: types.CloneStatusCloned}))

	tests := []struct {
		name string
		opt  ReposListOptions
		want []types.MinimalRepo
	}{
		{"OnlyCloned", ReposListOptions{OnlyCloned: true}, yours},
		{"NoCloned", ReposListOptions{NoCloned: true}, mine},
		{"NoCloned && OnlyCloned", ReposListOptions{NoCloned: true, OnlyCloned: true}, nil},
		{"Default", ReposListOptions{}, append(append([]types.MinimalRepo(nil), mine...), yours...)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos, err := db.Repos().ListMinimalRepos(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, test.want, repos)
		})
	}
}

func TestRepos_ListMinimalRepos_ids(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	mine := types.Repos(mustCreate(ctx, t, db, typestest.MakeGithubRepo()))
	mine = append(mine, mustCreate(ctx, t, db, typestest.MakeGitlabRepo())...)

	yours := types.Repos(mustCreate(ctx, t, db, typestest.MakeGitoliteRepo()))
	all := append(mine, yours...)

	tests := []struct {
		name string
		opt  ReposListOptions
		want []types.MinimalRepo
	}{
		{"Subset", ReposListOptions{IDs: mine.IDs()}, repoNamesFromRepos(mine)},
		{"All", ReposListOptions{IDs: all.IDs()}, repoNamesFromRepos(all)},
		{"Default", ReposListOptions{}, repoNamesFromRepos(all)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos, err := db.Repos().ListMinimalRepos(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, test.want, repos)
		})
	}
}

func TestRepos_ListMinimalRepos_pagination(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "r1"},
		{Name: "r2"},
		{Name: "r3"},
	}
	for _, repo := range createdRepos {
		mustCreate(ctx, t, db, repo)
	}

	type testcase struct {
		limit  int
		offset int
		exp    []api.RepoName
	}
	tests := []testcase{
		{limit: 1, offset: 0, exp: []api.RepoName{"r1"}},
		{limit: 1, offset: 1, exp: []api.RepoName{"r2"}},
		{limit: 1, offset: 2, exp: []api.RepoName{"r3"}},
		{limit: 2, offset: 0, exp: []api.RepoName{"r1", "r2"}},
		{limit: 2, offset: 2, exp: []api.RepoName{"r3"}},
		{limit: 3, offset: 0, exp: []api.RepoName{"r1", "r2", "r3"}},
		{limit: 3, offset: 3, exp: nil},
		{limit: 4, offset: 0, exp: []api.RepoName{"r1", "r2", "r3"}},
		{limit: 4, offset: 4, exp: nil},
	}
	for _, test := range tests {
		repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{LimitOffset: &LimitOffset{Limit: test.limit, Offset: test.offset}})
		if err != nil {
			t.Fatal(err)
		}
		if got := sortedRepoNames(reposFromRepoNames(repos)); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("for test case %v, got %v (want %v)", test, got, test.exp)
		}
	}
}

// TestRepos_ListMinimalRepos_query tests the behavior of Repos.ListMinimalRepos when called with
// a query.
// Test batch 1 (correct filtering)
func TestRepos_ListMinimalRepos_correctFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "abc/def"},
		{Name: "def/ghi"},
		{Name: "jkl/mno/pqr"},
		{Name: "github.com/abc/xyz"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query string
		want  []api.RepoName
	}{
		{"def", []api.RepoName{"abc/def", "def/ghi"}},
		{"ABC/DEF", []api.RepoName{"abc/def"}},
		{"xyz", []api.RepoName{"github.com/abc/xyz"}},
		{"mno/p", []api.RepoName{"jkl/mno/pqr"}},
	}
	for _, test := range tests {
		repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{Query: test.query})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(reposFromRepoNames(repos)); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%q: got repos %q, want %q", test.query, got, test.want)
		}
	}
}

// Test batch 2 (correct ranking)
func TestRepos_ListMinimalRepos_query2(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "a/def"},
		{Name: "b/def"},
		{Name: "c/def"},
		{Name: "def/ghi"},
		{Name: "def/jkl"},
		{Name: "def/mno"},
		{Name: "abc/m"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query string
		want  []api.RepoName
	}{
		{"def", []api.RepoName{"a/def", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"}},
		{"b/def", []api.RepoName{"b/def"}},
		{"def/", []api.RepoName{"def/ghi", "def/jkl", "def/mno"}},
		{"def/m", []api.RepoName{"def/mno"}},
	}
	for _, test := range tests {
		repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{Query: test.query})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(reposFromRepoNames(repos)); !reflect.DeepEqual(got, test.want) {
			t.Errorf("Unexpected repo result for query %q:\ngot:  %q\nwant: %q", test.query, got, test.want)
		}
	}
}

// Test sort
func TestRepos_ListMinimalRepos_sort(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "c/def"},
		{Name: "def/mno"},
		{Name: "b/def"},
		{Name: "abc/m"},
		{Name: "abc/def"},
		{Name: "def/jkl"},
		{Name: "def/ghi"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query   string
		orderBy RepoListOrderBy
		want    []api.RepoName
	}{
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListName,
			}},
			want: []api.RepoName{"abc/def", "abc/m", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListCreatedAt,
			}},
			want: []api.RepoName{"c/def", "def/mno", "b/def", "abc/m", "abc/def", "def/jkl", "def/ghi"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCreatedAt,
				Descending: true,
			}},
			want: []api.RepoName{"def/ghi", "def/jkl", "abc/def", "abc/m", "b/def", "def/mno", "c/def"},
		},
		{
			query: "def",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCreatedAt,
				Descending: true,
			}},
			want: []api.RepoName{"def/ghi", "def/jkl", "abc/def", "b/def", "def/mno", "c/def"},
		},
	}
	for _, test := range tests {
		repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{Query: test.query, OrderBy: test.orderBy})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(reposFromRepoNames(repos)); !reflect.DeepEqual(got, test.want) {
			t.Errorf("Unexpected repo result for query %q, orderBy %v:\ngot:  %q\nwant: %q", test.query, test.orderBy, got, test.want)
		}
	}
}

// TestRepos_ListMinimalRepos_patterns tests the behavior of Repos.List when called with
// IncludePatterns and ExcludePattern.
func TestRepos_ListMinimalRepos_patterns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	createdRepos := []*types.Repo{
		{Name: "a/b"},
		{Name: "c/d"},
		{Name: "e/f"},
		{Name: "g/h"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, db, repo)
	}
	tests := []struct {
		includePatterns []string
		excludePattern  string
		want            []api.RepoName
	}{
		{
			includePatterns: []string{"(a|c)"},
			want:            []api.RepoName{"a/b", "c/d"},
		},
		{
			includePatterns: []string{"(a|c)", "b"},
			want:            []api.RepoName{"a/b"},
		},
		{
			includePatterns: []string{"(a|c)"},
			excludePattern:  "d",
			want:            []api.RepoName{"a/b"},
		},
		{
			excludePattern: "(d|e)",
			want:           []api.RepoName{"a/b", "g/h"},
		},
	}
	for _, test := range tests {
		repos, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{
			IncludePatterns: test.includePatterns,
			ExcludePattern:  test.excludePattern,
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(reposFromRepoNames(repos)); !reflect.DeepEqual(got, test.want) {
			t.Errorf("include %q exclude %q: got repos %q, want %q", test.includePatterns, test.excludePattern, got, test.want)
		}
	}
}

func TestRepos_ListMinimalRepos_queryAndPatternsMutuallyExclusive(t *testing.T) {
	ctx := actor.WithInternalActor(context.Background())
	wantErr := "Query and IncludePatterns/ExcludePattern options are mutually exclusive"

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	t.Run("Query and IncludePatterns", func(t *testing.T) {
		_, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{Query: "x", IncludePatterns: []string{"y"}})
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("got error %v, want it to contain %q", err, wantErr)
		}
	})

	t.Run("Query and ExcludePattern", func(t *testing.T) {
		_, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{Query: "x", ExcludePattern: "y"})
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("got error %v, want it to contain %q", err, wantErr)
		}
	})
}

func TestRepos_ListMinimalRepos_UserIDAndExternalServiceIDsMutuallyExclusive(t *testing.T) {
	ctx := actor.WithInternalActor(context.Background())
	wantErr := "options ExternalServiceIDs, UserID and OrgID are mutually exclusive"

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	_, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{UserID: 1, ExternalServiceIDs: []int64{2}})
	if err == nil || !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("got error %v, want it to contain %q", err, wantErr)
	}
}

func TestRepos_ListMinimalRepos_useOr(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	archived := types.Repos{typestest.MakeGitlabRepo()}.With(func(r *types.Repo) { r.Archived = true })
	archived = mustCreate(ctx, t, db, archived[0])
	forks := types.Repos{typestest.MakeGitoliteRepo()}.With(func(r *types.Repo) { r.Fork = true })
	forks = mustCreate(ctx, t, db, forks[0])
	cloned := types.Repos{typestest.MakeGithubRepo()}
	cloned = mustCreateGitserverRepo(ctx, t, db, cloned[0], types.GitserverRepo{CloneStatus: types.CloneStatusCloned})

	archivedAndForks := append(archived, forks...)
	sort.Sort(archivedAndForks)
	all := append(archivedAndForks, cloned...)
	sort.Sort(all)

	tests := []struct {
		name string
		opt  ReposListOptions
		want []types.MinimalRepo
	}{
		{"Archived or Forks", ReposListOptions{OnlyArchived: true, OnlyForks: true, UseOr: true}, repoNamesFromRepos(archivedAndForks)},
		{"Archived or Forks Or Cloned", ReposListOptions{OnlyArchived: true, OnlyForks: true, OnlyCloned: true, UseOr: true}, repoNamesFromRepos(all)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos, err := db.Repos().ListMinimalRepos(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, test.want, repos)
		})
	}
}

func TestRepos_ListMinimalRepos_externalServiceID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	services := typestest.MakeExternalServices()
	service1 := services[0]
	service2 := services[1]
	if err := db.ExternalServices().Create(ctx, confGet, service1); err != nil {
		t.Fatal(err)
	}
	if err := db.ExternalServices().Create(ctx, confGet, service2); err != nil {
		t.Fatal(err)
	}

	mine := types.Repos{typestest.MakeGithubRepo(service1)}
	if err := db.Repos().Create(ctx, mine...); err != nil {
		t.Fatal(err)
	}

	yours := types.Repos{typestest.MakeGitlabRepo(service2)}
	if err := db.Repos().Create(ctx, yours...); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		opt  ReposListOptions
		want []types.MinimalRepo
	}{
		{"Some", ReposListOptions{ExternalServiceIDs: []int64{service1.ID}}, repoNamesFromRepos(mine)},
		{"Default", ReposListOptions{}, repoNamesFromRepos(append(mine, yours...))},
		{"NonExistant", ReposListOptions{ExternalServiceIDs: []int64{1000}}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos, err := db.Repos().ListMinimalRepos(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, test.want, repos)
		})
	}
}

// This function tests for both individual uses of ExternalRepoIncludeContains,
// ExternalRepoExcludeContains as well as combination of these two options.
func TestRepos_ListMinimalRepos_externalRepoContains(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	svc := &types.ExternalService{
		Kind:        extsvc.KindPerforce,
		DisplayName: "Perforce - Test",
		Config:      `{"p4.port": "ssl:111.222.333.444:1666", "p4.user": "admin", "p4.passwd": "pa$$word", "depots": [], "repositoryPathPattern": "perforce/{depot}"}`,
	}
	if err := db.ExternalServices().Create(ctx, confGet, svc); err != nil {
		t.Fatal(err)
	}

	var (
		perforceMarketing = &types.Repo{
			Name:    api.RepoName("perforce/Marketing"),
			URI:     "Marketing",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "//Marketing/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineering = &types.Repo{
			Name:    api.RepoName("perforce/Engineering"),
			URI:     "Engineering",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "//Engineering/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineeringFrontend = &types.Repo{
			Name:    api.RepoName("perforce/Engineering/Frontend"),
			URI:     "Engineering/Frontend",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "//Engineering/Frontend/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineeringBackend = &types.Repo{
			Name:    api.RepoName("perforce/Engineering/Backend"),
			URI:     "Engineering/Backend",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "//Engineering/Backend/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineeringHandbookFrontend = &types.Repo{
			Name:    api.RepoName("perforce/Engineering/Handbook/Frontend"),
			URI:     "Engineering/Handbook/Frontend",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "//Engineering/Handbook/Frontend/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineeringHandbookBackend = &types.Repo{
			Name:    api.RepoName("perforce/Engineering/Handbook/Backend"),
			URI:     "Engineering/Handbook/Backend",
			Private: true,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "//Engineering/Handbook/Backend/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
	)
	if err := db.Repos().Create(ctx,
		perforceMarketing,
		perforceEngineering,
		perforceEngineeringFrontend,
		perforceEngineeringBackend,
		perforceEngineeringHandbookFrontend,
		perforceEngineeringHandbookBackend); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		opt  ReposListOptions
		want []types.MinimalRepo
	}{
		{
			name: "only apply ExternalRepoIncludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			want: repoNamesFromRepos([]*types.Repo{perforceEngineering, perforceEngineeringFrontend, perforceEngineeringBackend, perforceEngineeringHandbookFrontend, perforceEngineeringHandbookBackend}),
		},
		{
			name: "only apply transformed '...' Perforce wildcard ExternalRepoIncludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//%/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			want: repoNamesFromRepos([]*types.Repo{perforceEngineeringBackend, perforceEngineeringHandbookBackend}),
		},
		{
			name: "only apply multiple transformed '...' Perforce wildcard ExternalRepoIncludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//%/%/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			// Only match this specific nested folder, and not the other Backends
			want: repoNamesFromRepos([]*types.Repo{perforceEngineeringHandbookBackend}),
		},
		{
			name: "only apply transformed '*' Perforce wildcard ExternalRepoIncludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//[^/]+/[^/]+/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			// Only match this specific nested folder, and not the other Backends
			want: repoNamesFromRepos([]*types.Repo{perforceEngineeringHandbookBackend}),
		},
		{
			name: "only apply transformed '*' Perforce partial wildcard ExternalRepoIncludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//[^/]+/Back[^/]+/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			// Only match this specific nested folder, and not the other Backends
			want: repoNamesFromRepos([]*types.Repo{perforceEngineeringBackend}),
		},
		{
			name: "only apply ExternalRepoExcludeContains",
			opt: ReposListOptions{
				ExternalRepoExcludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			want: repoNamesFromRepos([]*types.Repo{perforceMarketing}),
		},
		{
			name: "only apply transformed '...' Perforce wildcard ExternalRepoExcludeContains",
			opt: ReposListOptions{
				ExternalRepoExcludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//%/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			want: repoNamesFromRepos([]*types.Repo{perforceMarketing, perforceEngineering, perforceEngineeringFrontend, perforceEngineeringHandbookFrontend}),
		},
		{
			name: "only apply transformed '*' Perforce wildcard ExternalRepoExcludeContains",
			opt: ReposListOptions{
				ExternalRepoExcludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//[^/]+/[^/]+/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			// Only filter this very specific nesting level
			want: repoNamesFromRepos([]*types.Repo{perforceMarketing, perforceEngineering, perforceEngineeringFrontend, perforceEngineeringBackend, perforceEngineeringHandbookFrontend}),
		},
		{
			name: "apply both ExternalRepoIncludeContains and ExternalRepoExcludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternalRepoExcludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
					{
						ID:          "//Engineering/Handbook/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			want: repoNamesFromRepos([]*types.Repo{perforceEngineering, perforceEngineeringFrontend}),
		},
		{
			name: "apply both ExternalRepoIncludeContains and transformed '...' Perforce wildcard ExternalRepoExcludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternalRepoExcludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//%/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			want: repoNamesFromRepos([]*types.Repo{perforceEngineering, perforceEngineeringFrontend, perforceEngineeringHandbookFrontend}),
		},
		{
			name: "apply both ExternalRepoIncludeContains and transformed '*' Perforce wildcard ExternalRepoExcludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternalRepoExcludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/[^/]+/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			want: repoNamesFromRepos([]*types.Repo{perforceEngineering, perforceEngineeringFrontend, perforceEngineeringBackend, perforceEngineeringHandbookFrontend}),
		},
		{
			name: "apply both transformed '...' Perforce wildcard ExternalRepoIncludeContains and ExternalRepoExcludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//%/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternalRepoExcludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/Handbook/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			want: repoNamesFromRepos([]*types.Repo{perforceEngineeringBackend}),
		},
		{
			name: "apply both transformed '*' Perforce wildcard ExternalRepoIncludeContains and ExternalRepoExcludeContains",
			opt: ReposListOptions{
				ExternalRepoIncludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/[^/]+/Backend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternalRepoExcludeContains: []api.ExternalRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			want: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos, err := db.Repos().ListMinimalRepos(ctx, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, test.want, repos)
		})
	}
}

func TestRepos_createRepo_dupe(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	// Add a repo.
	createRepo(ctx, t, db, &types.Repo{Name: "a/b"})

	// Add another repo with the same name.
	createRepo(ctx, t, db, &types.Repo{Name: "a/b"})
}

func TestRepos_ListRepos_UserPublicRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	user, repo := initUserAndRepo(t, ctx, db)
	// create a repo we don't own
	_, otherRepo := initUserAndRepo(t, ctx, db)

	// register our interest in the other user's repo
	err := db.UserPublicRepos().SetUserRepo(ctx, UserPublicRepo{UserID: user.ID, RepoURI: otherRepo.URI, RepoID: otherRepo.ID})
	if err != nil {
		t.Fatal(err)
	}

	want := []types.MinimalRepo{
		{ID: repo.ID, Name: repo.Name},
	}

	have, err := db.Repos().ListMinimalRepos(ctx, ReposListOptions{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf(diff)
	}

	want = []types.MinimalRepo{
		{ID: repo.ID, Name: repo.Name},
		{ID: otherRepo.ID, Name: otherRepo.Name},
	}

	have, err = db.Repos().ListMinimalRepos(ctx, ReposListOptions{UserID: user.ID, IncludeUserPublicRepos: true})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf(diff)
	}
}

func TestGetFirstRepoNamesByCloneURL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := actor.WithInternalActor(context.Background())

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	services := typestest.MakeExternalServices()
	service1 := services[0]
	if err := db.ExternalServices().Create(ctx, confGet, service1); err != nil {
		t.Fatal(err)
	}

	repo1 := typestest.MakeGithubRepo(service1)
	if err := db.Repos().Create(ctx, repo1); err != nil {
		t.Fatal(err)
	}

	_, err := db.ExecContext(ctx, "UPDATE external_service_repos SET clone_url = 'https://github.com/foo/bar' WHERE repo_id = $1", repo1.ID)
	if err != nil {
		t.Fatal(err)
	}

	name, err := db.Repos().GetFirstRepoNameByCloneURL(ctx, "https://github.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	if name != "github.com/foo/bar" {
		t.Fatalf("Want %q, got %q", "github.com/foo/bar", name)
	}
}

func initUserAndRepo(t *testing.T, ctx context.Context, db DB) (*types.User, *types.Repo) {
	id := rand.String(8)
	user, err := db.Users().Create(ctx, NewUser{
		Email:                 id + "@example.com",
		Username:              "u" + id,
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: user.ID,
	})

	now := time.Now()

	// Create an external service
	service := types.ExternalService{
		Kind:            extsvc.KindGitHub,
		DisplayName:     "Github - Test",
		Config:          `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
		CreatedAt:       now,
		UpdatedAt:       now,
		NamespaceUserID: user.ID,
	}
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	err = db.ExternalServices().Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}

	repo := &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "r" + id,
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        api.RepoName("github.com/sourcegraph/" + rand.String(10)),
		Private:     false,
		URI:         "uri",
		Description: "description",
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			service.URN(): {
				ID:       service.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	}
	err = db.Repos().Create(ctx, repo)
	if err != nil {
		t.Fatal(err)
	}
	return user, repo
}

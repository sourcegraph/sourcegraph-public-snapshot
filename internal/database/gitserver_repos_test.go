package database

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
)

const shardID = "test"

func TestIterateRepoGitserverStatus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	repos := types.Repos{
		&types.Repo{Name: "github.com/sourcegraph/repo1"},
		&types.Repo{Name: "github.com/sourcegraph/repo2"},
		&types.Repo{Name: "github.com/sourcegraph/repo3"},
	}
	createTestRepos(ctx, t, db, repos)

	// Soft delete one of the repos
	if err := db.Repos().Delete(ctx, repos[2].ID); err != nil {
		t.Fatal(err)
	}

	if err := db.GitserverRepos().Update(ctx, &types.GitserverRepo{
		RepoID:      repos[0].ID,
		ShardID:     "shard-0",
		CloneStatus: types.CloneStatusCloned,
	}); err != nil {
		t.Fatal(err)
	}

	assert := func(t *testing.T, wantRepoCount, wantStatusCount int, options IterateRepoGitserverStatusOptions) {
		var statusCount int
		var seen []api.RepoName
		err := db.GitserverRepos().IterateRepoGitserverStatus(ctx, options, func(repo types.RepoGitserverStatus) error {
			seen = append(seen, repo.Name)
			statusCount++
			if repo.GitserverRepo.RepoID == 0 {
				t.Fatal("GitServerRepo has zero id")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("Seen: %v", seen)
		if len(seen) != wantRepoCount {
			t.Fatalf("Expected %d repos, got %d", wantRepoCount, len(seen))
		}

		if statusCount != wantStatusCount {
			t.Fatalf("Expected %d statuses, got %d", wantStatusCount, statusCount)
		}
	}

	t.Run("iterate with default options", func(t *testing.T) {
		assert(t, 2, 2, IterateRepoGitserverStatusOptions{})
	})
	t.Run("iterate only repos without shard", func(t *testing.T) {
		assert(t, 1, 1, IterateRepoGitserverStatusOptions{OnlyWithoutShard: true})
	})
	t.Run("include deleted", func(t *testing.T) {
		assert(t, 3, 3, IterateRepoGitserverStatusOptions{IncludeDeleted: true})
	})
	t.Run("include deleted, but still only without shard", func(t *testing.T) {
		assert(t, 2, 2, IterateRepoGitserverStatusOptions{OnlyWithoutShard: true, IncludeDeleted: true})
	})
}

func TestIteratePurgeableRepos(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := basestore.NewWithHandle(db.Handle())

	normalRepo := &types.Repo{Name: "normal"}
	blockedRepo := &types.Repo{Name: "blocked"}
	deletedRepo := &types.Repo{Name: "deleted"}
	notCloned := &types.Repo{Name: "notCloned"}

	createTestRepos(ctx, t, db, types.Repos{
		normalRepo,
		blockedRepo,
		deletedRepo,
		notCloned,
	})
	for _, repo := range []*types.Repo{normalRepo, blockedRepo, deletedRepo} {
		updateTestGitserverRepos(ctx, t, db, false, types.CloneStatusCloned, repo.ID)
	}
	// Delete & load soft-deleted name of repo
	if err := db.Repos().Delete(ctx, deletedRepo.ID); err != nil {
		t.Fatal(err)
	}
	deletedRepoNameStr, _, err := basestore.ScanFirstString(store.Query(ctx, sqlf.Sprintf("SELECT name FROM repo WHERE id = %s", deletedRepo.ID)))
	if err != nil {
		t.Fatal(err)
	}
	deletedRepoName := api.RepoName(deletedRepoNameStr)

	// Blocking a repo is currently done manually
	if _, err := db.ExecContext(ctx, `UPDATE repo set blocked = '{}' WHERE id = $1`, blockedRepo.ID); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name      string
		options   IteratePurgableReposOptions
		wantRepos []api.RepoName
	}{
		{
			name: "zero deletedBefore",
			options: IteratePurgableReposOptions{
				DeletedBefore: time.Time{},
				Limit:         0,
			},
			wantRepos: []api.RepoName{deletedRepoName, blockedRepo.Name},
		},
		{
			name: "deletedBefore now",
			options: IteratePurgableReposOptions{
				DeletedBefore: time.Now(),
				Limit:         0,
			},

			wantRepos: []api.RepoName{deletedRepoName, blockedRepo.Name},
		},
		{
			name: "deletedBefore 5 minutes ago",
			options: IteratePurgableReposOptions{
				DeletedBefore: time.Now().Add(-5 * time.Minute),
				Limit:         0,
			},
			wantRepos: []api.RepoName{blockedRepo.Name},
		},
		{
			name: "test limit",
			options: IteratePurgableReposOptions{
				DeletedBefore: time.Time{},
				Limit:         1,
			},
			wantRepos: []api.RepoName{deletedRepoName},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var have []api.RepoName
			if err := db.GitserverRepos().IteratePurgeableRepos(ctx, tt.options, func(repo api.RepoName) error {
				have = append(have, repo)
				return nil
			}); err != nil {
				t.Fatal(err)
			}

			sort.Slice(have, func(i, j int) bool { return have[i] < have[j] })

			if diff := cmp.Diff(have, tt.wantRepos); diff != "" {
				t.Fatalf("wrong iterated: %s", diff)
			}
		})
	}
}

func TestIterateWithNonemptyLastError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	type testRepo struct {
		name         string
		cloudDefault bool
		hasLastError bool
	}
	type testCase struct {
		name               string
		testRepos          []testRepo
		expectedReposFound []api.RepoName
	}
	testCases := []testCase{
		{
			name: "get repos with last error",
			testRepos: []testRepo{
				{
					name:         "github.com/sourcegraph/repo1",
					cloudDefault: true,
					hasLastError: true,
				},
				{
					name:         "github.com/sourcegraph/repo2",
					cloudDefault: true,
				},
			},
			expectedReposFound: []api.RepoName{"github.com/sourcegraph/repo1"},
		},
		{
			name: "filter out non cloud_default repos",
			testRepos: []testRepo{
				{
					name:         "github.com/sourcegraph/repo1",
					cloudDefault: false,
					hasLastError: true,
				},
				{
					name:         "github.com/sourcegraph/repo2",
					cloudDefault: true,
					hasLastError: true,
				},
			},
			expectedReposFound: []api.RepoName{"github.com/sourcegraph/repo2"},
		},
		{
			name: "no cloud_default repos with non-empty last errors",
			testRepos: []testRepo{
				{
					name:         "github.com/sourcegraph/repo1",
					cloudDefault: false,
					hasLastError: true,
				},
				{
					name:         "github.com/sourcegraph/repo2",
					cloudDefault: true,
					hasLastError: false,
				},
			},
			expectedReposFound: []api.RepoName{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			logger := logtest.Scoped(t)
			db := NewDB(logger, dbtest.NewDB(logger, t))
			now := time.Now()

			cloudDefaultService := createTestExternalService(ctx, t, now, db, true)
			nonCloudDefaultService := createTestExternalService(ctx, t, now, db, false)
			for i, tr := range tc.testRepos {
				testRepo := &types.Repo{
					Name:        api.RepoName(tr.name),
					URI:         tr.name,
					Description: "",
					ExternalRepo: api.ExternalRepoSpec{
						ID:          fmt.Sprintf("repo%d-external", i),
						ServiceType: extsvc.TypeGitHub,
						ServiceID:   "https://github.com",
					},
				}
				if tr.cloudDefault {
					testRepo = testRepo.With(
						typestest.Opt.RepoSources(cloudDefaultService.URN()),
					)
				} else {
					testRepo = testRepo.With(
						typestest.Opt.RepoSources(nonCloudDefaultService.URN()),
					)
				}
				createTestRepos(ctx, t, db, types.Repos{testRepo})

				if tr.hasLastError {
					if err := db.GitserverRepos().SetLastError(ctx, testRepo.Name, "an error", "test"); err != nil {
						t.Fatal(err)
					}
				}
			}

			foundRepos := make([]api.RepoName, 0, len(tc.testRepos))

			// Iterate and collect repos
			err := db.GitserverRepos().IterateWithNonemptyLastError(ctx, func(repo api.RepoName) error {
				foundRepos = append(foundRepos, repo)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(foundRepos) != len(tc.expectedReposFound) {
				t.Fatalf("expected %d repos with non empty last error, got %d", len(tc.expectedReposFound),
					len(foundRepos))
			}
			for i, fr := range foundRepos {
				if !fr.Equal(tc.expectedReposFound[i]) {
					t.Fatalf("expected repo %s got %s instead", fr, tc.expectedReposFound[i])
				}
			}

			total, err := db.GitserverRepos().TotalErroredCloudDefaultRepos(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if total != len(tc.expectedReposFound) {
				t.Fatalf("expected %d total errored repos, got %d instead", len(tc.expectedReposFound), total)
			}
		})
	}
}

func createTestExternalService(ctx context.Context, t *testing.T, now time.Time, db DB, cloudDefault bool) types.ExternalService {
	service := types.ExternalService{
		Kind:         extsvc.KindGitHub,
		DisplayName:  "Github - Test",
		Config:       `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
		CreatedAt:    now,
		UpdatedAt:    now,
		CloudDefault: cloudDefault,
	}

	// Create a new external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternalServices().Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestGitserverReposGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create one test repo
	_, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		RepoSizeBytes: 100,
	})

	// GetByID should now work
	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}
}

func TestGitserverReposGetByName(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		RepoSizeBytes: 100,
	})

	// GetByName should now work
	fromDB, err := db.GitserverRepos().GetByName(ctx, repo.Name)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}
}

func TestGitserverReposGetByNames(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	gitserverRepoStore := &gitserverRepoStore{Store: basestore.NewWithHandle(db.Handle())}

	// Creating a few repos
	repoNames := make([]api.RepoName, 5)
	gitserverRepos := make([]*types.GitserverRepo, 5)
	for i := 0; i < len(repoNames); i++ {
		repoName := fmt.Sprintf("github.com/sourcegraph/repo%d", i)
		repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
			Name: api.RepoName(repoName),
		})
		repoNames[i] = repo.Name
		gitserverRepos[i] = gitserverRepo
	}

	for i := 0; i < len(repoNames); i++ {
		have, err := gitserverRepoStore.GetByNames(ctx, repoNames[:i+1]...)
		if err != nil {
			t.Fatal(err)
		}
		haveRepos := make([]*types.GitserverRepo, 0, len(have))
		for _, r := range have {
			haveRepos = append(haveRepos, r)
		}
		sort.Slice(haveRepos, func(i, j int) bool {
			return haveRepos[i].RepoID < haveRepos[j].RepoID
		})
		if diff := cmp.Diff(gitserverRepos[:i+1], haveRepos, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
			t.Fatal(diff)
		}
	}
}

func TestSetCloneStatus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		RepoSizeBytes: 100,
		CloneStatus:   types.CloneStatusNotCloned,
	})

	// Set cloned
	err := db.GitserverRepos().SetCloneStatus(ctx, repo.Name, types.CloneStatusCloned, "")
	if err != nil {
		t.Fatal(err)
	}

	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.CloneStatus = types.CloneStatusCloned
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Setting clone status should work even if no row exists in gitserver table
	repo2 := &types.Repo{
		Name:         "github.com/sourcegraph/repo2",
		URI:          "github.com/sourcegraph/repo2",
		ExternalRepo: api.ExternalRepoSpec{},
	}

	// Create one test repo
	err = db.Repos().Create(ctx, repo2)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.GitserverRepos().SetCloneStatus(ctx, repo2.Name, types.CloneStatusCloned, shardID); err != nil {
		t.Fatal(err)
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, repo2.ID)
	if err != nil {
		t.Fatal(err)
	}
	gitserverRepo2 := &types.GitserverRepo{
		RepoID:      repo2.ID,
		ShardID:     shardID,
		CloneStatus: types.CloneStatusCloned,
	}
	if diff := cmp.Diff(gitserverRepo2, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt", "LastFetched", "LastChanged")); diff != "" {
		t.Fatal(diff)
	}

	// Setting the same status again should not touch the row
	if err := db.GitserverRepos().SetCloneStatus(ctx, repo2.Name, types.CloneStatusCloned, shardID); err != nil {
		t.Fatal(err)
	}
	after, err := db.GitserverRepos().GetByID(ctx, repo2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(fromDB, after); diff != "" {
		t.Fatal(diff)
	}
}

func TestSetLastError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	})

	// Set error.
	//
	// We are using a null terminated string for the last_error column. See
	// https://stackoverflow.com/a/38008565/1773961 on how to set null terminated strings in Go.
	err := db.GitserverRepos().SetLastError(ctx, repo.Name, "oops\x00", "")
	if err != nil {
		t.Fatal(err)
	}

	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.LastError = "oops"
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Remove error
	const emptyErr = ""
	err = db.GitserverRepos().SetLastError(ctx, repo.Name, emptyErr, "")
	if err != nil {
		t.Fatal(err)
	}

	fromDB, err = db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.LastError = emptyErr
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Set again to same value, updated_at should not change
	err = db.GitserverRepos().SetLastError(ctx, repo.Name, emptyErr, shardID)
	if err != nil {
		t.Fatal(err)
	}

	after, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.LastError = emptyErr
	if diff := cmp.Diff(fromDB, after); diff != "" {
		t.Fatal(diff)
	}

	// Setting to empty error should set the column to null
	count, _, err := basestore.ScanFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM gitserver_repos WHERE last_error IS NULL"))
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("Want %d, got %d", 1, count)
	}
}

func TestSetRepoSize(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		RepoSizeBytes: 100,
	})

	// Set repo size
	err := db.GitserverRepos().SetRepoSize(ctx, repo.Name, 200, "")
	if err != nil {
		t.Fatal(err)
	}

	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.RepoSizeBytes = 200
	// If we have size, we can assume it's cloned
	gitserverRepo.CloneStatus = types.CloneStatusCloned
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Setting repo size should work even if no row exists
	repo2 := &types.Repo{
		Name:         "github.com/sourcegraph/repo2",
		URI:          "github.com/sourcegraph/repo2",
		ExternalRepo: api.ExternalRepoSpec{},
	}

	// Create one test repo
	err = db.Repos().Create(ctx, repo2)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.GitserverRepos().SetRepoSize(ctx, repo2.Name, 300, ""); err != nil {
		t.Fatal(err)
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, repo2.ID)
	if err != nil {
		t.Fatal(err)
	}
	gitserverRepo2 := &types.GitserverRepo{
		RepoID:        repo2.ID,
		ShardID:       "",
		RepoSizeBytes: 300,
		// If we have size, we can assume it's cloned
		CloneStatus: types.CloneStatusCloned,
	}
	if diff := cmp.Diff(gitserverRepo2, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt", "LastFetched", "LastChanged", "CloneStatus")); diff != "" {
		t.Fatal(diff)
	}

	// Setting the same size should not touch the row
	if err := db.GitserverRepos().SetRepoSize(ctx, repo2.Name, 300, ""); err != nil {
		t.Fatal(err)
	}
	after, err := db.GitserverRepos().GetByID(ctx, repo2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(fromDB, after); diff != "" {
		t.Fatal(diff)
	}
}

func TestGitserverRepo_Update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	})

	// Change clone status
	gitserverRepo.CloneStatus = types.CloneStatusCloned
	if err := db.GitserverRepos().Update(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
	fromDB, err := db.GitserverRepos().GetByID(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Change error

	// We want to test if update can handle the writing a null character to the last_error
	// column. See https://stackoverflow.com/a/38008565/1773961 on how to set null terminated
	// strings in Go.
	gitserverRepo.LastError = "Oops\x00"
	if err := db.GitserverRepos().Update(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Set LastError to the expected error string but without the null character, because we expect
	// our code to work and strip it before writing to the DB.
	gitserverRepo.LastError = "Oops"
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Remove error
	gitserverRepo.LastError = ""
	if err := db.GitserverRepos().Update(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}
}

func TestGitserverRepoUpdateMany(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create two test repos
	repo1, gitserverRepo1 := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo1",
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	})
	repo2, gitserverRepo2 := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo2",
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	})

	// Change their clone statuses
	gitserverRepo1.CloneStatus = types.CloneStatusCloned
	gitserverRepo2.CloneStatus = types.CloneStatusCloning
	if err := db.GitserverRepos().Update(ctx, gitserverRepo1, gitserverRepo2); err != nil {
		t.Fatal(err)
	}

	// Confirm
	t.Run("repo1", func(t *testing.T) {
		fromDB, err := db.GitserverRepos().GetByID(ctx, repo1.ID)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(gitserverRepo1, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
			t.Fatal(diff)
		}
	})
	t.Run("repo2", func(t *testing.T) {
		fromDB, err := db.GitserverRepos().GetByID(ctx, repo2.ID)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(gitserverRepo2, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestSanitizeToUTF8(t *testing.T) {
	testSet := map[string]string{
		"test\x00":     "test",
		"test\x00test": "testtest",
		"\x00test":     "test",
	}

	for input, expected := range testSet {
		got := sanitizeToUTF8(input)
		if got != expected {
			t.Fatalf("Failed to sanitize to UTF-8, got %q but wanted %q", got, expected)
		}
	}
}

func TestGitserverRepoListReposWithoutSize(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		RepoSizeBytes: 0,
	})

	if _, err := db.Handle().ExecContext(ctx, fmt.Sprintf(
		`update gitserver_repos set repo_size_bytes = null where repo_id = %d;`,
		gitserverRepo.RepoID)); err != nil {
		t.Fatalf("unexpected error while updating gitserver repo: %s", err)
	}

	// Get it back from the db
	fromDB, err := db.GitserverRepos().GetByID(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Check that this repo is returned from ListReposWithoutSize
	if reposWithoutSize, err := db.GitserverRepos().ListReposWithoutSize(ctx); err != nil {
		t.Fatal(err)
	} else if len(reposWithoutSize) != 1 {
		t.Fatal("One repo without size should be returned")
	}

	// Setting the size
	gitserverRepo.RepoSizeBytes = 4040
	if err := db.GitserverRepos().Update(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}

	// Check that this repo is not returned now from ListReposWithoutSize
	if reposWithoutSize, err := db.GitserverRepos().ListReposWithoutSize(ctx); err != nil {
		t.Fatal(err)
	} else if len(reposWithoutSize) != 0 {
		t.Fatal("There should be no repos without size")
	}

	// Check that nothing except UpdatedAt and RepoSizeBytes has been changed
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt", "RepoSizeBytes")); diff != "" {
		t.Fatal(diff)
	}
}

func TestGitserverUpdateRepoSizes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	repo1, gitserverRepo1 := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name: "github.com/sourcegraph/repo1",
	})

	repo2, gitserverRepo2 := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name: "github.com/sourcegraph/repo2",
	})

	repo3, gitserverRepo3 := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name: "github.com/sourcegraph/repo3",
	})

	// Setting repo sizes in DB
	sizes := map[api.RepoID]int64{
		repo1.ID: 100,
		repo2.ID: 500,
		repo3.ID: 800,
	}
	numUpdated, err := db.GitserverRepos().UpdateRepoSizes(ctx, shardID, sizes)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := numUpdated, len(sizes); have != want {
		t.Fatalf("wrong number of repos updated. have=%d, want=%d", have, want)
	}

	// Updating sizes in test data for further diff comparison
	gitserverRepo1.RepoSizeBytes = sizes[gitserverRepo1.RepoID]
	gitserverRepo2.RepoSizeBytes = sizes[gitserverRepo2.RepoID]
	gitserverRepo3.RepoSizeBytes = sizes[gitserverRepo3.RepoID]

	// Checking repo diffs, excluding UpdatedAt. This is to verify that nothing except repo_size_bytes
	// has changed
	after1, err := db.GitserverRepos().GetByID(ctx, repo1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(gitserverRepo1, after1, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	after2, err := db.GitserverRepos().GetByID(ctx, repo2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(gitserverRepo2, after2, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	after3, err := db.GitserverRepos().GetByID(ctx, repo3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(gitserverRepo3, after3, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// update again to make sure they're not updated again
	numUpdated, err = db.GitserverRepos().UpdateRepoSizes(ctx, shardID, sizes)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := numUpdated, 0; have != want {
		t.Fatalf("wrong number of repos updated. have=%d, want=%d", have, want)
	}

	// update subset
	sizes = map[api.RepoID]int64{
		repo3.ID: 900,
	}
	numUpdated, err = db.GitserverRepos().UpdateRepoSizes(ctx, shardID, sizes)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := numUpdated, 1; have != want {
		t.Fatalf("wrong number of repos updated. have=%d, want=%d", have, want)
	}

	// update with different batch sizes
	gitserverRepoStore := &gitserverRepoStore{Store: basestore.NewWithHandle(db.Handle())}
	for _, batchSize := range []int64{1, 2, 3, 6} {
		sizes = map[api.RepoID]int64{
			repo1.ID: 123 + batchSize,
			repo2.ID: 456 + batchSize,
			repo3.ID: 789 + batchSize,
		}

		numUpdated, err = gitserverRepoStore.updateRepoSizesWithBatchSize(ctx, shardID, sizes, int(batchSize))
		if err != nil {
			t.Fatal(err)
		}
		if have, want := numUpdated, 3; have != want {
			t.Fatalf("wrong number of repos updated. have=%d, want=%d", have, want)
		}
	}
}

func createTestRepo(ctx context.Context, t *testing.T, db DB, payload *createTestRepoPayload) (*types.Repo, *types.GitserverRepo) {
	t.Helper()

	repo := &types.Repo{Name: payload.Name}

	// Create Repo
	err := db.Repos().Create(ctx, repo)
	if err != nil {
		t.Fatal(err)
	}

	// Get the gitserver repo
	gitserverRepo, err := db.GitserverRepos().GetByID(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}

	want := &types.GitserverRepo{
		RepoID:      repo.ID,
		CloneStatus: types.CloneStatusNotCloned,
	}
	if diff := cmp.Diff(want, gitserverRepo, cmpopts.IgnoreFields(types.GitserverRepo{}, "LastFetched", "LastChanged", "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	return repo, gitserverRepo
}

type createTestRepoPayload struct {
	// Repo related properties

	// Name is the name for this repository (e.g., "github.com/user/repo"). It
	// is the same as URI, unless the user configures a non-default
	// repositoryPathPattern.
	//
	// Previously, this was called RepoURI.
	Name api.RepoName

	// Gitserver related properties

	// Size of the repository in bytes.
	RepoSizeBytes int64
	CloneStatus   types.CloneStatus
}

func createTestRepos(ctx context.Context, t *testing.T, db DB, repos types.Repos) {
	t.Helper()
	err := db.Repos().Create(ctx, repos...)
	if err != nil {
		t.Fatal(err)
	}
}

func updateTestGitserverRepos(ctx context.Context, t *testing.T, db DB, hasLastError bool, cloneStatus types.CloneStatus, repoID api.RepoID) {
	t.Helper()
	gitserverRepo := &types.GitserverRepo{
		RepoID:      repoID,
		ShardID:     fmt.Sprintf("gitserver%d", repoID),
		CloneStatus: cloneStatus,
	}
	if hasLastError {
		gitserverRepo.LastError = "an error occurred"
	}
	if err := db.GitserverRepos().Update(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
}

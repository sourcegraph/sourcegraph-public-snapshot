package database

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	repos := types.Repos{
		&types.Repo{
			Name:         "github.com/sourcegraph/repo1",
			URI:          "github.com/sourcegraph/repo1",
			Description:  "",
			ExternalRepo: api.ExternalRepoSpec{},
			Sources:      nil,
		},
		&types.Repo{
			Name:         "github.com/sourcegraph/repo2",
			URI:          "github.com/sourcegraph/repo2",
			Description:  "",
			ExternalRepo: api.ExternalRepoSpec{},
			Sources:      nil,
		},
		&types.Repo{
			Name:         "github.com/sourcegraph/repo3",
			URI:          "github.com/sourcegraph/repo3",
			Description:  "",
			ExternalRepo: api.ExternalRepoSpec{},
			Sources:      nil,
		},
	}
	createTestRepos(ctx, t, db, repos)

	gitserverRepo := &types.GitserverRepo{
		RepoID:        repos[0].ID,
		ShardID:       "gitserver1",
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	}

	// Create one GitServerRepo
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}

	// Soft delete one of the repos
	if err := db.Repos().Delete(ctx, repos[2].ID); err != nil {
		t.Fatal(err)
	}

	assert := func(t *testing.T, wantRepoCount, wantStatusCount int, options IterateRepoGitserverStatusOptions) {
		var statusCount int
		var seen []api.RepoName
		err := db.GitserverRepos().IterateRepoGitserverStatus(ctx, options, func(repo types.RepoGitserverStatus) error {
			seen = append(seen, repo.Name)
			if repo.GitserverRepo != nil {
				statusCount++
				if repo.GitserverRepo.RepoID == 0 {
					t.Fatal("GitServerRepo has zero id")
				}
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
	t.Run("include deleted and without shard", func(t *testing.T) {
		assert(t, 2, 2, IterateRepoGitserverStatusOptions{OnlyWithoutShard: true, IncludeDeleted: true})
	})
}

func TestIteratePurgeableRepos(t *testing.T) {
	ctx := context.Background()
	db := NewDB(dbtest.NewDB(t))

	normalRepo := &types.Repo{
		Name: "normal",
	}
	blockedRepo := &types.Repo{
		Name: "blocked",
	}
	deletedRepo := &types.Repo{
		Name: "deleted",
	}
	notCloned := &types.Repo{
		Name: "notCloned",
	}

	createTestRepos(ctx, t, db, types.Repos{
		normalRepo,
		blockedRepo,
		deletedRepo,
		notCloned,
	})
	for _, repo := range []*types.Repo{normalRepo, blockedRepo, deletedRepo} {
		createTestGitserverRepos(ctx, t, db, false, types.CloneStatusCloned, repo.ID)
	}
	for _, repo := range []*types.Repo{notCloned} {
		createTestGitserverRepos(ctx, t, db, false, types.CloneStatusNotCloned, repo.ID)
	}
	if err := db.Repos().Delete(ctx, deletedRepo.ID); err != nil {
		t.Fatal(err)
	}
	// Blocking a repo is currently done manually
	if _, err := db.ExecContext(ctx, `UPDATE repo set blocked = '{}' WHERE id = $1`, blockedRepo.ID); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name         string
		options      IteratePurgableReposOptions
		blockedCount int
		deletedCount int
	}{
		{
			name: "zero deletedBefore",
			options: IteratePurgableReposOptions{
				DeletedBefore: time.Time{},
				Limit:         0,
			},
			blockedCount: 1,
			deletedCount: 1,
		},
		{
			name: "deletedBefore now",
			options: IteratePurgableReposOptions{
				DeletedBefore: time.Now(),
				Limit:         0,
			},

			blockedCount: 1,
			deletedCount: 1,
		},
		{
			name: "deletedBefore 5 minutes ago",
			options: IteratePurgableReposOptions{
				DeletedBefore: time.Now().Add(-5 * time.Minute),
				Limit:         0,
			},
			blockedCount: 1,
			deletedCount: 0,
		},
		{
			name: "test limit",
			options: IteratePurgableReposOptions{
				DeletedBefore: time.Time{},
				Limit:         1,
			},
			blockedCount: 0,
			deletedCount: 1,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var have []api.RepoName
			var blockedCount int
			var deletedCount int
			if err := db.GitserverRepos().IteratePurgeableRepos(ctx, tt.options, func(repo api.RepoName) error {
				if repo == "blocked" {
					blockedCount++
				}
				if strings.HasPrefix(string(repo), "DELETED") {
					deletedCount++
				}
				have = append(have, repo)
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			t.Log(have)
			if blockedCount != tt.blockedCount {
				t.Fatalf("Want %d blocked repos, have %d", tt.blockedCount, blockedCount)
			}
			if deletedCount != tt.deletedCount {
				t.Fatalf("Want %d deleted repos, have %d", tt.deletedCount, deletedCount)
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
			db := NewDB(dbtest.NewDB(t))
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
				createTestGitserverRepos(ctx, t, db, tr.hasLastError, types.CloneStatusNotCloned, testRepo.ID)
			}

			foundRepos := make([]types.RepoGitserverStatus, 0, len(tc.testRepos))

			// Iterate and collect repos
			err := db.GitserverRepos().IterateWithNonemptyLastError(ctx, func(repo types.RepoGitserverStatus) error {
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
				if !fr.Name.Equal(tc.expectedReposFound[i]) {
					t.Fatalf("expected repo %s got %s instead", fr.Name, tc.expectedReposFound[i])
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

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Create one test repo
	_, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		URI:           "github.com/sourcegraph/repo",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       "test",
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

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		URI:           "github.com/sourcegraph/repo",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       "test",
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

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	type testCase struct {
		name        string
		reposNumber int
		batchSize   int
	}
	testCases := []testCase{
		{
			name:        "GetReposByNames: repos=10, batch=3",
			reposNumber: 10,
			batchSize:   3,
		},
		{
			name:        "GetReposByNames: repos=5, batch=30",
			reposNumber: 5,
			batchSize:   30,
		},
		{
			name:        "GetReposByNames: repos=10, batch=5",
			reposNumber: 10,
			batchSize:   5,
		},
		{
			name:        "GetReposByNames: repos=1, batch=3",
			reposNumber: 1,
			batchSize:   3,
		},
	}

	repoIdx := 0
	gitserverRepoStore := &gitserverRepoStore{Store: basestore.NewWithHandle(db.Handle())}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Creating tc.reposNumber repos
			repoNames := make([]api.RepoName, 0)
			idToExpectedRepo := make(map[api.RepoID]*types.GitserverRepo, 0)
			for i := 0; i < tc.reposNumber; i++ {
				repoName := fmt.Sprintf("github.com/sourcegraph/repo%d", repoIdx)
				repoIdx++
				repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
					Name:         api.RepoName(repoName),
					URI:          repoName,
					ExternalRepo: api.ExternalRepoSpec{},
					ShardID:      shardID,
				})
				repoNames = append(repoNames, repo.Name)
				idToExpectedRepo[gitserverRepo.RepoID] = gitserverRepo
			}

			// GetByNames should now work
			fromDB, err := gitserverRepoStore.getByNames(ctx, tc.batchSize, repoNames...)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < tc.reposNumber; i++ {
				if diff := cmp.Diff(idToExpectedRepo[fromDB[i].RepoID], fromDB[i], cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestSetCloneStatus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		URI:           "github.com/sourcegraph/repo",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       shardID,
		RepoSizeBytes: 100,
		CloneStatus:   types.CloneStatusNotCloned,
	})

	// Set cloned
	err := db.GitserverRepos().SetCloneStatus(ctx, repo.Name, types.CloneStatusCloned, shardID)
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

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		URI:           "github.com/sourcegraph/repo",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       shardID,
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	})

	// Set error.
	//
	// We are using a null terminated string for the last_error column. See
	// https://stackoverflow.com/a/38008565/1773961 on how to set null terminated strings in Go.
	err := db.GitserverRepos().SetLastError(ctx, repo.Name, "oops\x00", shardID)
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
	err = db.GitserverRepos().SetLastError(ctx, repo.Name, emptyErr, shardID)
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

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		URI:           "github.com/sourcegraph/repo",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       shardID,
		RepoSizeBytes: 100,
	})

	// Set repo size
	err := db.GitserverRepos().SetRepoSize(ctx, repo.Name, 200, shardID)
	if err != nil {
		t.Fatal(err)
	}

	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.RepoSizeBytes = 200
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

	if err := db.GitserverRepos().SetRepoSize(ctx, repo2.Name, 300, shardID); err != nil {
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
	}
	if diff := cmp.Diff(gitserverRepo2, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt", "LastFetched", "LastChanged", "CloneStatus")); diff != "" {
		t.Fatal(diff)
	}

	// Setting the same size should not touch the row
	if err := db.GitserverRepos().SetRepoSize(ctx, repo2.Name, 300, shardID); err != nil {
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

func TestGitserverRepoUpsertNullShard(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	repo1 := &types.Repo{
		Name:         "github.com/sourcegraph/repo1",
		URI:          "github.com/sourcegraph/repo1",
		Description:  "",
		ExternalRepo: api.ExternalRepoSpec{},
		Sources:      nil,
	}

	// Create one test repo
	err := db.Repos().Create(ctx, repo1)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo := &types.GitserverRepo{
		RepoID:        repo1.ID,
		ShardID:       "",
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	}

	// Create one GitServerRepo
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo); err == nil {
		t.Fatal("Expected an error")
	}
}

func TestGitserverRepoUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		URI:           "github.com/sourcegraph/repo",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       shardID,
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	})

	// Change clone status
	gitserverRepo.CloneStatus = types.CloneStatusCloned
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo); err != nil {
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

	// We want to test if Upsert can handle the writing a null character to the last_error
	// column. See https://stackoverflow.com/a/38008565/1773961 on how to set null terminated
	// strings in Go.
	gitserverRepo.LastError = "Oops\x00"
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo); err != nil {
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
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo); err != nil {
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

func TestGitserverRepoUpsertMany(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Create two test repos
	repo1, gitserverRepo1 := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo1",
		URI:           "github.com/sourcegraph/repo1",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       shardID,
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	})
	repo2, gitserverRepo2 := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo2",
		URI:           "github.com/sourcegraph/repo2",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       shardID,
		CloneStatus:   types.CloneStatusNotCloned,
		RepoSizeBytes: 100,
	})

	// Change their clone statuses
	gitserverRepo1.CloneStatus = types.CloneStatusCloned
	gitserverRepo2.CloneStatus = types.CloneStatusCloning
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo1, gitserverRepo2); err != nil {
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

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	// Create one test repo
	repo, gitserverRepo := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo",
		URI:           "github.com/sourcegraph/repo",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       shardID,
		RepoSizeBytes: 100,
	})

	// Create one GitServerRepo without repo_size_bytes
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
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
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo); err != nil {
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

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	repo1, gitserverRepo1 := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo1",
		URI:           "github.com/sourcegraph/repo1",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       shardID,
		RepoSizeBytes: 100,
	})

	repo2, gitserverRepo2 := createTestRepo(ctx, t, db, &createTestRepoPayload{
		Name:          "github.com/sourcegraph/repo2",
		URI:           "github.com/sourcegraph/repo2",
		ExternalRepo:  api.ExternalRepoSpec{},
		ShardID:       shardID,
		RepoSizeBytes: 100,
	})

	// Setting repo sizes in DB
	sizes := map[api.RepoID]int64{
		repo1.ID: 100,
		repo2.ID: 500,
	}
	if err := db.GitserverRepos().UpdateRepoSizes(ctx, shardID, sizes); err != nil {
		t.Fatal(err)
	}

	// Updating sizes in test data for further diff comparison
	gitserverRepo1.RepoSizeBytes = sizes[gitserverRepo1.RepoID]
	gitserverRepo2.RepoSizeBytes = sizes[gitserverRepo2.RepoID]

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
}

func createTestRepo(ctx context.Context, t *testing.T, db DB, payload *createTestRepoPayload) (*types.Repo, *types.GitserverRepo) {
	t.Helper()

	repo := &types.Repo{
		Name:         payload.Name,
		URI:          payload.URI,
		Description:  payload.Description,
		ExternalRepo: payload.ExternalRepo,
		Sources:      payload.Sources,
	}

	// Create Repo
	err := db.Repos().Create(ctx, repo)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo := &types.GitserverRepo{
		RepoID:  repo.ID,
		ShardID: shardID,
	}

	// Create GitServerRepo
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}

	// Get it back and check whether it is not corrupted
	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
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
	// URI is the full name for this repository (e.g.,
	// "github.com/user/repo"). See the documentation for the Name field.
	URI string
	// Description is a brief description of the repository.
	Description string
	// ExternalRepo identifies this repository by its ID on the external service where it resides (and the external
	// service itself).
	ExternalRepo api.ExternalRepoSpec
	// Sources identifies all the repo sources this Repo belongs to.
	// The key is a URN created by extsvc.URN
	Sources map[string]*types.SourceInfo

	// Gitserver related properties

	// Size of the repository in bytes.
	RepoSizeBytes int64
	// Usually represented by a gitserver hostname
	ShardID     string
	CloneStatus types.CloneStatus
}

func createTestRepos(ctx context.Context, t *testing.T, db DB, repos types.Repos) {
	t.Helper()
	err := db.Repos().Create(ctx, repos...)
	if err != nil {
		t.Fatal(err)
	}
}

func createTestGitserverRepos(ctx context.Context, t *testing.T, db DB, hasLastError bool, cloneStatus types.CloneStatus, repoID api.RepoID) {
	t.Helper()
	gitserverRepo := &types.GitserverRepo{
		RepoID:      repoID,
		ShardID:     fmt.Sprintf("gitserver%d", repoID),
		CloneStatus: cloneStatus,
	}
	if hasLastError {
		gitserverRepo.LastError = "an error occurred"
	}
	if err := db.GitserverRepos().Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
}

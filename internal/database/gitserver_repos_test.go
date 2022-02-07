package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
)

func TestIterateRepoGitserverStatus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtest.NewDB(t)
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
	}
	createTestRepos(ctx, t, db, repos)

	gitserverRepo := &types.GitserverRepo{
		RepoID:      repos[0].ID,
		ShardID:     "gitserver1",
		CloneStatus: types.CloneStatusNotCloned,
	}

	// Create one GitServerRepo
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}

	var repoCount int
	var statusCount int

	// Iterate
	err := GitserverRepos(db).IterateRepoGitserverStatus(ctx, IterateRepoGitserverStatusOptions{}, func(repo types.RepoGitserverStatus) error {
		repoCount++
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

	wantRepoCount := 2
	if repoCount != wantRepoCount {
		t.Fatalf("Expected %d repos, got %d", wantRepoCount, repoCount)
	}

	wantStatusCount := 1
	if statusCount != wantStatusCount {
		t.Fatalf("Expected %d statuses, got %d", wantStatusCount, statusCount)
	}

	var noShardCount int
	// Iterate again against repos with no shard
	err = GitserverRepos(db).IterateRepoGitserverStatus(ctx, IterateRepoGitserverStatusOptions{OnlyWithoutShard: true}, func(repo types.RepoGitserverStatus) error {
		noShardCount++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	wantNoShardCount := 1
	if noShardCount != wantNoShardCount {
		t.Fatalf("Want %d, got %d", wantNoShardCount, noShardCount)
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
			db := dbtest.NewDB(t)
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

				createTestGitserverRepos(ctx, t, db, tr.hasLastError, testRepo.ID)
			}

			foundRepos := make([]types.RepoGitserverStatus, 0, len(tc.testRepos))

			// Iterate and collect repos
			err := GitserverRepos(db).IterateWithNonemptyLastError(ctx, func(repo types.RepoGitserverStatus) error {
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
		})
	}
}

func createTestExternalService(ctx context.Context, t *testing.T, now time.Time, db *sql.DB, cloudDefault bool) types.ExternalService {
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

	err := ExternalServices(db).Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestGitserverReposGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtest.NewDB(t)
	ctx := context.Background()

	_, err := GitserverRepos(db).GetByID(ctx, 1)
	if err == nil {
		t.Fatal("Expected an error")
	}

	repo1 := &types.Repo{
		Name:         "github.com/sourcegraph/repo1",
		URI:          "github.com/sourcegraph/repo1",
		Description:  "",
		ExternalRepo: api.ExternalRepoSpec{},
		Sources:      nil,
	}

	// Create one test repo
	err = Repos(db).Create(ctx, repo1)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo := &types.GitserverRepo{
		RepoID:      repo1.ID,
		ShardID:     "test",
		CloneStatus: types.CloneStatusNotCloned,
	}

	// Create GitServerRepo
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}

	// GetByID should now work
	fromDB, err := GitserverRepos(db).GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}
}

func TestSetCloneStatus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtest.NewDB(t)
	ctx := context.Background()
	const shardID = "test"

	repo1 := &types.Repo{
		Name:         "github.com/sourcegraph/repo1",
		URI:          "github.com/sourcegraph/repo1",
		ExternalRepo: api.ExternalRepoSpec{},
	}

	// Create one test repo
	err := Repos(db).Create(ctx, repo1)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo := &types.GitserverRepo{
		RepoID:      repo1.ID,
		ShardID:     shardID,
		CloneStatus: types.CloneStatusNotCloned,
	}

	// Create GitServerRepo
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}

	// Get it back
	fromDB, err := GitserverRepos(db).GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Set cloned
	err = GitserverRepos(db).SetCloneStatus(ctx, repo1.Name, types.CloneStatusCloned, shardID)
	if err != nil {
		t.Fatal(err)
	}

	fromDB, err = GitserverRepos(db).GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.CloneStatus = types.CloneStatusCloned
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Setting clone status should work even if no row exists
	repo2 := &types.Repo{
		Name:         "github.com/sourcegraph/repo2",
		URI:          "github.com/sourcegraph/repo2",
		ExternalRepo: api.ExternalRepoSpec{},
	}

	// Create one test repo
	err = Repos(db).Create(ctx, repo2)
	if err != nil {
		t.Fatal(err)
	}

	if err := GitserverRepos(db).SetCloneStatus(ctx, repo2.Name, types.CloneStatusCloned, shardID); err != nil {
		t.Fatal(err)
	}
	fromDB, err = GitserverRepos(db).GetByID(ctx, repo2.ID)
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
	if err := GitserverRepos(db).SetCloneStatus(ctx, repo2.Name, types.CloneStatusCloned, shardID); err != nil {
		t.Fatal(err)
	}
	after, err := GitserverRepos(db).GetByID(ctx, repo2.ID)
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

	db := dbtest.NewDB(t)
	ctx := context.Background()
	const shardID = "test"

	repo1 := &types.Repo{
		Name:         "github.com/sourcegraph/repo1",
		URI:          "github.com/sourcegraph/repo1",
		ExternalRepo: api.ExternalRepoSpec{},
	}

	// Create one test repo
	err := Repos(db).Create(ctx, repo1)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo := &types.GitserverRepo{
		RepoID:      repo1.ID,
		ShardID:     shardID,
		CloneStatus: types.CloneStatusNotCloned,
	}

	// Create GitServerRepo
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}

	// Set error.
	//
	// We are using a null terminated string for the last_error column. See
	// https://stackoverflow.com/a/38008565/1773961 on how to set null terminated strings in Go.
	err = GitserverRepos(db).SetLastError(ctx, repo1.Name, "oops\x00", shardID)
	if err != nil {
		t.Fatal(err)
	}

	fromDB, err := GitserverRepos(db).GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.LastError = "oops"
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Remove error
	err = GitserverRepos(db).SetLastError(ctx, repo1.Name, "", shardID)
	if err != nil {
		t.Fatal(err)
	}

	fromDB, err = GitserverRepos(db).GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.LastError = ""
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Set again to same value, updated_at should not change
	err = GitserverRepos(db).SetLastError(ctx, repo1.Name, "", shardID)
	if err != nil {
		t.Fatal(err)
	}

	after, err := GitserverRepos(db).GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo.LastError = ""
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

func TestGitserverRepoUpsertNullShard(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtest.NewDB(t)
	ctx := context.Background()

	repo1 := &types.Repo{
		Name:         "github.com/sourcegraph/repo1",
		URI:          "github.com/sourcegraph/repo1",
		Description:  "",
		ExternalRepo: api.ExternalRepoSpec{},
		Sources:      nil,
	}

	// Create one test repo
	err := Repos(db).Create(ctx, repo1)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo := &types.GitserverRepo{
		RepoID:      repo1.ID,
		ShardID:     "",
		CloneStatus: types.CloneStatusNotCloned,
	}

	// Create one GitServerRepo
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err == nil {
		t.Fatal("Expected an error")
	}
}

func TestGitserverRepoUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtest.NewDB(t)
	ctx := context.Background()

	repo1 := &types.Repo{
		Name:         "github.com/sourcegraph/repo1",
		URI:          "github.com/sourcegraph/repo1",
		Description:  "",
		ExternalRepo: api.ExternalRepoSpec{},
		Sources:      nil,
	}

	// Create two test repos
	err := Repos(db).Create(ctx, repo1)
	if err != nil {
		t.Fatal(err)
	}

	gitserverRepo := &types.GitserverRepo{
		RepoID:      repo1.ID,
		ShardID:     "abc",
		CloneStatus: types.CloneStatusNotCloned,
	}

	// Create one GitServerRepo
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}

	// Get it back from the db
	fromDB, err := GitserverRepos(db).GetByID(ctx, repo1.ID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}

	// Change clone status
	gitserverRepo.CloneStatus = types.CloneStatusCloned
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
	fromDB, err = GitserverRepos(db).GetByID(ctx, repo1.ID)
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
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
	fromDB, err = GitserverRepos(db).GetByID(ctx, repo1.ID)
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
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
	fromDB, err = GitserverRepos(db).GetByID(ctx, repo1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdatedAt")); diff != "" {
		t.Fatal(diff)
	}
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

func createTestRepos(ctx context.Context, t *testing.T, db dbutil.DB, repos types.Repos) {
	t.Helper()
	err := Repos(db).Create(ctx, repos...)
	if err != nil {
		t.Fatal(err)
	}
}

func createTestGitserverRepos(ctx context.Context, t *testing.T, db *sql.DB, hasLastError bool, repoID api.RepoID) {
	t.Helper()
	gitserverRepo := &types.GitserverRepo{
		RepoID:      repoID,
		ShardID:     fmt.Sprintf("gitserver%d", repoID),
		CloneStatus: types.CloneStatusNotCloned,
	}
	if hasLastError {
		gitserverRepo.LastError = "an error occurred"
	}
	if err := GitserverRepos(db).Upsert(ctx, gitserverRepo); err != nil {
		t.Fatal(err)
	}
}

package search

import (
	"context"
	"io/fs"
	"sort"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestFeatureFlaggedSelectOwnerJob(t *testing.T) {
	// We can run a quick exit check on the job runner since we don't need to use any clients or sender.
	t.Run("does not run if no features attached to job", func(t *testing.T) {
		selectJob := NewSelectOwnersJob(nil, nil)
		alert, err := selectJob.Run(context.Background(), job.RuntimeClients{}, nil)
		require.Nil(t, alert)
		var expectedErr *featureFlagError
		assert.ErrorAs(t, err, &expectedErr)
	})
	t.Run("does not run if own feature is false", func(t *testing.T) {
		selectJob := NewSelectOwnersJob(nil, &search.Features{CodeOwnershipSearch: false})
		alert, err := selectJob.Run(context.Background(), job.RuntimeClients{}, nil)
		require.Nil(t, alert)
		var expectedErr *featureFlagError
		assert.ErrorAs(t, err, &expectedErr)
	})
}

func TestGetCodeOwnersFromMatches(t *testing.T) {
	setupDB := func() *edb.MockEnterpriseDB {
		codeownersStore := edb.NewMockCodeownersStore()
		codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, nil)
		db := edb.NewMockEnterpriseDB()
		db.CodeownersFunc.SetDefaultReturn(codeownersStore)
		return db
	}

	t.Run("no results for no codeowners file", func(t *testing.T) {
		ctx := context.Background()

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.ReadFileFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, _ api.RepoName, _ api.CommitID, file string) ([]byte, error) {
			return nil, fs.ErrNotExist
		})

		rules := NewRulesCache(gitserverClient, setupDB())

		matches, err := getCodeOwnersFromMatches(ctx, &rules, []result.Match{
			&result.FileMatch{
				File: result.File{
					Path: "RepoWithNoCodeowners.md",
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, matches)
	})

	t.Run("no results for no owner matches", func(t *testing.T) {
		ctx := context.Background()

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.ReadFileFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, _ api.RepoName, _ api.CommitID, file string) ([]byte, error) {
			// return a codeowner path for no which doesn't match the path of the match below.
			return []byte("NO.md @test\n"), nil
		})
		rules := NewRulesCache(gitserverClient, setupDB())

		matches, err := getCodeOwnersFromMatches(ctx, &rules, []result.Match{
			&result.FileMatch{
				File: result.File{
					Path: "AnotherPath.md",
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, matches)
	})

	t.Run("returns person team and unknown owner matches", func(t *testing.T) {
		ctx := context.Background()

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.ReadFileFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, _ api.RepoName, _ api.CommitID, file string) ([]byte, error) {
			// README is owned by a user and a team.
			// code.go is owner by another user and an unknown entity.
			return []byte("README.md @testUserHandle @testTeamHandle\ncode.go user@email.com @unknown"), nil
		})
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := setupDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)

		rules := NewRulesCache(gitserverClient, db)

		personOwnerByHandle := newTestUser("testUserHandle")
		personOwnerByEmail := newTestUser("user@email.com")
		teamOwner := newTestTeam("testTeamHandle")

		mockUserStore.GetByUsernameFunc.SetDefaultHook(func(ctx context.Context, username string) (*types.User, error) {
			if username == "testUserHandle" {
				return personOwnerByHandle, nil
			}
			return nil, database.MockUserNotFoundErr
		})
		mockUserStore.GetByVerifiedEmailFunc.SetDefaultHook(func(ctx context.Context, email string) (*types.User, error) {
			if email == "user@email.com" {
				return personOwnerByEmail, nil
			}
			return nil, database.MockUserNotFoundErr
		})
		mockTeamStore.GetTeamByNameFunc.SetDefaultHook(func(ctx context.Context, name string) (*types.Team, error) {
			if name == "testTeamHandle" {
				return teamOwner, nil
			}
			return nil, database.TeamNotFoundError{}
		})

		matches, err := getCodeOwnersFromMatches(ctx, &rules, []result.Match{
			&result.FileMatch{
				File: result.File{
					Path: "README.md",
				},
			},
			&result.FileMatch{
				File: result.File{
					Path: "code.go",
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		want := []result.Match{
			&result.OwnerMatch{
				ResolvedOwner: &result.OwnerPerson{User: personOwnerByEmail, Email: "user@email.com"},
				InputRev:      nil,
				Repo:          types.MinimalRepo{},
				CommitID:      "",
				LimitHit:      0,
			},
			&result.OwnerMatch{
				ResolvedOwner: &result.OwnerPerson{Handle: "unknown"},
				InputRev:      nil,
				Repo:          types.MinimalRepo{},
				CommitID:      "",
				LimitHit:      0,
			},
			&result.OwnerMatch{
				ResolvedOwner: &result.OwnerPerson{User: personOwnerByHandle, Handle: "testUserHandle"},
				InputRev:      nil,
				Repo:          types.MinimalRepo{},
				CommitID:      "",
				LimitHit:      0,
			},
			&result.OwnerMatch{
				ResolvedOwner: &result.OwnerTeam{Team: teamOwner, Handle: "testTeamHandle"},
				InputRev:      nil,
				Repo:          types.MinimalRepo{},
				CommitID:      "",
				LimitHit:      0,
			},
		}
		sort.Slice(matches, func(x, y int) bool {
			return matches[x].Key().Less(matches[y].Key())
		})
		sort.Slice(want, func(x, y int) bool {
			return want[x].Key().Less(want[y].Key())
		})
		autogold.Expect(want).Equal(t, matches)
	})
}

func newTestUser(username string) *types.User {
	return &types.User{
		ID:          1,
		Username:    username,
		AvatarURL:   "https://sourcegraph.com/avatar/" + username,
		DisplayName: "User " + username,
	}
}

func newTestTeam(teamName string) *types.Team {
	return &types.Team{
		ID:          1,
		Name:        teamName,
		DisplayName: "Team " + teamName,
	}
}

package own_test

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type repoPath struct {
	Repo     api.RepoName
	CommitID api.CommitID
	Path     string
}

// repoFiles is a fake git client mapping a file
type repoFiles map[repoPath]string

func (fs repoFiles) ReadFile(_ context.Context, _ authz.SubRepoPermissionChecker, repoName api.RepoName, commitID api.CommitID, file string) ([]byte, error) {
	content, ok := fs[repoPath{Repo: repoName, CommitID: commitID, Path: file}]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(content), nil
}

func TestOwnersServesFilesAtVariousLocations(t *testing.T) {
	codeownersText := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{
				Pattern: "README.md",
				Owner:   []*codeownerspb.Owner{{Email: "owner@example.com"}},
			},
		},
	}).Repr()
	for name, repo := range map[string]repoFiles{
		"top-level": {{"repo", "SHA", "CODEOWNERS"}: codeownersText},
		".github":   {{"repo", "SHA", ".github/CODEOWNERS"}: codeownersText},
		".gitlab":   {{"repo", "SHA", ".gitlab/CODEOWNERS"}: codeownersText},
	} {
		t.Run(name, func(t *testing.T) {
			git := gitserver.NewMockClient()
			git.ReadFileFunc.SetDefaultHook(repo.ReadFile)

			codeownersStore := edb.NewMockCodeownersStore()
			codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, nil)
			db := edb.NewMockEnterpriseDB()
			db.CodeownersFunc.SetDefaultReturn(codeownersStore)

			got, err := own.NewService(git, db).RulesetForRepo(context.Background(), "repo", 1, "SHA")
			require.NoError(t, err)
			assert.Equal(t, codeownersText, got.Repr())
		})
	}
}

func TestOwnersCannotFindFile(t *testing.T) {
	codeownersFile := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{
				Pattern: "README.md",
				Owner:   []*codeownerspb.Owner{{Email: "owner@example.com"}},
			},
		},
	})
	repo := repoFiles{
		{"repo", "SHA", "notCODEOWNERS"}: codeownersFile.Repr(),
	}
	git := gitserver.NewMockClient()
	git.ReadFileFunc.SetDefaultHook(repo.ReadFile)

	codeownersStore := edb.NewMockCodeownersStore()
	codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, edb.CodeownersFileNotFoundError{})
	db := edb.NewMockEnterpriseDB()
	db.CodeownersFunc.SetDefaultReturn(codeownersStore)

	got, err := own.NewService(git, db).RulesetForRepo(context.Background(), "repo", 1, "SHA")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestOwnersServesIngestedFile(t *testing.T) {
	t.Run("return manually ingested codeowners file", func(t *testing.T) {
		codeownersProto := &codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pattern: "README.md",
					Owner:   []*codeownerspb.Owner{{Email: "owner@example.com"}},
				},
			},
		}
		codeownersText := codeowners.NewRuleset(codeownersProto).Repr()

		git := gitserver.NewMockClient()

		codeownersStore := edb.NewMockCodeownersStore()
		codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(&types.CodeownersFile{
			Proto: codeownersProto,
		}, nil)
		db := edb.NewMockEnterpriseDB()
		db.CodeownersFunc.SetDefaultReturn(codeownersStore)

		got, err := own.NewService(git, db).RulesetForRepo(context.Background(), "repo", 1, "SHA")
		require.NoError(t, err)
		assert.Equal(t, codeownersText, got.Repr())
	})
	t.Run("file not found and codeowners file does not exist return nil", func(t *testing.T) {
		git := gitserver.NewMockClient()
		git.ReadFileFunc.SetDefaultReturn(nil, nil)

		codeownersStore := edb.NewMockCodeownersStore()
		codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, edb.CodeownersFileNotFoundError{})
		db := edb.NewMockEnterpriseDB()
		db.CodeownersFunc.SetDefaultReturn(codeownersStore)

		got, err := own.NewService(git, db).RulesetForRepo(context.Background(), "repo", 1, "SHA")
		require.NoError(t, err)
		require.Nil(t, got)
	})
}

func TestResolveOwnersWithType(t *testing.T) {
	t.Run("no owners returns empty", func(t *testing.T) {
		git := gitserver.NewMockClient()
		got, err := own.NewService(git, database.NewMockDB()).ResolveOwnersWithType(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
	t.Run("no user or team match returns unknown owner", func(t *testing.T) {
		git := gitserver.NewMockClient()
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := own.NewService(git, db)

		mockUserStore.GetByUsernameFunc.SetDefaultReturn(nil, database.MockUserNotFoundErr)
		mockTeamStore.GetTeamByNameFunc.SetDefaultReturn(nil, database.TeamNotFoundError{})
		owners := []*codeownerspb.Owner{
			{Handle: "unknown"},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		assert.Equal(t, []codeowners.ResolvedOwner{
			newTestUnknownOwner("unknown", ""),
		}, got)
	})
	t.Run("user match from handle returns person owner", func(t *testing.T) {
		git := gitserver.NewMockClient()
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := own.NewService(git, db)

		handle := "person"
		testUser := newTestUser(handle)
		mockUserStore.GetByUsernameFunc.PushReturn(testUser, nil)
		mockTeamStore.GetTeamByNameFunc.SetDefaultReturn(nil, errors.New("I'm panicking because I should not be called"))
		owners := []*codeownerspb.Owner{
			{Handle: handle},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		assert.Equal(t, []codeowners.ResolvedOwner{
			&codeowners.Person{
				User:   testUser,
				Handle: handle,
			},
		}, got)
	})
	t.Run("user match from email returns person owner", func(t *testing.T) {
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := own.NewService(gitserver.NewMockClient(), db)

		email := "person@sourcegraph.com"
		testUser := newTestUser("person")
		mockUserStore.GetByVerifiedEmailFunc.PushReturn(testUser, nil)
		owners := []*codeownerspb.Owner{
			{Email: email},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		assert.Equal(t, []codeowners.ResolvedOwner{
			&codeowners.Person{
				User:  testUser,
				Email: email,
			},
		}, got)
	})
	t.Run("team match from handle returns team owner", func(t *testing.T) {
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := own.NewService(gitserver.NewMockClient(), db)

		handle := "team"
		testTeam := newTestTeam(handle)
		mockUserStore.GetByUsernameFunc.PushReturn(nil, database.MockUserNotFoundErr)
		mockTeamStore.GetTeamByNameFunc.PushReturn(testTeam, nil)
		owners := []*codeownerspb.Owner{
			{Handle: handle},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		assert.Equal(t, []codeowners.ResolvedOwner{
			&codeowners.Team{
				Team:   testTeam,
				Handle: handle,
			},
		}, got)
	})
	t.Run("no user match from email returns unknown owner", func(t *testing.T) {
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := own.NewService(gitserver.NewMockClient(), db)

		email := "superman"
		mockUserStore.GetByVerifiedEmailFunc.PushReturn(nil, database.MockUserNotFoundErr)
		owners := []*codeownerspb.Owner{
			{Email: email},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		assert.Equal(t, []codeowners.ResolvedOwner{
			newTestUnknownOwner("", email),
		}, got)
	})
	t.Run("mix of person, team, and unknown matches", func(t *testing.T) {
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := own.NewService(gitserver.NewMockClient(), db)

		userHandle := "userWithHandle"
		userEmail := "userWithEmail"
		teamHandle := "teamWithHandle"
		unknownOwnerEmail := "plato@sourcegraph.com"

		testUserWithHandle := newTestUser(userHandle)
		testUserWithEmail := newTestUser(userEmail)
		testTeamWithHandle := newTestTeam(teamHandle)
		testUnknownOwner := newTestUnknownOwner("", unknownOwnerEmail)

		mockUserStore.GetByUsernameFunc.SetDefaultHook(func(ctx context.Context, username string) (*itypes.User, error) {
			if username == userHandle {
				return testUserWithHandle, nil
			}
			return nil, database.MockUserNotFoundErr
		})
		mockUserStore.GetByVerifiedEmailFunc.SetDefaultHook(func(ctx context.Context, email string) (*itypes.User, error) {
			if email == userEmail {
				return testUserWithEmail, nil
			}
			return nil, database.MockUserNotFoundErr
		})
		mockTeamStore.GetTeamByNameFunc.SetDefaultHook(func(ctx context.Context, name string) (*itypes.Team, error) {
			if name == teamHandle {
				return testTeamWithHandle, nil
			}
			return nil, database.TeamNotFoundError{}
		})

		owners := []*codeownerspb.Owner{
			{Email: userEmail},
			{Handle: userHandle},
			{Email: unknownOwnerEmail},
			{Handle: teamHandle},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		want := []codeowners.ResolvedOwner{
			&codeowners.Person{User: testUserWithHandle, Handle: userHandle},
			&codeowners.Person{User: testUserWithEmail, Email: userEmail},
			&codeowners.Team{Team: testTeamWithHandle, Handle: teamHandle},
			testUnknownOwner,
		}
		sort.Slice(want, func(x, j int) bool {
			return want[x].Identifier() < want[j].Identifier()
		})
		sort.Slice(got, func(x, j int) bool {
			return got[x].Identifier() < got[j].Identifier()
		})
		assert.Equal(t, want, got)
	})
	t.Run("makes use of cache", func(t *testing.T) {
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := own.NewService(gitserver.NewMockClient(), db)

		email := "person@sourcegraph.com"
		testUser := newTestUser("person")
		mockUserStore.GetByVerifiedEmailFunc.PushReturn(testUser, nil)
		mockUserStore.GetByVerifiedEmailFunc.PushReturn(nil, errors.New("should have been cached"))
		owners := []*codeownerspb.Owner{
			{Email: email},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		assert.Equal(t, []codeowners.ResolvedOwner{
			&codeowners.Person{
				User:  testUser,
				Email: email,
			},
		}, got)
		// do it again
		got, err = ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		assert.Equal(t, []codeowners.ResolvedOwner{
			&codeowners.Person{
				User:  testUser,
				Email: email,
			},
		}, got)
	})
	t.Run("errors", func(t *testing.T) {
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := own.NewService(gitserver.NewMockClient(), db)

		email := "person@sourcegraph.com"
		var myError = errors.New("you shall not pass")
		mockUserStore.GetByVerifiedEmailFunc.PushReturn(nil, myError)
		owners := []*codeownerspb.Owner{
			{Email: email},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.Error(t, err)
		assert.ErrorIs(t, err, myError)
		assert.Empty(t, got)
	})
	t.Run("no errors if no handle or email", func(t *testing.T) {
		mockUserStore := database.NewMockUserStore()
		mockTeamStore := database.NewMockTeamStore()
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		ownService := own.NewService(gitserver.NewMockClient(), db)

		owners := []*codeownerspb.Owner{
			{},
		}

		got, err := ownService.ResolveOwnersWithType(context.Background(), owners)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func newTestUser(username string) *itypes.User {
	return &itypes.User{
		ID:          1,
		Username:    username,
		AvatarURL:   "https://sourcegraph.com/avatar/" + username,
		DisplayName: "User " + username,
	}
}

func newTestTeam(teamName string) *itypes.Team {
	return &itypes.Team{
		ID:          1,
		Name:        teamName,
		DisplayName: "Team " + teamName,
	}
}

// an unknown owner is just a person with no user set
func newTestUnknownOwner(handle, email string) codeowners.ResolvedOwner {
	return &codeowners.Person{
		Handle: handle,
		Email:  email,
	}
}

package own

import (
	"bytes"
	"context"
	"io"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	types2 "github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/own/types"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
)

const (
	repoOwnerID          = 71
	srcMainOwnerID       = 72
	srcMainSecondOwnerID = 73
	srcMainJavaOwnerID   = 74
	assignerID           = 76
	repoID               = 41
)

type repoPath struct {
	Repo     api.RepoName
	CommitID api.CommitID
	Path     string
}

// repoFiles is a fake git client mapping a file
type repoFiles map[repoPath]string

func (fs repoFiles) NewFileReader(_ context.Context, repoName api.RepoName, commitID api.CommitID, file string) (io.ReadCloser, error) {
	content, ok := fs[repoPath{Repo: repoName, CommitID: commitID, Path: file}]
	if !ok {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader([]byte(content))), nil
}

func TestOwnersServesFilesAtVariousLocations(t *testing.T) {
	codeownersText := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pattern: "README.md",
					Owner:   []*codeownerspb.Owner{{Email: "owner@example.com"}},
				},
			},
		},
	).Repr()
	for name, repo := range map[string]repoFiles{
		"top-level": {{"repo", "SHA", "CODEOWNERS"}: codeownersText},
		".github":   {{"repo", "SHA", ".github/CODEOWNERS"}: codeownersText},
		".gitlab":   {{"repo", "SHA", ".gitlab/CODEOWNERS"}: codeownersText},
	} {
		t.Run(name, func(t *testing.T) {
			git := gitserver.NewMockClient()
			git.NewFileReaderFunc.SetDefaultHook(repo.NewFileReader)

			reposStore := dbmocks.NewMockRepoStore()
			reposStore.GetFunc.SetDefaultReturn(&types2.Repo{ExternalRepo: api.ExternalRepoSpec{ServiceType: "github"}}, nil)
			codeownersStore := dbmocks.NewMockCodeownersStore()
			codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, nil)
			db := dbmocks.NewMockDB()
			db.ReposFunc.SetDefaultReturn(reposStore)
			db.CodeownersFunc.SetDefaultReturn(codeownersStore)

			got, err := NewService(git, db).RulesetForRepo(context.Background(), "repo", 1, "SHA")
			require.NoError(t, err)
			assert.Equal(t, codeownersText, got.Repr())
		})
	}
}

func TestOwnersCannotFindFile(t *testing.T) {
	codeownersFile := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pattern: "README.md",
					Owner:   []*codeownerspb.Owner{{Email: "owner@example.com"}},
				},
			},
		},
	)
	repo := repoFiles{
		{"repo", "SHA", "notCODEOWNERS"}: codeownersFile.Repr(),
	}
	git := gitserver.NewMockClient()
	git.NewFileReaderFunc.SetDefaultHook(repo.NewFileReader)

	codeownersStore := dbmocks.NewMockCodeownersStore()
	codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, database.CodeownersFileNotFoundError{})
	db := dbmocks.NewMockDB()
	db.CodeownersFunc.SetDefaultReturn(codeownersStore)
	reposStore := dbmocks.NewMockRepoStore()
	reposStore.GetFunc.SetDefaultReturn(&types2.Repo{ExternalRepo: api.ExternalRepoSpec{ServiceType: "github"}}, nil)
	db.ReposFunc.SetDefaultReturn(reposStore)
	got, err := NewService(git, db).RulesetForRepo(context.Background(), "repo", 1, "SHA")
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
		codeownersText := codeowners.NewRuleset(codeowners.IngestedRulesetSource{}, codeownersProto).Repr()

		git := gitserver.NewMockClient()

		codeownersStore := dbmocks.NewMockCodeownersStore()
		codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(&types.CodeownersFile{
			Proto: codeownersProto,
		}, nil)
		db := dbmocks.NewMockDB()
		db.CodeownersFunc.SetDefaultReturn(codeownersStore)
		reposStore := dbmocks.NewMockRepoStore()
		reposStore.GetFunc.SetDefaultReturn(&types2.Repo{ExternalRepo: api.ExternalRepoSpec{ServiceType: "github"}}, nil)
		db.ReposFunc.SetDefaultReturn(reposStore)

		got, err := NewService(git, db).RulesetForRepo(context.Background(), "repo", 1, "SHA")
		require.NoError(t, err)
		assert.Equal(t, codeownersText, got.Repr())
	})
	t.Run("file not found and codeowners file does not exist return nil", func(t *testing.T) {
		git := gitserver.NewMockClient()
		git.NewFileReaderFunc.SetDefaultReturn(nil, os.ErrNotExist)

		codeownersStore := dbmocks.NewMockCodeownersStore()
		codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, database.CodeownersFileNotFoundError{})
		db := dbmocks.NewMockDB()
		db.CodeownersFunc.SetDefaultReturn(codeownersStore)
		reposStore := dbmocks.NewMockRepoStore()
		reposStore.GetFunc.SetDefaultReturn(&types2.Repo{ExternalRepo: api.ExternalRepoSpec{ServiceType: "github"}}, nil)
		db.ReposFunc.SetDefaultReturn(reposStore)

		got, err := NewService(git, db).RulesetForRepo(context.Background(), "repo", 1, "SHA")
		require.NoError(t, err)
		require.Nil(t, got)
	})
}

func TestAssignedOwners(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating 2 users.
	user1, err := db.Users().Create(ctx, database.NewUser{Username: "user1"})
	require.NoError(t, err)
	user2, err := db.Users().Create(ctx, database.NewUser{Username: "user2"})
	require.NoError(t, err)

	// Create repo
	var repoID api.RepoID = 1
	require.NoError(t, db.Repos().Create(ctx, &itypes.Repo{
		ID:   repoID,
		Name: "github.com/sourcegraph/sourcegraph",
	}))

	store := db.AssignedOwners()
	require.NoError(t, store.Insert(ctx, user1.ID, repoID, "src/test", user2.ID))
	require.NoError(t, store.Insert(ctx, user2.ID, repoID, "src/test", user1.ID))
	require.NoError(t, store.Insert(ctx, user2.ID, repoID, "src/main", user1.ID))

	s := NewService(nil, db)
	var exampleCommitID api.CommitID = "sha"
	got, err := s.AssignedOwnership(ctx, repoID, exampleCommitID)
	// Erase the time for comparison
	for _, summaries := range got {
		for i := range summaries {
			summaries[i].AssignedAt = time.Time{}
		}
	}
	require.NoError(t, err)
	want := AssignedOwners{
		"src/test": []database.AssignedOwnerSummary{
			{
				OwnerUserID:       user1.ID,
				FilePath:          "src/test",
				RepoID:            repoID,
				WhoAssignedUserID: user2.ID,
			},
			{
				OwnerUserID:       user2.ID,
				FilePath:          "src/test",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
		},
		"src/main": []database.AssignedOwnerSummary{
			{
				OwnerUserID:       user2.ID,
				FilePath:          "src/main",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("AssignedOwnership -want+got: %s", diff)
	}
}

func TestAssignedOwnersMatch(t *testing.T) {
	var (
		repoOwner = database.AssignedOwnerSummary{
			OwnerUserID:       repoOwnerID,
			FilePath:          "",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
		srcMainOwner = database.AssignedOwnerSummary{
			OwnerUserID:       srcMainOwnerID,
			FilePath:          "src/main",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
		srcMainSecondOwner = database.AssignedOwnerSummary{
			OwnerUserID:       srcMainSecondOwnerID,
			FilePath:          "src/main",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
		srcMainJavaOwner = database.AssignedOwnerSummary{
			OwnerUserID:       srcMainJavaOwnerID,
			FilePath:          "src/main/java",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
		srcTestOwner = database.AssignedOwnerSummary{
			OwnerUserID:       srcMainJavaOwnerID,
			FilePath:          "src/test",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
	)
	owners := AssignedOwners{
		"": []database.AssignedOwnerSummary{
			repoOwner,
		},
		"src/main": []database.AssignedOwnerSummary{
			srcMainOwner,
			srcMainSecondOwner,
		},
		"src/main/java": []database.AssignedOwnerSummary{
			srcMainJavaOwner,
		},
		"src/test": []database.AssignedOwnerSummary{
			srcTestOwner,
		},
	}
	order := func(os []database.AssignedOwnerSummary) {
		sort.Slice(os, func(i, j int) bool {
			if os[i].OwnerUserID < os[j].OwnerUserID {
				return true
			}
			if os[i].FilePath < os[j].FilePath {
				return true
			}
			return false
		})
	}
	for _, testCase := range []struct {
		path string
		want []database.AssignedOwnerSummary
	}{
		{
			path: "",
			want: []database.AssignedOwnerSummary{
				repoOwner,
			},
		},
		{
			path: "resources/pom.xml",
			want: []database.AssignedOwnerSummary{
				repoOwner,
			},
		},
		{
			path: "src/main",
			want: []database.AssignedOwnerSummary{
				repoOwner,
				srcMainOwner,
				srcMainSecondOwner,
			},
		},
		{
			path: "src/main/java/com/sourcegraph/GitServer.java",
			want: []database.AssignedOwnerSummary{
				repoOwner,
				srcMainOwner,
				srcMainSecondOwner,
				srcMainJavaOwner,
			},
		},
		{
			path: "src/test/java/com/sourcegraph/GitServerTest.java",
			want: []database.AssignedOwnerSummary{
				repoOwner,
				srcTestOwner,
			},
		},
	} {
		got := owners.Match(testCase.path)
		order(got)
		order(testCase.want)
		if diff := cmp.Diff(testCase.want, got); diff != "" {
			t.Errorf("path: %q, unexpected owners (-want+got): %s", testCase.path, diff)
		}
	}
}

func TestAssignedTeams(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating a user and 2 teams.
	user1, err := db.Users().Create(ctx, database.NewUser{Username: "user1"})
	require.NoError(t, err)
	team1 := createTeam(t, ctx, db, "team-a")
	team2 := createTeam(t, ctx, db, "team-a2")

	// Create repo
	var repoID api.RepoID = 1
	require.NoError(t, db.Repos().Create(ctx, &itypes.Repo{
		ID:   repoID,
		Name: "github.com/sourcegraph/sourcegraph",
	}))

	store := db.AssignedTeams()
	require.NoError(t, store.Insert(ctx, team1.ID, repoID, "src/test", user1.ID))
	require.NoError(t, store.Insert(ctx, team2.ID, repoID, "src/test", user1.ID))
	require.NoError(t, store.Insert(ctx, team2.ID, repoID, "src/main", user1.ID))

	s := NewService(nil, db)
	var exampleCommitID api.CommitID = "sha"
	got, err := s.AssignedTeams(ctx, repoID, exampleCommitID)
	// Erase the time for comparison
	for _, summaries := range got {
		for i := range summaries {
			summaries[i].AssignedAt = time.Time{}
		}
	}
	require.NoError(t, err)
	want := AssignedTeams{
		"src/test": []database.AssignedTeamSummary{
			{
				OwnerTeamID:       team1.ID,
				FilePath:          "src/test",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
			{
				OwnerTeamID:       team2.ID,
				FilePath:          "src/test",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
		},
		"src/main": []database.AssignedTeamSummary{
			{
				OwnerTeamID:       team2.ID,
				FilePath:          "src/main",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("AssignedTeams -want+got: %s", diff)
	}
}

func TestAssignedTeamsMatch(t *testing.T) {
	var (
		repoOwner = database.AssignedTeamSummary{
			OwnerTeamID:       repoOwnerID,
			FilePath:          "",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
		srcMainOwner = database.AssignedTeamSummary{
			OwnerTeamID:       srcMainOwnerID,
			FilePath:          "src/main",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
		srcMainSecondOwner = database.AssignedTeamSummary{
			OwnerTeamID:       srcMainSecondOwnerID,
			FilePath:          "src/main",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
		srcMainJavaOwner = database.AssignedTeamSummary{
			OwnerTeamID:       srcMainJavaOwnerID,
			FilePath:          "src/main/java",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
		srcTestOwner = database.AssignedTeamSummary{
			OwnerTeamID:       srcMainJavaOwnerID,
			FilePath:          "src/test",
			RepoID:            repoID,
			WhoAssignedUserID: assignerID,
		}
	)
	owners := AssignedTeams{
		"": []database.AssignedTeamSummary{
			repoOwner,
		},
		"src/main": []database.AssignedTeamSummary{
			srcMainOwner,
			srcMainSecondOwner,
		},
		"src/main/java": []database.AssignedTeamSummary{
			srcMainJavaOwner,
		},
		"src/test": []database.AssignedTeamSummary{
			srcTestOwner,
		},
	}
	order := func(os []database.AssignedTeamSummary) {
		sort.Slice(os, func(i, j int) bool {
			if os[i].OwnerTeamID < os[j].OwnerTeamID {
				return true
			}
			if os[i].FilePath < os[j].FilePath {
				return true
			}
			return false
		})
	}
	for _, testCase := range []struct {
		path string
		want []database.AssignedTeamSummary
	}{
		{
			path: "",
			want: []database.AssignedTeamSummary{
				repoOwner,
			},
		},
		{
			path: "resources/pom.xml",
			want: []database.AssignedTeamSummary{
				repoOwner,
			},
		},
		{
			path: "src/main",
			want: []database.AssignedTeamSummary{
				repoOwner,
				srcMainOwner,
				srcMainSecondOwner,
			},
		},
		{
			path: "src/main/java/com/sourcegraph/GitServer.java",
			want: []database.AssignedTeamSummary{
				repoOwner,
				srcMainOwner,
				srcMainSecondOwner,
				srcMainJavaOwner,
			},
		},
		{
			path: "src/test/java/com/sourcegraph/GitServerTest.java",
			want: []database.AssignedTeamSummary{
				repoOwner,
				srcTestOwner,
			},
		},
	} {
		got := owners.Match(testCase.path)
		order(got)
		order(testCase.want)
		if diff := cmp.Diff(testCase.want, got); diff != "" {
			t.Errorf("path: %q, unexpected owners (-want+got): %s", testCase.path, diff)
		}
	}
}

func createTeam(t *testing.T, ctx context.Context, db database.DB, teamName string) *itypes.Team {
	t.Helper()
	team, err := db.Teams().CreateTeam(ctx, &itypes.Team{Name: teamName})
	require.NoError(t, err)
	return team
}

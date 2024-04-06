package search

import (
	"bytes"
	"context"
	"hash/fnv"
	"io"
	"io/fs"
	"sort"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/mockjob"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetCodeOwnersFromMatches(t *testing.T) {
	setupDB := func() *dbmocks.MockDB {
		codeownersStore := dbmocks.NewMockCodeownersStore()
		codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, nil)
		repoStore := dbmocks.NewMockRepoStore()
		repoStore.GetFunc.SetDefaultReturn(&types.Repo{ExternalRepo: api.ExternalRepoSpec{ServiceType: "github"}}, nil)
		db := dbmocks.NewMockDB()
		db.CodeownersFunc.SetDefaultReturn(codeownersStore)
		db.AssignedOwnersFunc.SetDefaultReturn(dbmocks.NewMockAssignedOwnersStore())
		db.AssignedTeamsFunc.SetDefaultReturn(dbmocks.NewMockAssignedTeamsStore())
		db.ReposFunc.SetDefaultReturn(repoStore)
		return db
	}

	t.Run("no results for no codeowners file", func(t *testing.T) {
		ctx := context.Background()

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.NewFileReaderFunc.SetDefaultReturn(nil, fs.ErrNotExist)

		rules := NewRulesCache(gitserverClient, setupDB())

		matches, hasNoResults, err := getCodeOwnersFromMatches(ctx, &rules, []result.Match{
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
		assert.Equal(t, true, hasNoResults)
	})

	t.Run("no results for no owner matches", func(t *testing.T) {
		ctx := context.Background()

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, s string) (io.ReadCloser, error) {
			// return a codeowner path for no which doesn't match the path of the match below.
			return io.NopCloser(bytes.NewReader([]byte("NO.md @test\n"))), nil
		})
		rules := NewRulesCache(gitserverClient, setupDB())

		matches, hasNoResults, err := getCodeOwnersFromMatches(ctx, &rules, []result.Match{
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
		assert.Equal(t, true, hasNoResults)
	})

	t.Run("returns person team and unknown owner matches", func(t *testing.T) {
		ctx := context.Background()

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, s string) (io.ReadCloser, error) {
			// README is owned by a user and a team.
			// code.go is owner by another user and an unknown entity.
			return io.NopCloser(bytes.NewReader([]byte("README.md @testUserHandle @testTeamHandle\ncode.go user@email.com @unknown"))), nil
		})
		mockUserStore := dbmocks.NewMockUserStore()
		mockTeamStore := dbmocks.NewMockTeamStore()
		mockEmailStore := dbmocks.NewMockUserEmailsStore()
		db := setupDB()
		db.UsersFunc.SetDefaultReturn(mockUserStore)
		db.UserEmailsFunc.SetDefaultReturn(mockEmailStore)
		db.TeamsFunc.SetDefaultReturn(mockTeamStore)
		db.AssignedOwnersFunc.SetDefaultReturn(dbmocks.NewMockAssignedOwnersStore())
		db.AssignedTeamsFunc.SetDefaultReturn(dbmocks.NewMockAssignedTeamsStore())
		db.UserExternalAccountsFunc.SetDefaultReturn(dbmocks.NewMockUserExternalAccountsStore())

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
		mockEmailStore.ListByUserFunc.SetDefaultHook(func(_ context.Context, opts database.UserEmailsListOptions) ([]*database.UserEmail, error) {
			switch opts.UserID {
			case personOwnerByEmail.ID:
				return []*database.UserEmail{
					{
						UserID: personOwnerByEmail.ID,
						Email:  "user@email.com",
					},
				}, nil
			default:
				return nil, nil
			}
		})
		mockTeamStore.GetTeamByNameFunc.SetDefaultHook(func(ctx context.Context, name string) (*types.Team, error) {
			if name == "testTeamHandle" {
				return teamOwner, nil
			}
			return nil, database.TeamNotFoundError{}
		})

		mockJob := mockjob.NewMockJob()
		mockJob.RunFunc.SetDefaultHook(func(ctx context.Context, _ job.RuntimeClients, s streaming.Sender) (*search.Alert, error) {
			s.Send(streaming.SearchEvent{
				Results: []result.Match{
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
				},
			})
			return nil, nil
		})
		j := &selectOwnersJob{
			child: mockJob,
		}
		clients := job.RuntimeClients{
			Gitserver: gitserverClient,
			DB:        db,
		}
		s := streaming.NewAggregatingStream()
		_, err := j.Run(ctx, clients, s) // TODO: handle alert
		if err != nil {
			t.Fatal(err)
		}
		want := result.Matches{
			&result.OwnerMatch{
				ResolvedOwner: &result.OwnerPerson{
					User:   personOwnerByEmail,
					Email:  "user@email.com",
					Handle: "user@email.com", // This is username in the mock storage.
				},
				InputRev: nil,
				Repo:     types.MinimalRepo{},
				CommitID: "",
				LimitHit: 0,
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
		matches := s.Results
		sort.Slice(matches, func(x, y int) bool {
			return matches[x].Key().Less(matches[y].Key())
		})
		sort.Slice(want, func(x, y int) bool {
			return want[x].Key().Less(want[y].Key())
		})
		autogold.Expect(want).Equal(t, matches)
		// TODO: What about hasnoresults?
	})
}

func newTestUser(username string) *types.User {
	h := fnv.New32a()
	h.Write([]byte(username))
	return &types.User{
		ID:          int32(h.Sum32()),
		Username:    username,
		AvatarURL:   "https://sourcegraph.com/avatar/" + username,
		DisplayName: "User " + username,
	}
}

func newTestTeam(teamName string) *types.Team {
	h := fnv.New32a()
	h.Write([]byte(teamName))
	return &types.Team{
		ID:          int32(h.Sum32()),
		Name:        teamName,
		DisplayName: "Team " + teamName,
	}
}

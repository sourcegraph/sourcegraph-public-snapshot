package search

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestApplyCodeOwnershipFiltering(t *testing.T) {
	type args struct {
		includeOwners []string
		excludeOwners []string
		matches       []result.Match
		repoContent   map[string]string
	}
	tests := []struct {
		name  string
		args  args
		setup func(db *dbmocks.MockDB)
		want  autogold.Value
	}{
		{
			// TODO: We should display an error in search describing why the result is empty.
			name: "filters all matches if we include an owner and have no code owners file",
			args: args{
				includeOwners: []string{"@test"},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
				},
			},
			want: autogold.Expect([]result.Match{}),
		},
		{
			name: "selects only results matching owners",
			args: args{
				includeOwners: []string{"@test"},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "package.json",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "README.md @test\n",
				},
			},
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "README.md",
					},
				},
			}),
		},
		{
			name: "match username without search term containing a leading @",
			args: args{
				includeOwners: []string{"test"},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "README.md @test\n",
				},
			},
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "README.md",
					},
				},
			}),
		},
		{
			name: "match on email",
			args: args{
				includeOwners: []string{"test@example.com"},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "README.md test@example.com\n",
				},
			},
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "README.md",
					},
				},
			}),
		},
		{
			name: "selects only results without excluded owners",
			args: args{
				includeOwners: []string{},
				excludeOwners: []string{"@test"},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "package.json",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "README.md @test\n",
				},
			},
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "package.json",
					},
				},
			}),
		},
		{
			name: "do not match on email if search term includes leading @",
			args: args{
				includeOwners: []string{"@test@example.com"},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "README.md test@example.com\n",
				},
			},
			want: autogold.Expect([]result.Match{}),
		},
		{
			name: "selects results with any owner assigned",
			args: args{
				includeOwners: []string{""},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "package.json",
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "/test/AbstractFactoryTest.java",
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "/test/fixture-data.json",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": strings.Join([]string{
						"README.md @test",
						"/test/* @example",
						"/test/*.json", // explicitly unassigned ownership
					}, "\n"),
				},
			},
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "README.md",
					},
				},
				&result.FileMatch{
					File: result.File{
						Path: "/test/AbstractFactoryTest.java",
					},
				},
			}),
		},
		{
			name: "selects results without an owner",
			args: args{
				includeOwners: []string{},
				excludeOwners: []string{""},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "package.json",
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "/test/AbstractFactoryTest.java",
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "/test/fixture-data.json",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": strings.Join([]string{
						"README.md @test",
						"/test/* @example",
						"/test/*.json", // explicitly unassigned ownership
					}, "\n"),
				},
			},
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "package.json",
					},
				},
				&result.FileMatch{
					File: result.File{
						Path: "/test/fixture-data.json",
					},
				},
			}),
		},
		{
			name: "selects result with assigned owner",
			args: args{
				includeOwners: []string{"test"},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "src/main/README.md",
						},
					},
				},
				// No CODEOWNERS
				repoContent: map[string]string{},
			},
			setup: assignedOwnerSetup(
				"src/main",
				&types.User{
					ID:       42,
					Username: "test",
				},
			),
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "src/main/README.md",
					},
				},
			}),
		},
		{
			name: "selects results with AND-ed include owners specified",
			args: args{
				includeOwners: []string{"assigned", "codeowner"},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							// assigned owns src/main,
							// but @codeowner does not own the file
							Path: "src/main/onlyAssigned.md",
						},
					},
					&result.FileMatch{
						File: result.File{
							// @codeowner owns all go files,
							// and assigned owns src/main
							Path: "src/main/bothMatch.go",
						},
					},
					&result.FileMatch{
						File: result.File{
							// @codeowner owns all go files
							// but assigned only owns src/main
							// and this is in src/test.
							Path: "src/test/onlyCodeowner.go",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "*.go @codeowner",
				},
			},
			setup: assignedOwnerSetup(
				"src/main",
				&types.User{
					ID:       42,
					Username: "assigned",
				},
			),
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "src/main/bothMatch.go",
					},
				},
			}),
		},
		{
			name: "selects results with exclude owner and include owner specified",
			args: args{
				includeOwners: []string{"codeowner"},
				excludeOwners: []string{"assigned"},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							// assigned owns src/main,
							// but @codeowner does not own the file
							Path: "src/main/onlyAssigned.md",
						},
					},
					&result.FileMatch{
						File: result.File{
							// @codeowner owns all go files,
							// and assigned owns src/main
							Path: "src/main/bothMatch.go",
						},
					},
					&result.FileMatch{
						File: result.File{
							// @codeowner owns all go files
							// but assigned only owns src/main
							// and this is in src/test.
							Path: "src/test/onlyCodeowner.go",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "*.go @codeowner",
				},
			},
			setup: assignedOwnerSetup(
				"src/main",
				&types.User{
					ID:       42,
					Username: "assigned",
				},
			),
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "src/test/onlyCodeowner.go",
					},
				},
			}),
		},
		{
			name: "selects results with AND-ed exclude owners specified",
			args: args{
				includeOwners: []string{},
				excludeOwners: []string{"assigned", "codeowner"},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							// assigned owns src/main,
							// but @codeowner does not own the file
							Path: "src/main/onlyAssigned.md",
						},
					},
					&result.FileMatch{
						File: result.File{
							// @codeowner owns all go files,
							// and assigned owns src/main
							Path: "src/main/bothMatch.go",
						},
					},
					&result.FileMatch{
						File: result.File{
							// @codeowner owns all go files
							// but assigned only owns src/main
							// and this is in src/test.
							Path: "src/test/onlyCodeowner.go",
						},
					},
					&result.FileMatch{
						File: result.File{
							// @codeowner owns all go files
							// but assigned only owns src/main
							// and this is in src/test.
							Path: "src/test/noOwners.txt",
						},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "*.go @codeowner",
				},
			},
			setup: assignedOwnerSetup(
				"src/main",
				&types.User{
					ID:       42,
					Username: "assigned",
				},
			),
			want: autogold.Expect([]result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "src/test/noOwners.txt",
					},
				},
			}),
		},
		{
			name: "match commits where any file is owned by included owner",
			args: args{
				includeOwners: []string{"@owner"},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.CommitMatch{
						ModifiedFiles: []string{"file1.notOwned", "file2.owned"},
					},
					&result.CommitMatch{
						ModifiedFiles: []string{"file3.notOwned", "file4.notOwned"},
					},
					&result.CommitMatch{
						ModifiedFiles: []string{"file5.owned"},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "*.owned @owner\n",
				},
			},
			want: autogold.Expect([]result.Match{
				&result.CommitMatch{
					ModifiedFiles: []string{"file1.notOwned", "file2.owned"},
				},
				&result.CommitMatch{
					ModifiedFiles: []string{"file5.owned"},
				},
			}),
		},
		{
			name: "discard commits where any file is owned by excluded owner",
			args: args{
				includeOwners: []string{},
				excludeOwners: []string{"@owner"},
				matches: []result.Match{
					&result.CommitMatch{
						ModifiedFiles: []string{"file1.notOwned", "file2.owned"},
					},
					&result.CommitMatch{
						ModifiedFiles: []string{"file3.notOwned", "file4.notOwned"},
					},
					&result.CommitMatch{
						ModifiedFiles: []string{"file5.owned"},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": "*.owned @owner\n",
				},
			},
			want: autogold.Expect([]result.Match{
				&result.CommitMatch{
					ModifiedFiles: []string{"file3.notOwned", "file4.notOwned"},
				},
			}),
		},
		{
			name: "discard commits through exclude owners despite having include owners",
			args: args{
				includeOwners: []string{"@includeOwner"},
				excludeOwners: []string{"@excludeOwner"},
				matches: []result.Match{
					&result.CommitMatch{
						ModifiedFiles: []string{"file1.included", "file2"},
					},
					&result.CommitMatch{
						ModifiedFiles: []string{"file3.included", "file4.excluded"},
					},
					&result.CommitMatch{
						ModifiedFiles: []string{"file5.excluded", "file3"},
					},
				},
				repoContent: map[string]string{
					"CODEOWNERS": strings.Join([]string{
						"*.included @includeOwner",
						"*.excluded @excludeOwner",
					}, "\n"),
				},
			},
			want: autogold.Expect([]result.Match{
				&result.CommitMatch{
					ModifiedFiles: []string{"file1.included", "file2"},
				},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			gitserverClient := gitserver.NewMockClient()
			gitserverClient.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, file string) (io.ReadCloser, error) {
				content, ok := tt.args.repoContent[file]
				if !ok {
					return nil, fs.ErrNotExist
				}
				return io.NopCloser(bytes.NewReader([]byte(content))), nil
			})

			codeownersStore := dbmocks.NewMockCodeownersStore()
			codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, nil)
			db := dbmocks.NewMockDB()
			db.CodeownersFunc.SetDefaultReturn(codeownersStore)
			usersStore := dbmocks.NewMockUserStore()
			usersStore.GetByUsernameFunc.SetDefaultReturn(nil, nil)
			usersStore.GetByVerifiedEmailFunc.SetDefaultReturn(nil, nil)
			db.UsersFunc.SetDefaultReturn(usersStore)
			usersEmailsStore := dbmocks.NewMockUserEmailsStore()
			usersEmailsStore.GetVerifiedEmailsFunc.SetDefaultReturn(nil, nil)
			db.UserEmailsFunc.SetDefaultReturn(usersEmailsStore)
			assignedOwnersStore := dbmocks.NewMockAssignedOwnersStore()
			assignedOwnersStore.ListAssignedOwnersForRepoFunc.SetDefaultReturn(nil, nil)
			db.AssignedOwnersFunc.SetDefaultReturn(assignedOwnersStore)
			assignedTeamsStore := dbmocks.NewMockAssignedTeamsStore()
			assignedTeamsStore.ListAssignedTeamsForRepoFunc.SetDefaultReturn(nil, nil)
			db.AssignedTeamsFunc.SetDefaultReturn(assignedTeamsStore)
			userExternalAccountsStore := dbmocks.NewMockUserExternalAccountsStore()
			userExternalAccountsStore.ListFunc.SetDefaultReturn(nil, nil)
			db.UserExternalAccountsFunc.SetDefaultReturn(userExternalAccountsStore)
			db.TeamsFunc.SetDefaultReturn(dbmocks.NewMockTeamStore())
			repoStore := dbmocks.NewMockRepoStore()
			repoStore.GetFunc.SetDefaultReturn(&types.Repo{ExternalRepo: api.ExternalRepoSpec{ServiceType: "github"}}, nil)
			db.ReposFunc.SetDefaultReturn(repoStore)
			if tt.setup != nil {
				tt.setup(db)
			}

			// TODO(#52450): Invoke filterHasOwnersJob.Run rather than duplicate code here.
			rules := NewRulesCache(gitserverClient, db)

			var includeBags []own.Bag
			for _, o := range tt.args.includeOwners {
				b := own.ByTextReference(ctx, db, o)
				includeBags = append(includeBags, b)
			}
			var excludeBags []own.Bag
			for _, o := range tt.args.excludeOwners {
				b := own.ByTextReference(ctx, db, o)
				excludeBags = append(excludeBags, b)
			}
			matches, _ := applyCodeOwnershipFiltering(
				ctx,
				&rules,
				includeBags,
				tt.args.includeOwners,
				excludeBags,
				tt.args.excludeOwners,
				tt.args.matches)
			tt.want.Equal(t, matches)
		})
	}
}

func assignedOwnerSetup(path string, user *types.User) func(*dbmocks.MockDB) {
	return func(db *dbmocks.MockDB) {
		assignedOwners := []*database.AssignedOwnerSummary{
			{
				OwnerUserID: user.ID,
				FilePath:    path,
			},
		}
		usersStore := dbmocks.NewMockUserStore()
		usersStore.GetByUsernameFunc.SetDefaultHook(func(_ context.Context, name string) (*types.User, error) {
			if name == user.Username {
				return user, nil
			}
			return nil, database.NewUserNotFoundErr()
		})
		usersStore.GetByVerifiedEmailFunc.SetDefaultReturn(nil, nil)
		db.UsersFunc.SetDefaultReturn(usersStore)
		assignedOwnersStore := dbmocks.NewMockAssignedOwnersStore()
		assignedOwnersStore.ListAssignedOwnersForRepoFunc.SetDefaultReturn(assignedOwners, nil)
		db.AssignedOwnersFunc.SetDefaultReturn(assignedOwnersStore)
	}
}

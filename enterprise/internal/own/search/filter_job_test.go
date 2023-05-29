package search

import (
	"context"
	"io/fs"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestFeatureFlaggedFileHasOwnerJob(t *testing.T) {
	// We can run a quick exit check on the job runner since we don't need to use any clients or sender.
	t.Run("does not run if no features attached to job", func(t *testing.T) {
		selectJob := NewFileHasOwnersJob(nil, nil, nil, nil)
		alert, err := selectJob.Run(context.Background(), job.RuntimeClients{}, nil)
		require.Nil(t, alert)
		var expectedErr *featureFlagError
		assert.ErrorAs(t, err, &expectedErr)
	})
	t.Run("does not run if own feature is false", func(t *testing.T) {
		selectJob := NewFileHasOwnersJob(nil, &search.Features{CodeOwnershipSearch: false}, nil, nil)
		alert, err := selectJob.Run(context.Background(), job.RuntimeClients{}, nil)
		require.Nil(t, alert)
		var expectedErr *featureFlagError
		assert.ErrorAs(t, err, &expectedErr)
	})
}

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
		setup func(db *edb.MockEnterpriseDB)
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
			name: "selects results with AND-ed owners specified",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			gitserverClient := gitserver.NewMockClient()
			gitserverClient.ReadFileFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, _ api.RepoName, _ api.CommitID, file string) ([]byte, error) {
				content, ok := tt.args.repoContent[file]
				if !ok {
					return nil, fs.ErrNotExist
				}
				return []byte(content), nil
			})

			codeownersStore := edb.NewMockCodeownersStore()
			codeownersStore.GetCodeownersForRepoFunc.SetDefaultReturn(nil, nil)
			db := edb.NewMockEnterpriseDB()
			db.CodeownersFunc.SetDefaultReturn(codeownersStore)
			usersStore := database.NewMockUserStore()
			usersStore.GetByUsernameFunc.SetDefaultReturn(nil, nil)
			usersStore.GetByVerifiedEmailFunc.SetDefaultReturn(nil, nil)
			db.UsersFunc.SetDefaultReturn(usersStore)
			usersEmailsStore := database.NewMockUserEmailsStore()
			usersEmailsStore.GetVerifiedEmailsFunc.SetDefaultReturn(nil, nil)
			db.UserEmailsFunc.SetDefaultReturn(usersEmailsStore)
			assignedOwnersStore := database.NewMockAssignedOwnersStore()
			assignedOwnersStore.ListAssignedOwnersForRepoFunc.SetDefaultReturn(nil, nil)
			db.AssignedOwnersFunc.SetDefaultReturn(assignedOwnersStore)
			userExternalAccountsStore := database.NewMockUserExternalAccountsStore()
			userExternalAccountsStore.ListFunc.SetDefaultReturn(nil, nil)
			db.UserExternalAccountsFunc.SetDefaultReturn(userExternalAccountsStore)
			db.TeamsFunc.SetDefaultReturn(database.NewMockTeamStore())
			if tt.setup != nil {
				tt.setup(db)
			}

			// TODO(#52450): Invoke filterHasOwnersJob.Run rather than duplicate code here.
			rules := NewRulesCache(gitserverClient, db)

			var includeBags []own.Bag
			for _, o := range tt.args.includeOwners {
				b, err := own.ByTextReference(ctx, db, o)
				require.NoError(t, err)
				includeBags = append(includeBags, b)
			}
			var excludeBags []own.Bag
			for _, o := range tt.args.excludeOwners {
				b, err := own.ByTextReference(ctx, db, o)
				require.NoError(t, err)
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
			//require.NoError(t, err)
			tt.want.Equal(t, matches)
		})
	}
}

func assignedOwnerSetup(path string, user *types.User) func(*edb.MockEnterpriseDB) {
	return func(db *edb.MockEnterpriseDB) {
		assignedOwners := []*database.AssignedOwnerSummary{
			{
				OwnerUserID: user.ID,
				FilePath:    path,
			},
		}
		usersStore := database.NewMockUserStore()
		usersStore.GetByUsernameFunc.SetDefaultHook(func(_ context.Context, name string) (*types.User, error) {
			if name == user.Username {
				return user, nil
			}
			return nil, database.NewUserNotFoundErr()
		})
		usersStore.GetByVerifiedEmailFunc.SetDefaultReturn(nil, nil)
		db.UsersFunc.SetDefaultReturn(usersStore)
		assignedOwnersStore := database.NewMockAssignedOwnersStore()
		assignedOwnersStore.ListAssignedOwnersForRepoFunc.SetDefaultReturn(assignedOwners, nil)
		db.AssignedOwnersFunc.SetDefaultReturn(assignedOwnersStore)
	}
}

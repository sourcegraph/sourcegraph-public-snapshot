package search

import (
	"context"
	"io/fs"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func TestApplyCodeOwnershipFiltering(t *testing.T) {
	type args struct {
		includeOwners []string
		excludeOwners []string
		matches       []result.Match
		repoContent   map[string]string
	}
	tests := []struct {
		name string
		args args
		want autogold.Value
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

			rules := NewRulesCache(gitserverClient, db)

			matches, _ := applyCodeOwnershipFiltering(ctx, &rules, tt.args.includeOwners, tt.args.excludeOwners, tt.args.matches)

			tt.want.Equal(t, matches)
		})
	}
}

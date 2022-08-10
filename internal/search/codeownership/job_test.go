package codeownership

import (
	"context"
	"errors"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func Test_applyCodeOwnershipFiltering(t *testing.T) {
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
			name: "filters all matches if we include an owner and have no code owners file",
			args: args{
				includeOwners: []string{"@sqs"},
				excludeOwners: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
				},
			},
			want: autogold.Want("no results", []result.Match{}),
		},
		{
			name: "filters results based on code owners file",
			args: args{
				includeOwners: []string{"@sqs"},
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
					"CODEOWNERS": "README.md @sqs\n",
				},
			},
			want: autogold.Want("results matching ownership", []result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "README.md",
					},
				},
			}),
		},
		{
			name: "filters results based on code owners file in a subdirectory",
			args: args{
				includeOwners: []string{"@sqs"},
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
					".github/CODEOWNERS": "README.md @sqs\n",
				},
			},
			want: autogold.Want("results matching ownership", []result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "README.md",
					},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := database.NewMockDB()
			rules := NewRulesCache()

			gitserver.Mocks.ReadFile = func(_ api.CommitID, file string) ([]byte, error) {
				content, ok := tt.args.repoContent[file]
				if !ok {
					return nil, errors.New("file does not exist")
				}
				return []byte(content), nil
			}
			t.Cleanup(func() { gitserver.Mocks.ReadFile = nil })

			matches, _ := applyCodeOwnershipFiltering(ctx, gitserver.NewClient(db), &rules, tt.args.includeOwners, tt.args.excludeOwners, tt.args.matches)

			tt.want.Equal(t, matches)
		})
	}
}

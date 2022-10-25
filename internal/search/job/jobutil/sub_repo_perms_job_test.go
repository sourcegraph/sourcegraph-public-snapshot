package jobutil

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestApplySubRepoFiltering(t *testing.T) {
	unauthorizedFileName := "README.md"
	errorFileName := "file.go"
	var userWithSubRepoPerms int32 = 1234

	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultReturn(true)
	checker.PermissionsFunc.SetDefaultHook(func(c context.Context, user int32, rc authz.RepoContent) (authz.Perms, error) {
		if user == userWithSubRepoPerms {
			switch rc.Path {
			case unauthorizedFileName:
				// This file should be filtered out
				return authz.None, nil
			case errorFileName:
				// Simulate an error case, should be filtered out
				return authz.None, errors.New(errorFileName)
			}
		}
		return authz.Read, nil
	})
	checker.FilePermissionsFuncFunc.SetDefaultHook(func(ctx context.Context, userID int32, repo api.RepoName) (authz.FilePermissionFunc, error) {
		return func(path string) (authz.Perms, error) {
			return checker.Permissions(ctx, userID, authz.RepoContent{Repo: repo, Path: path})
		}, nil
	})

	type args struct {
		ctxActor *actor.Actor
		matches  []result.Match
	}
	tests := []struct {
		name        string
		args        args
		wantMatches []result.Match
		wantErr     string
	}{
		{
			name: "read from user with no perms",
			args: args{
				ctxActor: actor.FromUser(789),
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: unauthorizedFileName,
						},
					},
				},
			},
			wantMatches: []result.Match{
				&result.FileMatch{
					File: result.File{
						Path: unauthorizedFileName,
					},
				},
			},
		},
		{
			name: "read for user with sub-repo perms",
			args: args{
				ctxActor: actor.FromUser(userWithSubRepoPerms),
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "not-unauthorized.md",
						},
					},
				},
			},
			wantMatches: []result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "not-unauthorized.md",
					},
				},
			},
		},
		{
			name: "drop match due to auth for user with sub-repo perms",
			args: args{
				ctxActor: actor.FromUser(userWithSubRepoPerms),
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: unauthorizedFileName,
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "random-name.md",
						},
					},
				},
			},
			wantMatches: []result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "random-name.md",
					},
				},
			},
		},
		{
			name: "drop match due to auth for user with sub-repo perms and error",
			args: args{
				ctxActor: actor.FromUser(userWithSubRepoPerms),
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: errorFileName,
						},
					},
					&result.FileMatch{
						File: result.File{
							Path: "random-name.md",
						},
					},
				},
			},
			wantMatches: []result.Match{
				&result.FileMatch{
					File: result.File{
						Path: "random-name.md",
					},
				},
			},
			wantErr: "subRepoFilterFunc",
		},
		{
			name: "repo matches should be ignored",
			args: args{
				ctxActor: actor.FromUser(userWithSubRepoPerms),
				matches: []result.Match{
					&result.RepoMatch{
						Name: "foo",
						ID:   1,
					},
				},
			},
			wantMatches: []result.Match{
				&result.RepoMatch{
					Name: "foo",
					ID:   1,
				},
			},
		},
		{
			name: "should filter commit matches where the user doesn't have access to any file in the ModifiedFiles",
			args: args{
				ctxActor: actor.FromUser(userWithSubRepoPerms),
				matches: []result.Match{
					&result.CommitMatch{
						ModifiedFiles: []string{unauthorizedFileName},
					},
					&result.CommitMatch{
						ModifiedFiles: []string{unauthorizedFileName, "another-file.txt"},
					},
				},
			},
			wantMatches: []result.Match{
				&result.CommitMatch{
					ModifiedFiles: []string{unauthorizedFileName, "another-file.txt"},
				},
			},
		},
		{
			name: "should filter commit matches where the diff is empty",
			args: args{
				ctxActor: actor.FromUser(userWithSubRepoPerms),
				matches: []result.Match{
					&result.CommitMatch{
						ModifiedFiles: []string{unauthorizedFileName, "another-file.txt"},
						DiffPreview:   &result.MatchedString{Content: ""},
					},
				},
			},
			wantMatches: []result.Match{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), tt.args.ctxActor)
			matches, err := applySubRepoFiltering(ctx, logtest.Scoped(t), checker, tt.args.matches)
			if diff := cmp.Diff(matches, tt.wantMatches, cmpopts.IgnoreUnexported(search.RepoStatusMap{})); diff != "" {
				t.Fatal(diff)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected err, got none")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected err %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}

package run

import (
	"context"
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

func TestApplySubRepoPerms(t *testing.T) {
	unauthorizedFileName := "README.md"
	errorFileName := "file.go"
	var userWithSubRepoPerms int32 = 1234

	srp := authz.NewMockSubRepoPermissionChecker()
	srp.EnabledFunc.SetDefaultReturn(true)
	srp.PermissionsFunc.SetDefaultHook(func(c context.Context, user int32, rc authz.RepoContent) (authz.Perms, error) {
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

	type args struct {
		ctxActor *actor.Actor
		event    *streaming.SearchEvent
	}
	tests := []struct {
		name      string
		args      args
		wantEvent *streaming.SearchEvent
		wantErr   string
	}{
		{
			name: "read from user with no perms",
			args: args{
				ctxActor: actor.FromUser(789),
				event: &streaming.SearchEvent{
					Results: []result.Match{
						&result.FileMatch{
							File: result.File{
								Path: unauthorizedFileName,
							},
						},
					},
				},
			},
			wantEvent: &streaming.SearchEvent{
				Results: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: unauthorizedFileName,
						},
					},
				},
			},
		},
		{
			name: "read for user with sub-repo perms",
			args: args{
				ctxActor: actor.FromUser(userWithSubRepoPerms),
				event: &streaming.SearchEvent{
					Results: []result.Match{
						&result.FileMatch{
							File: result.File{
								Path: "not-unauthorized.md",
							},
						},
					},
				},
			},
			wantEvent: &streaming.SearchEvent{
				Results: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "not-unauthorized.md",
						},
					},
				},
			},
		},
		{
			name: "drop match due to auth for user with sub-repo perms",
			args: args{
				ctxActor: actor.FromUser(userWithSubRepoPerms),
				event: &streaming.SearchEvent{
					Results: []result.Match{
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
			},
			wantEvent: &streaming.SearchEvent{
				Results: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "random-name.md",
						},
					},
				},
			},
		},
		{
			name: "drop match due to auth for user with sub-repo perms and error",
			args: args{
				ctxActor: actor.FromUser(userWithSubRepoPerms),
				event: &streaming.SearchEvent{
					Results: []result.Match{
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
			},
			wantEvent: &streaming.SearchEvent{
				Results: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "random-name.md",
						},
					},
				},
			},
			wantErr: "applySubRepoPerms",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), tt.args.ctxActor)
			err := applySubRepoPerms(ctx, srp, tt.args.event)
			if diff := cmp.Diff(tt.args.event, tt.wantEvent, cmpopts.IgnoreUnexported(search.RepoStatusMap{})); diff != "" {
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

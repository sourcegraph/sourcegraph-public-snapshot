package discussions

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestURLToInlineTarget(t *testing.T) {
	db.Mocks.Repos.Get = func(_ context.Context, repoID api.RepoID) (*types.Repo, error) {
		return &types.Repo{Name: "myrepo"}, nil
	}
	defer func() { db.Mocks.Repos.Get = nil }()
	ctx := context.Background()
	stringPtr := func(v string) *string { return &v }
	int32Ptr := func(v int32) *int32 { return &v }
	int64Ptr := func(v int64) *int64 { return &v }

	tests := map[string]struct {
		target              *types.DiscussionThreadTargetRepo
		threadID, commentID *int64
		wantURL             string
	}{
		"no targetpath": {
			target:  &types.DiscussionThreadTargetRepo{RepoID: 123},
			wantURL: "",
		},
		"path": {
			target:  &types.DiscussionThreadTargetRepo{RepoID: 123, Path: stringPtr("p")},
			wantURL: "/myrepo/-/blob/p#tab=discussions",
		},
		"threadID": {
			target:   &types.DiscussionThreadTargetRepo{RepoID: 123, Path: stringPtr("p")},
			threadID: int64Ptr(1),
			wantURL:  "/myrepo/-/blob/p#tab=discussions&threadID=1",
		},
		"threadID and commentID": {
			target:    &types.DiscussionThreadTargetRepo{RepoID: 123, Path: stringPtr("p")},
			threadID:  int64Ptr(1),
			commentID: int64Ptr(2),
			wantURL:   "/myrepo/-/blob/p#commentID=2&tab=discussions&threadID=1",
		},
		"start line": {
			target:  &types.DiscussionThreadTargetRepo{RepoID: 123, Path: stringPtr("p"), StartLine: int32Ptr(1)},
			wantURL: "/myrepo/-/blob/p#L2&tab=discussions",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			u, err := URLToInlineTarget(ctx, test.target, test.threadID, test.commentID)
			if err != nil {
				t.Fatal(err)
			}
			if test.wantURL == "" && u == nil {
				return
			}
			if got := u.String(); got != test.wantURL {
				t.Errorf("got URL %q, want %q", got, test.wantURL)
			}
		})
	}
}

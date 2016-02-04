package gitcmd

import (
	"reflect"
	"testing"

	"src.sourcegraph.com/sourcegraph/pkg/vcs"
)

func TestParseRemoteUpdate(t *testing.T) {
	for _, tc := range []struct {
		stderr []byte
		want   vcs.UpdateResult
	}{
		{
			stderr: []byte(""),
			want:   vcs.UpdateResult{},
		},

		{
			stderr: []byte(`From https://example.com/user/repo.git
   e8569f7..de0ad17  master     -> master
 * [new branch]      new-branch -> new-branch
`),
			want: vcs.UpdateResult{
				Changes: []vcs.Change{
					{Op: vcs.FFUpdatedOp, Branch: "master"},
					{Op: vcs.NewOp, Branch: "new-branch"},
				},
			},
		},

		{
			stderr: []byte(`From https://example.com/user/repo.git
   990cfc0..a65b539  foo-branch -> foo-branch
   d6d0813..e8569f7  master     -> master
`),
			want: vcs.UpdateResult{
				Changes: []vcs.Change{
					{Op: vcs.FFUpdatedOp, Branch: "foo-branch"},
					{Op: vcs.FFUpdatedOp, Branch: "master"},
				},
			},
		},

		{
			stderr: []byte(`From https://example.com/user/repo.git
 x [deleted]         (none)     -> master-backup-new-branch
`),
			want: vcs.UpdateResult{
				Changes: []vcs.Change{
					{Op: vcs.DeletedOp, Branch: "master-backup-new-branch"},
				},
			},
		},

		{
			stderr: []byte(`From https://example.com/user/repo.git
 * [new branch]      another-looooooong-branch-wheeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee -> another-looooooong-branch-wheeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
   de0ad17..143ee1d  master     -> master
`),
			want: vcs.UpdateResult{
				Changes: []vcs.Change{
					{Op: vcs.NewOp, Branch: "another-looooooong-branch-wheeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"},
					{Op: vcs.FFUpdatedOp, Branch: "master"},
				},
			},
		},

		{
			stderr: []byte(`From https://example.com/user/repo.git
 * [new branch]      gofmt-circleci -> gofmt-circleci
   0bccbc3..fb8ec00  master     -> master
 + ca1c467...939c2da refs/pull/291/merge -> refs/pull/291/merge  (forced update)
 + 2ca958e...2cf60d2 refs/pull/334/merge -> refs/pull/334/merge  (forced update)
 * [new ref]         refs/pull/338/head -> refs/pull/338/head
 * [new ref]         refs/pull/344/head -> refs/pull/344/head
 * [new ref]         refs/pull/344/merge -> refs/pull/344/merge
`),
			want: vcs.UpdateResult{
				Changes: []vcs.Change{
					{Op: vcs.NewOp, Branch: "gofmt-circleci"},
					{Op: vcs.FFUpdatedOp, Branch: "master"},
					{Op: vcs.ForceUpdatedOp, Branch: "refs/pull/291/merge"},
					{Op: vcs.ForceUpdatedOp, Branch: "refs/pull/334/merge"},
					{Op: vcs.NewOp, Branch: "refs/pull/338/head"},
					{Op: vcs.NewOp, Branch: "refs/pull/344/head"},
					{Op: vcs.NewOp, Branch: "refs/pull/344/merge"},
				},
			},
		},
	} {
		got, err := parseRemoteUpdate(tc.stderr)
		if err != nil {
			t.Errorf("got non-nil error: %v", err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("got %v, want %v", got, tc.want)
			continue
		}
	}
}

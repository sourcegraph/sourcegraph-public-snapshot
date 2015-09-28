package vcs_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/git"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/hg"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/hgcmd"
	"sourcegraph.com/sqs/pbtypes"
)

var times = []string{
	appleTime("2006-01-02T15:04:05Z"),
	appleTime("2014-05-06T19:20:21Z"),
}

var nonexistentCommitID = vcs.CommitID(strings.Repeat("a", 40))

func TestRepository_ResolveBranch(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
		// Some versions of Mercurial don't create .hg/cache until another command
		// is ran that uses branches. Ran into this on Mercurial 2.0.2.
		"hg branches >/dev/null",
	}
	tests := map[string]struct {
		repo interface {
			ResolveBranch(string) (vcs.CommitID, error)
		}
		branch       string
		wantCommitID vcs.CommitID
	}{
		"git libgit2": {
			repo:         makeGitRepositoryLibGit2(t, gitCommands...),
			branch:       "master",
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			branch:       "master",
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
		"hg native": {
			repo:         makeHgRepositoryNative(t, hgCommands...),
			branch:       "default",
			wantCommitID: "e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf",
		},
		"hg cmd": {
			repo:         makeHgRepositoryCmd(t, hgCommands...),
			branch:       "default",
			wantCommitID: "e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf",
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveBranch(test.branch)
		if err != nil {
			t.Errorf("%s: ResolveBranch: %s", label, err)
			continue
		}

		if commitID != test.wantCommitID {
			t.Errorf("%s: got commitID == %v, want %v", label, commitID, test.wantCommitID)
		}
	}
}

func TestRepository_ResolveBranch_error(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
	}
	tests := map[string]struct {
		repo interface {
			ResolveBranch(string) (vcs.CommitID, error)
		}
		branch  string
		wantErr error
	}{
		"git libgit2": {
			repo:    makeGitRepositoryLibGit2(t, gitCommands...),
			branch:  "doesntexist",
			wantErr: vcs.ErrBranchNotFound,
		},
		"git cmd": {
			repo:    makeGitRepositoryCmd(t, gitCommands...),
			branch:  "doesntexist",
			wantErr: vcs.ErrBranchNotFound,
		},
		"hg": {
			repo:    makeHgRepositoryNative(t, hgCommands...),
			branch:  "doesntexist",
			wantErr: vcs.ErrBranchNotFound,
		},
		"hg cmd": {
			repo:    makeHgRepositoryCmd(t, hgCommands...),
			branch:  "doesntexist",
			wantErr: vcs.ErrBranchNotFound,
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveBranch(test.branch)
		if err != test.wantErr {
			t.Errorf("%s: ResolveBranch: %s", label, err)
			continue
		}

		if commitID != "" {
			t.Errorf("%s: got commitID == %v, want empty", label, commitID)
		}
	}
}

func TestRepository_ResolveRevision(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
	}
	tests := map[string]struct {
		repo interface {
			ResolveRevision(string) (vcs.CommitID, error)
		}
		spec         string
		wantCommitID vcs.CommitID
	}{
		"git libgit2": {
			repo:         makeGitRepositoryLibGit2(t, gitCommands...),
			spec:         "master",
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			spec:         "master",
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
		"hg": {
			repo:         makeHgRepositoryNative(t, hgCommands...),
			spec:         "tip",
			wantCommitID: "e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf",
		},
		"hg cmd": {
			repo:         makeHgRepositoryCmd(t, hgCommands...),
			spec:         "tip",
			wantCommitID: "e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf",
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveRevision(test.spec)
		if err != nil {
			t.Errorf("%s: ResolveRevision: %s", label, err)
			continue
		}

		if commitID != test.wantCommitID {
			t.Errorf("%s: got commitID == %v, want %v", label, commitID, test.wantCommitID)
		}
	}
}

func TestRepository_ResolveRevision_error(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
	}
	tests := map[string]struct {
		repo interface {
			ResolveRevision(string) (vcs.CommitID, error)
		}
		spec    string
		wantErr error
	}{
		"git libgit2 testcase1": {
			repo:    makeGitRepositoryLibGit2(t, gitCommands...),
			spec:    "doesntexist",
			wantErr: vcs.ErrRevisionNotFound,
		},
		"git cmd testcase1": {
			repo:    makeGitRepositoryCmd(t, gitCommands...),
			spec:    "doesntexist",
			wantErr: vcs.ErrRevisionNotFound,
		},
		"hg testcase1": {
			repo:    makeHgRepositoryNative(t, hgCommands...),
			spec:    "doesntexist",
			wantErr: vcs.ErrRevisionNotFound,
		},
		"hg cmd testcase1": {
			repo:    makeHgRepositoryCmd(t, hgCommands...),
			spec:    "doesntexist",
			wantErr: vcs.ErrRevisionNotFound,
		},

		// These revisions look like valid commit hashes (and may be valid after more commits are made),
		// but they are not present in the current repository, hence we want vcs.ErrRevisionNotFound.
		"git libgit2 testcase2": {
			repo:    makeGitRepositoryLibGit2(t, gitCommands...),
			spec:    "2874b2ef9be165966e5620fc29b592c041262721",
			wantErr: vcs.ErrRevisionNotFound,
		},
		"git cmd testcase2": {
			repo:    makeGitRepositoryCmd(t, gitCommands...),
			spec:    "2874b2ef9be165966e5620fc29b592c041262721",
			wantErr: vcs.ErrRevisionNotFound,
		},
		"hg testcase2": {
			repo:    makeHgRepositoryNative(t, hgCommands...),
			spec:    "2874b2ef9be165966e5620fc29b592c041262721",
			wantErr: vcs.ErrRevisionNotFound,
		},
		"hg cmd testcase2": {
			repo:    makeHgRepositoryCmd(t, hgCommands...),
			spec:    "2874b2ef9be165966e5620fc29b592c041262721",
			wantErr: vcs.ErrRevisionNotFound,
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveRevision(test.spec)
		if err != test.wantErr {
			t.Errorf("%s: ResolveRevision: got %v, want %v", label, err, test.wantErr)
			continue
		}

		if commitID != "" {
			t.Errorf("%s: got commitID == %v, want empty", label, commitID)
		}
	}
}

func TestRepository_ResolveTag(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag t",
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
		"hg tag t",
	}
	tests := map[string]struct {
		repo interface {
			ResolveTag(string) (vcs.CommitID, error)
		}
		tag          string
		wantCommitID vcs.CommitID
	}{
		"git libgit2": {
			repo:         makeGitRepositoryLibGit2(t, gitCommands...),
			tag:          "t",
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			tag:          "t",
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
		"hg": {
			repo:         makeHgRepositoryNative(t, hgCommands...),
			tag:          "t",
			wantCommitID: "e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf",
		},
		"hg cmd": {
			repo:         makeHgRepositoryCmd(t, hgCommands...),
			tag:          "t",
			wantCommitID: "e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf",
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveTag(test.tag)
		if err != nil {
			t.Errorf("%s: ResolveTag: %s", label, err)
			continue
		}

		if commitID != test.wantCommitID {
			t.Errorf("%s: got commitID == %v, want %v", label, commitID, test.wantCommitID)
		}
	}
}

func TestRepository_ResolveTag_error(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
	}
	tests := map[string]struct {
		repo interface {
			ResolveTag(string) (vcs.CommitID, error)
		}
		tag     string
		wantErr error
	}{
		"git libgit2": {
			repo:    makeGitRepositoryLibGit2(t, gitCommands...),
			tag:     "doesntexist",
			wantErr: vcs.ErrTagNotFound,
		},
		"git cmd": {
			repo:    makeGitRepositoryCmd(t, gitCommands...),
			tag:     "doesntexist",
			wantErr: vcs.ErrTagNotFound,
		},
		"hg": {
			repo:    makeHgRepositoryNative(t, hgCommands...),
			tag:     "doesntexist",
			wantErr: vcs.ErrTagNotFound,
		},
		"hg cmd": {
			repo:    makeHgRepositoryCmd(t, hgCommands...),
			tag:     "doesntexist",
			wantErr: vcs.ErrTagNotFound,
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveTag(test.tag)
		if err != test.wantErr {
			t.Errorf("%s: ResolveTag: %s", label, err)
			continue
		}

		if commitID != "" {
			t.Errorf("%s: got commitID == %v, want empty", label, commitID)
		}
	}
}

func TestRepository_Branches(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git checkout -b b0",
		"git checkout -b b1",
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg branch b0",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
		"touch --date=2014-05-06T19:20:21Z g || touch -t " + times[1] + " g",
		"hg branch b1",
		"hg add g",
		"hg commit -m foo --date '2006-12-09 15:19:44 UTC' --user 'a <a@a.com>'",
	}
	tests := map[string]struct {
		repo interface {
			Branches(vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		wantBranches []*vcs.Branch
	}{
		"git libgit2": {
			repo:         makeGitRepositoryLibGit2(t, gitCommands...),
			wantBranches: []*vcs.Branch{{Name: "b0", Head: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}, {Name: "b1", Head: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}, {Name: "master", Head: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}},
		},
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			wantBranches: []*vcs.Branch{{Name: "b0", Head: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}, {Name: "b1", Head: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}, {Name: "master", Head: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}},
		},
		"hg": {
			repo:         makeHgRepositoryNative(t, hgCommands...),
			wantBranches: []*vcs.Branch{{Name: "b0", Head: "4edb70f7b9dd1ce8e95242525377098f477a89c3"}, {Name: "b1", Head: "843c6421bd707b885cc3849b8eb0b5b2b9298e8b"}},
		},
		"hg cmd": {
			repo:         makeHgRepositoryCmd(t, hgCommands...),
			wantBranches: []*vcs.Branch{{Name: "b0", Head: "4edb70f7b9dd1ce8e95242525377098f477a89c3"}, {Name: "b1", Head: "843c6421bd707b885cc3849b8eb0b5b2b9298e8b"}},
		},
	}

	for label, test := range tests {
		branches, err := test.repo.Branches(vcs.BranchesOptions{})
		if err != nil {
			t.Errorf("%s: Branches: %s", label, err)
			continue
		}

		if !reflect.DeepEqual(branches, test.wantBranches) {
			t.Errorf("%s: got branches == %v, want %v", label, asJSON(branches), asJSON(test.wantBranches))
		}
	}
}

func TestRepository_Branches_MergedInto(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"git checkout -b b0",
		"echo 123 > some_other_file",
		"git add some_other_file",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -am foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -am foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",

		"git checkout HEAD^ -b b1",
		"git merge b0",
	}

	for label, test := range map[string]struct {
		repo interface {
			Branches(vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		wantBranches map[string][]*vcs.Branch
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			wantBranches: map[string][]*vcs.Branch{
				"b1": []*vcs.Branch{
					{Name: "b0", Head: "6520a4539a4cb664537c712216a53d80dd79bbdc"},
					{Name: "b1", Head: "6520a4539a4cb664537c712216a53d80dd79bbdc"},
				},
			},
		},
	} {
		for branch, mergedInto := range test.wantBranches {
			branches, err := test.repo.Branches(vcs.BranchesOptions{MergedInto: branch})
			if err != nil {
				t.Errorf("%s: Branches: %s", label, err)
				continue
			}
			if !reflect.DeepEqual(branches, mergedInto) {
				t.Errorf("%s: got branches == %v, want %v", label, asJSON(branches), asJSON(test.wantBranches))
			}
		}
	}
}

func TestRepository_Branches_ContainsCommit(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m base --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m master --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git checkout HEAD^ -b branch2",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m branch2 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}

	tests := map[string]struct {
		repo interface {
			Branches(vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		commitToWantBranches map[string][]*vcs.Branch
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			commitToWantBranches: map[string][]*vcs.Branch{
				"920c0e9d7b287b030ac9770fd7ba3ee9dc1760d9": []*vcs.Branch{{Name: "branch2", Head: "920c0e9d7b287b030ac9770fd7ba3ee9dc1760d9"}},
				"1224d334dfe08f4693968ea618ad63ae86ec16ca": []*vcs.Branch{{Name: "master", Head: "1224d334dfe08f4693968ea618ad63ae86ec16ca"}},
				"2816a72df28f699722156e545d038a5203b959de": []*vcs.Branch{{Name: "master", Head: "1224d334dfe08f4693968ea618ad63ae86ec16ca"}, {Name: "branch2", Head: "920c0e9d7b287b030ac9770fd7ba3ee9dc1760d9"}},
			},
		},
	}

	for label, test := range tests {
		for commit, wantBranches := range test.commitToWantBranches {
			branches, err := test.repo.Branches(vcs.BranchesOptions{ContainsCommit: commit})
			if err != nil {
				t.Errorf("%s: Branches: %s", label, err)
				continue
			}

			if !reflect.DeepEqual(branches, wantBranches) {
				t.Errorf("%s: got branches == %v, want %v", label, asJSON(branches), asJSON(wantBranches))
			}
		}
	}
}

func TestRepository_Branches_BehindAheadCounts(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo0 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git branch old_work",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo2 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo3 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo4 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo5 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git checkout -b dev",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo6 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo7 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo8 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git checkout old_work",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo9 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo interface {
			Branches(vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		wantBranches []*vcs.Branch
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			wantBranches: []*vcs.Branch{
				{Counts: &vcs.BehindAhead{Behind: 5, Ahead: 1}, Name: "old_work", Head: "26692c614c59ddaef4b57926810aac7d5f0e94f0"},
				{Counts: &vcs.BehindAhead{Behind: 0, Ahead: 3}, Name: "dev", Head: "6724953367f0cd9a7755bac46ee57f4ab0c1aad8"},
				{Counts: &vcs.BehindAhead{Behind: 0, Ahead: 0}, Name: "master", Head: "8ea26e077a8fb9aa502c3fe2cfa3ce4e052d1a76"},
			},
		},
	}

	for label, test := range tests {
		branches, err := test.repo.Branches(vcs.BranchesOptions{BehindAheadBranch: "master"})
		if err != nil {
			t.Errorf("%s: Branches: %s", label, err)
			continue
		}

		if !reflect.DeepEqual(branches, test.wantBranches) {
			t.Errorf("%s: got branches == %v, want %v", label, asJSON(branches), asJSON(test.wantBranches))
		}
	}
}

func TestRepository_Branches_IncludeCommit(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo0 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git checkout -b b0",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit --allow-empty -m foo1 --author='b <b@b.com>' --date 2006-01-02T15:04:06Z",
	}
	tests := map[string]struct {
		repo interface {
			Branches(vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		wantBranches []*vcs.Branch
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			wantBranches: []*vcs.Branch{
				{
					Name: "master", Head: "a3c1537db9797215208eec56f8e7c9c37f8358ca",
					Commit: &vcs.Commit{
						ID:        "a3c1537db9797215208eec56f8e7c9c37f8358ca",
						Author:    vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
						Committer: &vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
						Message:   "foo0",
						Parents:   nil,
					},
				},
				{
					Name: "b0", Head: "c4a53701494d1d788b1ceeb8bf32e90224962473",
					Commit: &vcs.Commit{
						ID:        "c4a53701494d1d788b1ceeb8bf32e90224962473",
						Author:    vcs.Signature{"b", "b@b.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
						Committer: &vcs.Signature{"b", "b@b.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
						Message:   "foo1",
						Parents:   []vcs.CommitID{"a3c1537db9797215208eec56f8e7c9c37f8358ca"},
					},
				},
			},
		},
	}

	for label, test := range tests {
		branches, err := test.repo.Branches(vcs.BranchesOptions{IncludeCommit: true})
		if err != nil {
			t.Errorf("%s: Branches: %s", label, err)
			continue
		}

		if !reflect.DeepEqual(branches, test.wantBranches) {
			t.Errorf("%s: got branches == %v, want %v", label, asJSON(branches), asJSON(test.wantBranches))
		}
	}
}

func TestRepository_Tags(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag t0",
		"git tag t1",
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
		"hg tag t0 --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
		"hg tag t1 --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
	}
	tests := map[string]struct {
		repo interface {
			Tags() ([]*vcs.Tag, error)
		}
		wantTags []*vcs.Tag
	}{
		"git libgit2": {
			repo:     makeGitRepositoryLibGit2(t, gitCommands...),
			wantTags: []*vcs.Tag{{Name: "t0", CommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}, {Name: "t1", CommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}},
		},
		"git cmd": {
			repo:     makeGitRepositoryCmd(t, gitCommands...),
			wantTags: []*vcs.Tag{{Name: "t0", CommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}, {Name: "t1", CommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}},
		},
		"hg": {
			repo:     makeHgRepositoryNative(t, hgCommands...),
			wantTags: []*vcs.Tag{{Name: "t0", CommitID: "e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf"}, {Name: "t1", CommitID: "6a6ae0da9d7c3bf48de61e5584d6eb5dcba0750c"}, {Name: "tip", CommitID: "217f213c2dbe4ce6573ec0b0dbd3e7abafaf8fba"}},
		},
		"hg cmd": {
			repo:     makeHgRepositoryCmd(t, hgCommands...),
			wantTags: []*vcs.Tag{{Name: "t0", CommitID: "e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf"}, {Name: "t1", CommitID: "6a6ae0da9d7c3bf48de61e5584d6eb5dcba0750c"}, {Name: "tip", CommitID: "217f213c2dbe4ce6573ec0b0dbd3e7abafaf8fba"}},
		},
	}

	for label, test := range tests {
		tags, err := test.repo.Tags()
		if err != nil {
			t.Errorf("%s: Tags: %s", label, err)
			continue
		}

		if !reflect.DeepEqual(tags, test.wantTags) {
			t.Errorf("%s: got tags == %v, want %v", label, tags, test.wantTags)
		}
	}
}

func TestRepository_GetCommit(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	wantGitCommit := &vcs.Commit{
		ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
		Author:    vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
		Committer: &vcs.Signature{"c", "c@c.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
		Message:   "bar",
		Parents:   []vcs.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
		"touch --date=2006-01-02T15:04:05Z g || touch -t " + times[0] + " g",
		"hg add g",
		"hg commit -m bar --date '2006-12-06 13:18:30 UTC' --user 'a <a@a.com>'",
	}
	wantHgCommit := &vcs.Commit{
		ID:      "c6320cdba5ebc6933bd7c94751dcd633d6aa0759",
		Author:  vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-12-06T13:18:30Z")},
		Message: "bar",
		Parents: []vcs.CommitID{"e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf"},
	}
	tests := map[string]struct {
		repo interface {
			GetCommit(vcs.CommitID) (*vcs.Commit, error)
		}
		id         vcs.CommitID
		wantCommit *vcs.Commit
	}{
		"git libgit2": {
			repo:       makeGitRepositoryLibGit2(t, gitCommands...),
			id:         "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommit: wantGitCommit,
		},
		"git cmd": {
			repo:       makeGitRepositoryCmd(t, gitCommands...),
			id:         "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommit: wantGitCommit,
		},
		"hg": {
			repo:       makeHgRepositoryNative(t, hgCommands...),
			id:         "c6320cdba5ebc6933bd7c94751dcd633d6aa0759",
			wantCommit: wantHgCommit,
		},
		"hg cmd": {
			repo:       makeHgRepositoryCmd(t, hgCommands...),
			id:         "c6320cdba5ebc6933bd7c94751dcd633d6aa0759",
			wantCommit: wantHgCommit,
		},
	}

	for label, test := range tests {
		commit, err := test.repo.GetCommit(test.id)
		if err != nil {
			t.Errorf("%s: GetCommit: %s", label, err)
			continue
		}

		if !commitsEqual(commit, test.wantCommit) {
			t.Errorf("%s: got commit == %+v, want %+v", label, commit, test.wantCommit)
		}

		// Test that trying to get a nonexistent commit returns ErrCommitNotFound.
		if _, err := test.repo.GetCommit(nonexistentCommitID); err != vcs.ErrCommitNotFound {
			t.Errorf("%s: for nonexistent commit: got err %v, want %v", label, err, vcs.ErrCommitNotFound)
		}
	}
}

func TestRepository_Commits(t *testing.T) {
	t.Parallel()

	// TODO(sqs): test CommitsOptions.Base

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	wantGitCommits := []*vcs.Commit{
		{
			ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			Author:    vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &vcs.Signature{"c", "c@c.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Message:   "bar",
			Parents:   []vcs.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
		},
		{
			ID:        "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
			Author:    vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "foo",
			Parents:   nil,
		},
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
		"touch --date=2006-01-02T15:04:05Z g || touch -t " + times[0] + " g",
		"hg add g",
		"hg commit -m bar --date '2006-12-06 13:18:30 UTC' --user 'a <a@a.com>'",
	}
	wantHgCommits := []*vcs.Commit{
		{
			ID:      "c6320cdba5ebc6933bd7c94751dcd633d6aa0759",
			Author:  vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-12-06T13:18:30Z")},
			Message: "bar",
			Parents: []vcs.CommitID{"e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf"},
		},
		{
			ID:      "e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf",
			Author:  vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-12-06T13:18:29Z")},
			Message: "foo",
			Parents: nil,
		},
	}
	tests := map[string]struct {
		repo interface {
			Commits(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error)
		}
		id          vcs.CommitID
		wantCommits []*vcs.Commit
		wantTotal   uint
	}{
		"git libgit2": {
			repo:        makeGitRepositoryLibGit2(t, gitCommands...),
			id:          "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommits: wantGitCommits,
			wantTotal:   2,
		},
		"git cmd": {
			repo:        makeGitRepositoryCmd(t, gitCommands...),
			id:          "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommits: wantGitCommits,
			wantTotal:   2,
		},
		"hg native": {
			repo:        makeHgRepositoryNative(t, hgCommands...),
			id:          "c6320cdba5ebc6933bd7c94751dcd633d6aa0759",
			wantCommits: wantHgCommits,
			wantTotal:   2,
		},
		"hg cmd": {
			repo:        makeHgRepositoryCmd(t, hgCommands...),
			id:          "c6320cdba5ebc6933bd7c94751dcd633d6aa0759",
			wantCommits: wantHgCommits,
			wantTotal:   2,
		},
	}

	for label, test := range tests {
		commits, total, err := test.repo.Commits(vcs.CommitsOptions{Head: test.id})
		if err != nil {
			t.Errorf("%s: Commits: %s", label, err)
			continue
		}

		if total != test.wantTotal {
			t.Errorf("%s: got %d total commits, want %d", label, total, test.wantTotal)
		}

		if len(commits) != len(test.wantCommits) {
			t.Errorf("%s: got %d commits, want %d", label, len(commits), len(test.wantCommits))
		}

		for i := 0; i < len(commits) || i < len(test.wantCommits); i++ {
			var gotC, wantC *vcs.Commit
			if i < len(commits) {
				gotC = commits[i]
			}
			if i < len(test.wantCommits) {
				wantC = test.wantCommits[i]
			}
			if !commitsEqual(gotC, wantC) {
				t.Errorf("%s: got commit %d == %+v, want %+v", label, i, gotC, wantC)
			}
		}

		// Test that trying to get a nonexistent commit returns ErrCommitNotFound.
		if _, _, err := test.repo.Commits(vcs.CommitsOptions{Head: nonexistentCommitID}); err != vcs.ErrCommitNotFound {
			t.Errorf("%s: for nonexistent commit: got err %v, want %v", label, err, vcs.ErrCommitNotFound)
		}
	}
}

func TestRepository_Commits_options(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:08Z git commit --allow-empty -m qux --author='a <a@a.com>' --date 2006-01-02T15:04:08Z",
	}
	wantGitCommits := []*vcs.Commit{
		{
			ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			Author:    vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &vcs.Signature{"c", "c@c.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Message:   "bar",
			Parents:   []vcs.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
		},
	}
	wantGitCommits2 := []*vcs.Commit{
		{
			ID:        "ade564eba4cf904492fb56dcd287ac633e6e082c",
			Author:    vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Committer: &vcs.Signature{"c", "c@c.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Message:   "qux",
			Parents:   []vcs.CommitID{"b266c7e3ca00b1a17ad0b1449825d0854225c007"},
		},
	}
	hgCommands := []string{
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
		"touch --date=2006-01-02T15:04:05Z g || touch -t " + times[0] + " g",
		"hg add g",
		"hg commit -m bar --date '2006-12-06 13:18:30 UTC' --user 'a <a@a.com>'",
		"touch --date=2006-01-02T15:04:05Z h || touch -t " + times[0] + " h",
		"hg add h",
		"hg commit -m qux --date '2006-12-06 13:18:30 UTC' --user 'a <a@a.com>'",
	}
	wantHgCommits := []*vcs.Commit{
		{
			ID:      "c6320cdba5ebc6933bd7c94751dcd633d6aa0759",
			Author:  vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-12-06T13:18:30Z")},
			Message: "bar",
			Parents: []vcs.CommitID{"e8e11ff1be92a7be71b9b5cdb4cc674b7dc9facf"},
		},
	}
	tests := map[string]struct {
		repo interface {
			Commits(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error)
		}
		opt         vcs.CommitsOptions
		wantCommits []*vcs.Commit
		wantTotal   uint
	}{
		"git libgit2": {
			repo:        makeGitRepositoryLibGit2(t, gitCommands...),
			opt:         vcs.CommitsOptions{Head: "ade564eba4cf904492fb56dcd287ac633e6e082c", N: 1, Skip: 1},
			wantCommits: wantGitCommits,
			wantTotal:   3,
		},
		"git cmd": {
			repo:        makeGitRepositoryCmd(t, gitCommands...),
			opt:         vcs.CommitsOptions{Head: "ade564eba4cf904492fb56dcd287ac633e6e082c", N: 1, Skip: 1},
			wantCommits: wantGitCommits,
			wantTotal:   3,
		},
		"git libgit2 Head": {
			repo: makeGitRepositoryLibGit2(t, gitCommands...),
			opt: vcs.CommitsOptions{
				Head: "ade564eba4cf904492fb56dcd287ac633e6e082c",
				Base: "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			},
			wantCommits: wantGitCommits2,
			wantTotal:   1,
		},
		"git cmd Head": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			opt: vcs.CommitsOptions{
				Head: "ade564eba4cf904492fb56dcd287ac633e6e082c",
				Base: "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			},
			wantCommits: wantGitCommits2,
			wantTotal:   1,
		},
		"hg native": {
			repo:        makeHgRepositoryNative(t, hgCommands...),
			opt:         vcs.CommitsOptions{Head: "443def46748a0c02c312bb4fdc6231d6ede45f49", N: 1, Skip: 1},
			wantCommits: wantHgCommits,
			wantTotal:   3,
		},
		"hg cmd": {
			repo:        makeHgRepositoryCmd(t, hgCommands...),
			opt:         vcs.CommitsOptions{Head: "443def46748a0c02c312bb4fdc6231d6ede45f49", N: 1, Skip: 1},
			wantCommits: wantHgCommits,
			wantTotal:   3,
		},
	}

	for label, test := range tests {
		commits, total, err := test.repo.Commits(test.opt)
		if err != nil {
			t.Errorf("%s: Commits(): %s", label, err)
			continue
		}

		if total != test.wantTotal {
			t.Errorf("%s: got %d total commits, want %d", label, total, test.wantTotal)
		}

		if len(commits) != len(test.wantCommits) {
			t.Errorf("%s: got %d commits, want %d", label, len(commits), len(test.wantCommits))
		}

		for i := 0; i < len(commits) || i < len(test.wantCommits); i++ {
			var gotC, wantC *vcs.Commit
			if i < len(commits) {
				gotC = commits[i]
			}
			if i < len(test.wantCommits) {
				wantC = test.wantCommits[i]
			}
			if !commitsEqual(gotC, wantC) {
				t.Errorf("%s: got commit %d == %+v, want %+v", label, i, gotC, wantC)
			}
		}
	}
}

func TestRepository_Commits_options_path(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"touch file1",
		"touch --date=2006-01-02T15:04:05Z file1 || touch -t " + times[0] + " file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m commit3 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	wantGitCommits := []*vcs.Commit{
		{
			ID:        "546a3ef26e581624ef997cb8c0ba01ee475fc1dc",
			Author:    vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &vcs.Signature{"a", "a@a.com", mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "commit2",
			Parents:   []vcs.CommitID{"a04652fa1998a0a7d2f2f77ecb7021de943d3aab"},
		},
	}
	tests := map[string]struct {
		repo interface {
			Commits(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error)
		}
		opt         vcs.CommitsOptions
		wantCommits []*vcs.Commit
		wantTotal   uint
	}{
		"git cmd Path 0": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			opt: vcs.CommitsOptions{
				Head: "master",
				Path: "doesnt-exist",
			},
			wantCommits: nil,
			wantTotal:   0,
		},
		"git cmd Path 1": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			opt: vcs.CommitsOptions{
				Head: "master",
				Path: "file1",
			},
			wantCommits: wantGitCommits,
			wantTotal:   1,
		},
	}

	for label, test := range tests {
		commits, total, err := test.repo.Commits(test.opt)
		if err != nil {
			t.Errorf("%s: Commits(): %s", label, err)
			continue
		}

		if total != test.wantTotal {
			t.Errorf("%s: got %d total commits, want %d", label, total, test.wantTotal)
		}

		if len(commits) != len(test.wantCommits) {
			t.Errorf("%s: got %d commits, want %d", label, len(commits), len(test.wantCommits))
		}

		for i := 0; i < len(commits) || i < len(test.wantCommits); i++ {
			var gotC, wantC *vcs.Commit
			if i < len(commits) {
				gotC = commits[i]
			}
			if i < len(test.wantCommits) {
				wantC = test.wantCommits[i]
			}
			if !commitsEqual(gotC, wantC) {
				t.Errorf("%s: got commit %d == %+v, want %+v", label, i, gotC, wantC)
			}
		}
	}
}

func TestRepository_FileSystem_Symlinks(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"touch file1",
		"ln -s file1 link1",
		"touch --date=2006-01-02T15:04:05Z file1 link1 || touch -t " + times[0] + " file1 link1",
		"git add link1 file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	hgCommands := []string{
		"touch file1",
		"ln -s file1 link1",
		"touch --date=2006-01-02T15:04:05Z file1 link1 || touch -t " + times[0] + " file1 link1",
		"hg add link1 file1",
		"hg commit -m commit1 --user 'a <a@a.com>' --date '2006-01-02 15:04:05 UTC'",
	}

	tests := map[string]struct {
		repo interface {
			FileSystem(vcs.CommitID) (vfs.FileSystem, error)
		}
		commitID vcs.CommitID

		testFileInfoSys bool // whether to check the SymlinkInfo in FileInfo.Sys()
	}{
		// TODO(sqs): implement Lstat and symlink handling for git libgit2, git
		// cmd, and hg cmd.

		"git libgit2": {
			repo:     makeGitRepositoryLibGit2(t, gitCommands...),
			commitID: "85d3a39020cf28af4b887552fcab9e31a49f2ced",

			testFileInfoSys: true,
		},
		"git cmd": {
			repo:     makeGitRepositoryCmd(t, gitCommands...),
			commitID: "85d3a39020cf28af4b887552fcab9e31a49f2ced",

			testFileInfoSys: true,
		},
		"hg native": {
			repo:     makeHgRepositoryNative(t, hgCommands...),
			commitID: "c3fed02bbbc0b58418f32a363b8263aa46b0349e",
			// TODO(sqs): implement SymlinkInfo
		},
		// "hg cmd": {
		// 	repo:     &HgRepositoryCmd{initHgRepository(t, hgCommands...)},
		// 	commitID: "c3fed02bbbc0b58418f32a363b8263aa46b0349e",
		// },
	}
	for label, test := range tests {
		fs, err := test.repo.FileSystem(test.commitID)
		if err != nil {
			t.Errorf("%s: FileSystem: %s", label, err)
			continue
		}

		// file1 should be a file.
		file1Info, err := fs.Stat("file1")
		if err != nil {
			t.Errorf("%s: fs.Stat(file1): %s", label, err)
			continue
		}
		if !file1Info.Mode().IsRegular() {
			t.Errorf("%s: file1 Stat !IsRegular (mode: %o)", label, file1Info.Mode())
		}

		checkSymlinkFileInfo := func(label string, link os.FileInfo) {
			if link.Mode()&os.ModeSymlink == 0 {
				t.Errorf("%s: link mode is not symlink (mode: %o)", label, link.Mode())
			}
			if want := "link1"; link.Name() != want {
				t.Errorf("%s: got link.Name() == %q, want %q", label, link.Name(), want)
			}
			if test.testFileInfoSys {
				si, ok := link.Sys().(vcs.SymlinkInfo)
				if !ok {
					t.Errorf("%s: link.Sys(): got %v %T, want SymlinkInfo", label, si, si)
				}
				if want := "file1"; si.Dest != want {
					t.Errorf("%s: (SymlinkInfo).Dest: got %q, want %q", label, si.Dest, want)
				}
			}
		}

		// link1 should be a link.
		link1Linfo, err := fs.Lstat("link1")
		if err != nil {
			t.Errorf("%s: fs.Lstat(link1): %s", label, err)
			continue
		}
		checkSymlinkFileInfo(label+" (Lstat)", link1Linfo)

		// Also check the FileInfo returned by ReadDir to ensure it's
		// consistent with the FileInfo returned by Lstat.
		entries, err := fs.ReadDir(".")
		if err != nil {
			t.Errorf("%s: fs.ReadDir(.): %s", label, err)
			continue
		}
		if got, want := len(entries), 2; got != want {
			t.Errorf("%s: got len(entries) == %d, want %d", label, got, want)
			continue
		}
		checkSymlinkFileInfo(label+" (ReadDir)", entries[1])

		// link1 stat should follow the link to file1.
		link1Info, err := fs.Stat("link1")
		if err != nil {
			t.Errorf("%s: fs.Stat(link1): %s", label, err)
			continue
		}
		if !link1Info.Mode().IsRegular() {
			t.Errorf("%s: link1 Stat !IsRegular (mode: %o)", label, link1Info.Mode())
		}
		if link1Info.Name() != "link1" {
			t.Errorf("%s: got link1 Name %q, want %q", label, link1Info.Name(), "link1")
		}
	}
}

func TestRepository_FileSystem(t *testing.T) {
	t.Parallel()

	file1MTime, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	if err != nil {
		t.Fatal(err)
	}

	// In all tests, repo should contain two commits. The first commit (whose ID
	// is in the 'first' field) has a file at dir1/file1 with the contents
	// "myfile1" and the mtime 2006-01-02T15:04:05Z. The second commit (whose ID
	// is in the 'second' field) adds a file at file2 (in the top-level
	// directory of the repository) with the contents "infile2" and the mtime
	// 2014-05-06T19:20:21Z.
	//
	// TODO(sqs): add symlinks, etc.
	gitCommands := []string{
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --date=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t " + times[0] + " dir1 dir1/file1",
		"git add dir1/file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo -n infile2 > file2",
		"touch --date=2014-05-06T19:20:21Z file2 || touch -t " + times[1] + " file2",
		"git add file2",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --author='a <a@a.com>' --date 2014-05-06T19:20:21Z",
	}
	hgCommands := []string{
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --date=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t " + times[0] + " dir1 dir1/file1",
		"hg add dir1/file1",
		"hg commit -m commit1 --user 'a <a@a.com>' --date '2006-01-02 15:04:05 UTC'",
		"echo -n infile2 > file2",
		"touch --date=2014-05-06T19:20:21Z file2 || touch -t " + times[1] + " file2",
		"hg add file2",
		"hg commit -m commit2 --user 'a <a@a.com>' --date '2014-05-06 19:20:21 UTC'",
	}
	tests := map[string]struct {
		repo interface {
			FileSystem(vcs.CommitID) (vfs.FileSystem, error)
		}
		first, second vcs.CommitID
	}{
		"git libgit2": {
			repo:   makeGitRepositoryLibGit2(t, gitCommands...),
			first:  "b6602ca96bdc0ab647278577a3c6edcb8fe18fb0",
			second: "ace35f1597e087fe2d302ed6cb2763174e6b9660",
		},
		"git cmd": {
			repo:   makeGitRepositoryCmd(t, gitCommands...),
			first:  "b6602ca96bdc0ab647278577a3c6edcb8fe18fb0",
			second: "ace35f1597e087fe2d302ed6cb2763174e6b9660",
		},
		"hg native": {
			repo:   makeHgRepositoryNative(t, hgCommands...),
			first:  "0b3260387c55ff0834b520fd7f5d4f4a15c22827",
			second: "810c55b76823441dabb1249837e7ebceab50ce1a",
		},
		"hg cmd": {
			repo:   makeHgRepositoryCmd(t, hgCommands...),
			first:  "0b3260387c55ff0834b520fd7f5d4f4a15c22827",
			second: "810c55b76823441dabb1249837e7ebceab50ce1a",
		},
	}

	for label, test := range tests {
		fs1, err := test.repo.FileSystem(test.first)
		if err != nil {
			t.Errorf("%s: FileSystem: %s", label, err)
			continue
		}

		// notafile should not exist.
		if _, err = fs1.Stat("notafile"); !os.IsNotExist(err) {
			t.Errorf("%s: fs1.Stat(notafile): got err %v, want os.IsNotExist", label, err)
			continue
		}

		// dir1 should exist and be a dir.
		dir1Info, err := fs1.Stat("dir1")
		if err != nil {
			t.Errorf("%s: fs1.Stat(dir1): %s", label, err)
			continue
		}
		if !dir1Info.Mode().IsDir() {
			t.Errorf("%s: dir1 stat !IsDir", label)
		}
		if name := dir1Info.Name(); name != "dir1" {
			t.Errorf("%s: got dir1 name %q, want 'dir1'", label, name)
		}

		// dir1 should contain one entry: file1.
		dir1Entries, err := fs1.ReadDir("dir1")
		if err != nil {
			t.Errorf("%s: fs1.ReadDir(dir1): %s", label, err)
			continue
		}
		if len(dir1Entries) != 1 {
			t.Errorf("%s: got %d dir1 entries, want 1", label, len(dir1Entries))
			continue
		}
		if file1Info := dir1Entries[0]; file1Info.Name() != "file1" {
			t.Errorf("%s: got dir1 entry name == %q, want 'file1'", label, file1Info.Name())
		}

		// dir1/file1 should exist, contain "infile1", have the right mtime, and be a file.
		file1, err := fs1.Open("dir1/file1")
		if err != nil {
			t.Errorf("%s: fs1.Open(dir1/file1): %s", label, err)
			continue
		}
		file1Data, err := ioutil.ReadAll(file1)
		if err != nil {
			t.Errorf("%s: ReadAll(file1): %s", label, err)
			continue
		}
		if !bytes.Equal(file1Data, []byte("infile1")) {
			t.Errorf("%s: got file1Data == %q, want %q", label, string(file1Data), "infile1")
		}
		file1Info, err := fs1.Stat("dir1/file1")
		if err != nil {
			t.Errorf("%s: fs1.Stat(dir1/file1): %s", label, err)
			continue
		}
		if !file1Info.Mode().IsRegular() {
			t.Errorf("%s: file1 stat !IsRegular", label)
		}
		if name := file1Info.Name(); name != "file1" {
			t.Errorf("%s: got file1 name %q, want 'file1'", label, name)
		}
		if size, want := file1Info.Size(), int64(len("infile1")); size != want {
			t.Errorf("%s: got file1 size %d, want %d", label, size, want)
		}
		if mtime, want := file1Info.ModTime(), file1MTime; !mtime.Equal(want) {
			t.Errorf("%s: got file1 mtime %v, want %v", label, mtime, want)
		}

		// file2 shouldn't exist in the 1st commit.
		_, err = fs1.Open("file2")
		if !os.IsNotExist(err) {
			t.Errorf("%s: fs1.Open(file2): got err %v, want os.IsNotExist (file2 should not exist in this commit)", label, err)
		}

		// file2 should exist in the 2nd commit.
		fs2, err := test.repo.FileSystem(test.second)
		if err != nil {
			t.Errorf("%s: FileSystem: %s", label, err)
		}
		_, err = fs2.Open("file2")
		if err != nil {
			t.Errorf("%s: fs2.Open(file2): %s", label, err)
			continue
		}

		// file1 should also exist in the 2nd commit.
		file1Info, err = fs2.Stat("dir1/file1")
		if err != nil {
			t.Errorf("%s: fs2.Stat(dir1/file1): %s", label, err)
			continue
		}
		file1, err = fs2.Open("dir1/file1")
		if err != nil {
			t.Errorf("%s: fs2.Open(dir1/file1): %s", label, err)
			continue
		}

		// root should exist (via Stat).
		root, err := fs2.Stat(".")
		if err != nil {
			t.Errorf("%s: fs2.Stat(.): %s", label, err)
			continue
		}
		if !root.Mode().IsDir() {
			t.Errorf("%s: got root !IsDir", label)
		}

		// root should have 2 entries: dir1 and file2.
		rootEntries, err := fs2.ReadDir(".")
		if err != nil {
			t.Errorf("%s: fs2.ReadDir(.): %s", label, err)
			continue
		}
		if got, want := len(rootEntries), 2; got != want {
			t.Errorf("%s: got len(rootEntries) == %d, want %d", label, got, want)
			continue
		}
		if e0 := rootEntries[0]; !(e0.Name() == "dir1" && e0.Mode().IsDir()) {
			t.Errorf("%s: got root entry 0 %q IsDir=%v, want 'dir1' IsDir=true", label, e0.Name(), e0.Mode().IsDir())
		}
		if e1 := rootEntries[1]; !(e1.Name() == "file2" && !e1.Mode().IsDir()) {
			t.Errorf("%s: got root entry 1 %q IsDir=%v, want 'file2' IsDir=false", label, e1.Name(), e1.Mode().IsDir())
		}

		// dir1 should still only contain one entry: file1.
		dir1Entries, err = fs2.ReadDir("dir1")
		if err != nil {
			t.Errorf("%s: fs1.ReadDir(dir1): %s", label, err)
			continue
		}
		if len(dir1Entries) != 1 {
			t.Errorf("%s: got %d dir1 entries, want 1", label, len(dir1Entries))
			continue
		}
		if file1Info := dir1Entries[0]; file1Info.Name() != "file1" {
			t.Errorf("%s: got dir1 entry name == %q, want 'file1'", label, file1Info.Name())
		}
	}
}

func TestRepository_FileSystem_gitSubmodules(t *testing.T) {
	t.Parallel()

	submodDir := initGitRepository(t,
		"touch f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	)
	const submodCommit = "94aa9078934ce2776ccbb589569eca5ef575f12e"

	gitCommands := []string{
		"git submodule add " + submodDir + " submod",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m 'add submodule' --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo interface {
			ResolveBranch(string) (vcs.CommitID, error)
			FileSystem(vcs.CommitID) (vfs.FileSystem, error)
		}
	}{
		"git libgit2": {
			repo: makeGitRepositoryLibGit2(t, gitCommands...),
		},
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveBranch("master")
		if err != nil {
			t.Fatal(err)
		}

		fs, err := test.repo.FileSystem(commitID)
		if err != nil {
			t.Errorf("%s: FileSystem: %s", label, err)
			continue
		}

		checkSubmoduleFileInfo := func(label string, submod os.FileInfo) {
			if want := "submod"; submod.Name() != want {
				t.Errorf("%s: submod.Name(): got %q, want %q", label, submod.Name(), want)
			}
			// A submodule should have a special file mode and should
			// store information about its origin.
			if mode := submod.Mode(); mode&vcs.ModeSubmodule == 0 {
				t.Errorf("%s: submod.Mode(): got %o, want & vcs.ModeSubmodule (%o) != 0", label, mode, vcs.ModeSubmodule)
			}
			si, ok := submod.Sys().(vcs.SubmoduleInfo)
			if !ok {
				t.Errorf("%s: submod.Sys(): got %v, want SubmoduleInfo", label, si)
			}
			if want := submodDir; si.URL != want {
				t.Errorf("%s: (SubmoduleInfo).URL: got %q, want %q", label, si.URL, want)
			}
			if si.CommitID != submodCommit {
				t.Errorf("%s: (SubmoduleInfo).CommitID: got %q, want %q", label, si.CommitID, submodCommit)
			}
		}

		// Check the submodule os.FileInfo both when it's returned by
		// Stat and when it's returned in a list by ReadDir.
		submod, err := fs.Stat("submod")
		if err != nil {
			t.Errorf("%s: fs.Stat(submod): %s", label, err)
			continue
		}
		checkSubmoduleFileInfo(label+" (Stat)", submod)
		entries, err := fs.ReadDir(".")
		if err != nil {
			t.Errorf("%s: fs.ReadDir(.): %s", label, err)
			continue
		}
		// .gitmodules file is entries[0]
		checkSubmoduleFileInfo(label+" (ReadDir)", entries[1])

		sr, err := fs.Open("submod")
		if err != nil {
			t.Errorf("%s: fs.Open(submod): %s", label, err)
			continue
		}
		if _, err := ioutil.ReadAll(sr); err != nil {
			t.Errorf("%s: ReadAll(submod file): %s", label, err)
			continue
		}
	}
}

func TestOpen(t *testing.T) {
	t.Parallel()
	tests := []struct{ vcs, dir string }{
		{"git", initGitRepository(t)},
		{"hg", initHgRepository(t, "touch x", "hg add x", "hg commit -m foo")},
	}

	for _, test := range tests {
		_, err := vcs.Open(test.vcs, test.dir)
		if err != nil {
			t.Errorf("Open(%q, %q): %s", test.vcs, test.dir, err)
			continue
		}
	}
}

func TestClone(t *testing.T) {
	t.Parallel()
	tests := []struct{ vcs, url, dir string }{
		{"git", initGitRepository(t, "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z --allow-empty"), makeTmpDir(t, "git-clone")},
		{"hg", initHgRepository(t, "touch x", "hg add x", "hg commit -m foo"), makeTmpDir(t, "hg-clone")},
	}

	for _, test := range tests {
		_, err := vcs.Clone(test.vcs, test.url, test.dir, vcs.CloneOpt{})
		if err != nil {
			t.Errorf("Clone(%q, %q, %q): %s", test.vcs, test.url, test.dir, err)
			continue
		}
	}
}

func TestRepository_UpdateEverything(t *testing.T) {
	t.Parallel()

	tests := []struct {
		vcs, baseDir, headDir string

		opener func(dir string) (vcs.Repository, error)

		// newCmds should commit a file "newfile" in the repository
		// root and tag the commit with "second". This is used to test
		// that UpdateEverything picks up the new file from the
		// mirror's origin.
		newCmds []string
	}{
		{
			"git", initGitRepository(t, "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z --allow-empty", "git tag initial"), makeTmpDir(t, "git-clone"),
			func(dir string) (vcs.Repository, error) { return gitcmd.Open(dir) },
			[]string{"touch newfile", "git add newfile", "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m newfile --author='a <a@a.com>' --date 2006-01-02T15:04:05Z", "git tag second"},
		},
		{
			"hg", initHgRepository(t, "touch x", "hg add x", "hg commit -m foo", "hg tag initial"), makeTmpDir(t, "hg-clone"),
			func(dir string) (vcs.Repository, error) { return hgcmd.Open(dir) },
			[]string{"touch newfile", "hg add newfile", "hg commit -m newfile", "hg tag second"},
		},
	}

	for _, test := range tests {
		_, err := vcs.Clone(test.vcs, test.baseDir, test.headDir, vcs.CloneOpt{Bare: true, Mirror: true})
		if err != nil {
			t.Errorf("Clone(%q, %q, %q): %s", test.vcs, test.baseDir, test.headDir, err)
			continue
		}

		r, err := test.opener(test.headDir)
		if err != nil {
			t.Errorf("opener[->%s](%q): %s", reflect.TypeOf(test.opener).Out(0), test.headDir, err)
			continue
		}

		initial, err := r.ResolveTag("initial")
		if err != nil {
			t.Errorf("%s: ResolveTag(%q): %s", test.vcs, "initial", err)
			continue
		}
		fs1, err := r.FileSystem(initial)
		if err != nil {
			t.Errorf("%s: FileSystem(%q): %s", test.vcs, initial, err)
			continue
		}

		// newfile does not yet exist in either the mirror or origin.
		_, err = fs1.Stat("newfile")
		if !os.IsNotExist(err) {
			t.Errorf("%s: fs1.Stat(newfile): got err %v, want os.IsNotExist", test.vcs, err)
			continue
		}

		// run the newCmds to create the new file in the origin repository (NOT
		// the mirror repository; we want to test that UpdateEverything updates the
		// mirror repository).
		for _, cmd := range test.newCmds {
			c := exec.Command("bash", "-c", cmd)
			c.Dir = test.baseDir
			out, err := c.CombinedOutput()
			if err != nil {
				t.Fatalf("%s: exec `%s` failed: %s. Output was:\n\n%s", test.vcs, cmd, err, out)
			}
		}

		// update the mirror.
		err = r.(vcs.RemoteUpdater).UpdateEverything(vcs.RemoteOpts{})
		if err != nil {
			t.Errorf("%s: UpdateEverything: %s", test.vcs, err)
			continue
		}

		// reopen the mirror because the tags/commits changed (after
		// UpdateEverything) and we currently have no way to reload the existing
		// repository.
		r, err = test.opener(test.headDir)
		if err != nil {
			t.Errorf("opener[->%s](%q): %s", reflect.TypeOf(test.opener).Out(0), test.headDir, err)
			continue
		}

		// newfile should exist in the mirror now.
		second, err := r.ResolveTag("second")
		if err != nil {
			t.Errorf("%s: ResolveTag(%q): %s", test.vcs, "second", err)
			continue
		}
		fs2, err := r.FileSystem(second)
		if err != nil {
			t.Errorf("%s: FileSystem(%q): %s", test.vcs, second, err)
			continue
		}
		_, err = fs2.Stat("newfile")
		if err != nil {
			t.Errorf("%s: fs2.Stat(newfile): got err %v, want nil", test.vcs, err)
			continue
		}
	}
}

// initGitRepository initializes a new Git repository and runs cmds in a new
// temporary directory (returned as dir).
func initGitRepository(t testing.TB, cmds ...string) (dir string) {
	dir = makeTmpDir(t, "git")
	cmds = append([]string{"git init"}, cmds...)
	for _, cmd := range cmds {
		c := exec.Command("bash", "-c", cmd)
		c.Dir = dir
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}
	return dir
}

// makeGitRepositoryCmd calls initGitRepository to create a new Git
// (cmd implementation) repository and run cmds in it, and then
// returns the repository.
func makeGitRepositoryCmd(t testing.TB, cmds ...string) *gitcmd.Repository {
	dir := initGitRepository(t, cmds...)
	r, err := gitcmd.Open(dir)
	if err != nil {
		t.Fatalf("gitcmd.Open(%q) failed: %s", dir, err)
	}
	return r
}

// makeGitRepositoryLibGit2 calls initGitRepository to create a new Git
// repository and run cmds in it, and then returns the libgit2-backed
// repository.
func makeGitRepositoryLibGit2(t testing.TB, cmds ...string) *git.Repository {
	dir := initGitRepository(t, cmds...)
	r, err := git.Open(dir)
	if err != nil {
		t.Fatalf("git.Open(%q) failed: %s", dir, err)
	}
	return r
}

// initHgRepository initializes a new Hg repository and runs cmds in a new
// temporary directory (returned as dir).
func initHgRepository(t testing.TB, cmds ...string) (dir string) {
	dir = makeTmpDir(t, "hg")
	cmds = append([]string{"hg init"}, cmds...)
	for _, cmd := range cmds {
		c := exec.Command("bash", "-c", cmd)
		c.Dir = dir
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}
	return dir
}

// makeHgRepositoryCmd calls initHgRepository to create a new Hg (cmd
// implementation) repository and run cmds in it, and then returns the
// repository.
func makeHgRepositoryCmd(t testing.TB, cmds ...string) *hgcmd.Repository {
	dir := initHgRepository(t, cmds...)
	r, err := hgcmd.Open(dir)
	if err != nil {
		t.Fatalf("hgcmd.Open(%q) failed: %s", dir, err)
	}
	return r
}

// makeHgRepositoryNative calls initHgRepository to create a new Hg repository and run
// cmds in it, and then returns the native repository.
func makeHgRepositoryNative(t testing.TB, cmds ...string) *hg.Repository {
	dir := initHgRepository(t, cmds...)
	r, err := hg.Open(dir)
	if err != nil {
		t.Fatalf("hg.Open(%q) failed: %s", dir, err)
	}
	return r
}

func commitsEqual(a, b *vcs.Commit) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a.Author.Date != b.Author.Date {
		return false
	}
	a.Author.Date = b.Author.Date
	if ac, bc := a.Committer, b.Committer; ac != nil && bc != nil {
		if ac.Date != bc.Date {
			return false
		}
		ac.Date = bc.Date
	} else if !(ac == nil && bc == nil) {
		return false
	}
	return reflect.DeepEqual(a, b)
}

func mustParseTime(layout, value string) pbtypes.Timestamp {
	tm, err := time.Parse(layout, value)
	if err != nil {
		panic(err.Error())
	}
	return pbtypes.NewTimestamp(tm)
}

func appleTime(t string) string {
	ti, _ := time.Parse(time.RFC3339, t)
	return ti.Local().Format("200601021504.05")
}

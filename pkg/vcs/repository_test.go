package vcs_test

import (
	"archive/zip"
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
	"sourcegraph.com/sqs/pbtypes"
)

var times = []string{
	appleTime("2006-01-02T15:04:05Z"),
	appleTime("2014-05-06T19:20:21Z"),
}

var nonexistentCommitID = vcs.CommitID(strings.Repeat("a", 40))

var ctx = context.Background()

func TestRepository_ResolveBranch(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo         vcs.Repository
		branch       string
		wantCommitID vcs.CommitID
	}{
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			branch:       "master",
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveRevision(ctx, test.branch)
		if err != nil {
			t.Errorf("%s: ResolveRevision: %s", label, err)
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
	tests := map[string]struct {
		repo    vcs.Repository
		branch  string
		wantErr error
	}{
		"git cmd": {
			repo:    makeGitRepositoryCmd(t, gitCommands...),
			branch:  "doesntexist",
			wantErr: vcs.ErrRevisionNotFound,
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveRevision(ctx, test.branch)
		if err != test.wantErr {
			t.Errorf("%s: ResolveRevision: %s", label, err)
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
	tests := map[string]struct {
		repo         vcs.Repository
		tag          string
		wantCommitID vcs.CommitID
	}{
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			tag:          "t",
			wantCommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveRevision(ctx, test.tag)
		if err != nil {
			t.Errorf("%s: ResolveRevision: %s", label, err)
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
	tests := map[string]struct {
		repo    vcs.Repository
		tag     string
		wantErr error
	}{
		"git cmd": {
			repo:    makeGitRepositoryCmd(t, gitCommands...),
			tag:     "doesntexist",
			wantErr: vcs.ErrRevisionNotFound,
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveRevision(ctx, test.tag)
		if err != test.wantErr {
			t.Errorf("%s: ResolveRevision: %s", label, err)
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
	tests := map[string]struct {
		repo interface {
			Branches(context.Context, vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		wantBranches []*vcs.Branch
	}{
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			wantBranches: []*vcs.Branch{{Name: "b0", Head: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}, {Name: "b1", Head: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}, {Name: "master", Head: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}},
		},
	}

	for label, test := range tests {
		branches, err := test.repo.Branches(ctx, vcs.BranchesOptions{})
		if err != nil {
			t.Errorf("%s: Branches: %s", label, err)
			continue
		}
		sort.Sort(vcs.Branches(branches))
		sort.Sort(vcs.Branches(test.wantBranches))

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

		"git checkout --orphan b2",
		"echo 234 > somefile",
		"git add somefile",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -am foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}

	gitBranches := map[string][]*vcs.Branch{
		"6520a4539a4cb664537c712216a53d80dd79bbdc": { // b1
			{Name: "b0", Head: "6520a4539a4cb664537c712216a53d80dd79bbdc"},
			{Name: "b1", Head: "6520a4539a4cb664537c712216a53d80dd79bbdc"},
		},
		"c3c691fc0fb1844a53b62b179e2fa9fdaf875718": { // b2
			{Name: "b2", Head: "c3c691fc0fb1844a53b62b179e2fa9fdaf875718"},
		},
	}

	for label, test := range map[string]struct {
		repo interface {
			Branches(context.Context, vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		wantBranches map[string][]*vcs.Branch
	}{
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			wantBranches: gitBranches,
		},
	} {
		for branch, mergedInto := range test.wantBranches {
			branches, err := test.repo.Branches(ctx, vcs.BranchesOptions{MergedInto: branch})
			if err != nil {
				t.Errorf("%s: Branches: %s", label, err)
				continue
			}
			if !reflect.DeepEqual(branches, mergedInto) {
				t.Errorf("%s: MergedInto %q: got branches == %v, want %v", label, branch, asJSON(branches), asJSON(mergedInto))
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

	// Pre-sorted branches
	gitWantBranches := map[string][]*vcs.Branch{
		"920c0e9d7b287b030ac9770fd7ba3ee9dc1760d9": {{Name: "branch2", Head: "920c0e9d7b287b030ac9770fd7ba3ee9dc1760d9"}},
		"1224d334dfe08f4693968ea618ad63ae86ec16ca": {{Name: "master", Head: "1224d334dfe08f4693968ea618ad63ae86ec16ca"}},
		"2816a72df28f699722156e545d038a5203b959de": {{Name: "branch2", Head: "920c0e9d7b287b030ac9770fd7ba3ee9dc1760d9"}, {Name: "master", Head: "1224d334dfe08f4693968ea618ad63ae86ec16ca"}},
	}

	tests := map[string]struct {
		repo interface {
			Branches(context.Context, vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		commitToWantBranches map[string][]*vcs.Branch
	}{
		"git cmd": {
			repo:                 makeGitRepositoryCmd(t, gitCommands...),
			commitToWantBranches: gitWantBranches,
		},
	}

	for label, test := range tests {
		for commit, wantBranches := range test.commitToWantBranches {
			branches, err := test.repo.Branches(ctx, vcs.BranchesOptions{ContainsCommit: commit})
			if err != nil {
				t.Errorf("%s: Branches: %s", label, err)
				continue
			}

			sort.Sort(vcs.Branches(branches))
			if !reflect.DeepEqual(branches, wantBranches) {
				t.Errorf("%s: ContainsCommit %q: got branches == %v, want %v", label, commit, asJSON(branches), asJSON(wantBranches))
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
	gitBranches := []*vcs.Branch{
		{Counts: &vcs.BehindAhead{Behind: 5, Ahead: 1}, Name: "old_work", Head: "26692c614c59ddaef4b57926810aac7d5f0e94f0"},
		{Counts: &vcs.BehindAhead{Behind: 0, Ahead: 3}, Name: "dev", Head: "6724953367f0cd9a7755bac46ee57f4ab0c1aad8"},
		{Counts: &vcs.BehindAhead{Behind: 0, Ahead: 0}, Name: "master", Head: "8ea26e077a8fb9aa502c3fe2cfa3ce4e052d1a76"},
	}
	sort.Sort(vcs.Branches(gitBranches))

	tests := map[string]struct {
		repo interface {
			Branches(context.Context, vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		wantBranches []*vcs.Branch
	}{
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			wantBranches: gitBranches,
		},
	}

	for label, test := range tests {
		branches, err := test.repo.Branches(ctx, vcs.BranchesOptions{BehindAheadBranch: "master"})
		if err != nil {
			t.Errorf("%s: Branches: %s", label, err)
			continue
		}
		sort.Sort(vcs.Branches(branches))

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
	wantBranchesGit := []*vcs.Branch{
		{
			Name: "b0", Head: "c4a53701494d1d788b1ceeb8bf32e90224962473",
			Commit: &vcs.Commit{
				ID:        "c4a53701494d1d788b1ceeb8bf32e90224962473",
				Author:    vcs.Signature{Name: "b", Email: "b@b.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &vcs.Signature{Name: "b", Email: "b@b.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Message:   "foo1",
				Parents:   []vcs.CommitID{"a3c1537db9797215208eec56f8e7c9c37f8358ca"},
			},
		},
		{
			Name: "master", Head: "a3c1537db9797215208eec56f8e7c9c37f8358ca",
			Commit: &vcs.Commit{
				ID:        "a3c1537db9797215208eec56f8e7c9c37f8358ca",
				Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "foo0",
				Parents:   nil,
			},
		},
	}

	tests := map[string]struct {
		repo interface {
			Branches(context.Context, vcs.BranchesOptions) ([]*vcs.Branch, error)
		}
		wantBranches []*vcs.Branch
	}{
		"git cmd": {
			repo:         makeGitRepositoryCmd(t, gitCommands...),
			wantBranches: wantBranchesGit,
		},
	}

	for label, test := range tests {
		branches, err := test.repo.Branches(ctx, vcs.BranchesOptions{IncludeCommit: true})
		if err != nil {
			t.Errorf("%s: Branches: %s", label, err)
			continue
		}
		sort.Sort(vcs.Branches(branches))

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
	tests := map[string]struct {
		repo interface {
			Tags(context.Context) ([]*vcs.Tag, error)
		}
		wantTags []*vcs.Tag
	}{
		"git cmd": {
			repo:     makeGitRepositoryCmd(t, gitCommands...),
			wantTags: []*vcs.Tag{{Name: "t0", CommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}, {Name: "t1", CommitID: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"}},
		},
	}

	for label, test := range tests {
		tags, err := test.repo.Tags(ctx)
		if err != nil {
			t.Errorf("%s: Tags: %s", label, err)
			continue
		}
		sort.Sort(vcs.Tags(tags))
		sort.Sort(vcs.Tags(test.wantTags))

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
		Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
		Committer: &vcs.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
		Message:   "bar",
		Parents:   []vcs.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
	}
	tests := map[string]struct {
		repo interface {
			GetCommit(context.Context, vcs.CommitID) (*vcs.Commit, error)
		}
		id         vcs.CommitID
		wantCommit *vcs.Commit
	}{
		"git cmd": {
			repo:       makeGitRepositoryCmd(t, gitCommands...),
			id:         "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommit: wantGitCommit,
		},
	}

	for label, test := range tests {
		commit, err := test.repo.GetCommit(ctx, test.id)
		if err != nil {
			t.Errorf("%s: GetCommit: %s", label, err)
			continue
		}

		if !commitsEqual(commit, test.wantCommit) {
			t.Errorf("%s: got commit == %+v, want %+v", label, commit, test.wantCommit)
		}

		// Test that trying to get a nonexistent commit returns ErrRevisionNotFound.
		if _, err := test.repo.GetCommit(ctx, nonexistentCommitID); err != vcs.ErrRevisionNotFound {
			t.Errorf("%s: for nonexistent commit: got err %v, want %v", label, err, vcs.ErrRevisionNotFound)
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
			Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &vcs.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Message:   "bar",
			Parents:   []vcs.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
		},
		{
			ID:        "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
			Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "foo",
			Parents:   nil,
		},
	}
	tests := map[string]struct {
		repo interface {
			Commits(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error)
		}
		id          vcs.CommitID
		wantCommits []*vcs.Commit
		wantTotal   uint
	}{
		"git cmd": {
			repo:        makeGitRepositoryCmd(t, gitCommands...),
			id:          "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommits: wantGitCommits,
			wantTotal:   2,
		},
	}

	for label, test := range tests {
		commits, total, err := test.repo.Commits(ctx, vcs.CommitsOptions{Head: test.id})
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

		// Test that trying to get a nonexistent commit returns ErrRevisionNotFound.
		if _, _, err := test.repo.Commits(ctx, vcs.CommitsOptions{Head: nonexistentCommitID}); err != vcs.ErrRevisionNotFound {
			t.Errorf("%s: for nonexistent commit: got err %v, want %v", label, err, vcs.ErrRevisionNotFound)
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
			Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &vcs.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Message:   "bar",
			Parents:   []vcs.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
		},
	}
	wantGitCommits2 := []*vcs.Commit{
		{
			ID:        "ade564eba4cf904492fb56dcd287ac633e6e082c",
			Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Committer: &vcs.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Message:   "qux",
			Parents:   []vcs.CommitID{"b266c7e3ca00b1a17ad0b1449825d0854225c007"},
		},
	}
	tests := map[string]struct {
		repo interface {
			Commits(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error)
		}
		opt         vcs.CommitsOptions
		wantCommits []*vcs.Commit
		wantTotal   uint
	}{
		"git cmd": {
			repo:        makeGitRepositoryCmd(t, gitCommands...),
			opt:         vcs.CommitsOptions{Head: "ade564eba4cf904492fb56dcd287ac633e6e082c", N: 1, Skip: 1},
			wantCommits: wantGitCommits,
			wantTotal:   3,
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
	}

	for label, test := range tests {
		commits, total, err := test.repo.Commits(ctx, test.opt)
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
			Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "commit2",
			Parents:   []vcs.CommitID{"a04652fa1998a0a7d2f2f77ecb7021de943d3aab"},
		},
	}
	tests := map[string]struct {
		repo interface {
			Commits(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error)
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
		commits, total, err := test.repo.Commits(ctx, test.opt)
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

	var gitCommitID vcs.CommitID

	if runtime.GOOS == "windows" {
		gitCommitID = ""
	} else {
		gitCommitID = "85d3a39020cf28af4b887552fcab9e31a49f2ced"
	}

	tests := map[string]struct {
		repo     *gitcmd.Repository
		commitID vcs.CommitID
	}{
		// TODO(sqs): implement Lstat and symlink handling for git, git
		// cmd, and hg cmd.

		"git cmd": {
			repo:     makeGitRepositoryCmd(t, gitCommands...),
			commitID: gitCommitID,
		},
	}
	for label, test := range tests {
		ctx := context.Background()

		var commitID string
		if test.commitID == "" {
			commitID = computeCommitHash(test.repo.URL, true)
		} else {
			commitID = string(test.commitID)
		}
		fs := vcs.FileSystem(test.repo, vcs.CommitID(commitID))

		// file1 should be a file.
		file1Info, err := fs.Stat(ctx, "file1")
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
		}

		// link1 should be a link.
		link1Linfo, err := fs.Lstat(ctx, "link1")
		if err != nil {
			t.Errorf("%s: fs.Lstat(link1): %s", label, err)
			continue
		}
		if runtime.GOOS != "windows" {
			// TODO(alexsaveliev) make it work on Windows too
			checkSymlinkFileInfo(label+" (Lstat)", link1Linfo)
		}

		// Also check the FileInfo returned by ReadDir to ensure it's
		// consistent with the FileInfo returned by Lstat.
		entries, err := fs.ReadDir(ctx, ".")
		if err != nil {
			t.Errorf("%s: fs.ReadDir(.): %s", label, err)
			continue
		}
		if got, want := len(entries), 2; got != want {
			t.Errorf("%s: got len(entries) == %d, want %d", label, got, want)
			continue
		}
		if runtime.GOOS != "windows" {
			// TODO(alexsaveliev) make it work on Windows too
			checkSymlinkFileInfo(label+" (ReadDir)", entries[1])
		}

		// link1 stat should follow the link to file1.
		link1Info, err := fs.Stat(ctx, "link1")
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
		if link1Info.Size() != 0 {
			t.Errorf("%s: got link1 Size %d, want %d", label, link1Info.Size(), 0)
		}
	}
}

func TestRepository_FileSystem(t *testing.T) {
	t.Parallel()

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
		"echo -n infile2 > 'file 2'",
		"touch --date=2014-05-06T19:20:21Z 'file 2' || touch -t " + times[1] + " 'file 2'",
		"git add 'file 2'",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --author='a <a@a.com>' --date 2014-05-06T19:20:21Z",
	}
	tests := map[string]struct {
		repo          vcs.Repository
		first, second vcs.CommitID
	}{
		"git cmd": {
			repo:   makeGitRepositoryCmd(t, gitCommands...),
			first:  "b6602ca96bdc0ab647278577a3c6edcb8fe18fb0",
			second: "c5151eceb40d5e625716589b745248e1a6c6228d",
		},
	}

	for label, test := range tests {
		fs1 := vcs.FileSystem(test.repo, test.first)

		// notafile should not exist.
		if _, err := fs1.Stat(ctx, "notafile"); !os.IsNotExist(err) {
			t.Errorf("%s: fs1.Stat(notafile): got err %v, want os.IsNotExist", label, err)
			continue
		}

		// dir1 should exist and be a dir.
		dir1Info, err := fs1.Stat(ctx, "dir1")
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
		if dir1Info.Size() != 0 {
			t.Errorf("%s: got dir1 size %d, want 0", label, dir1Info.Size())
		}

		// dir1 should contain one entry: file1.
		dir1Entries, err := fs1.ReadDir(ctx, "dir1")
		if err != nil {
			t.Errorf("%s: fs1.ReadDir(dir1): %s", label, err)
			continue
		}
		if len(dir1Entries) != 1 {
			t.Errorf("%s: got %d dir1 entries, want 1", label, len(dir1Entries))
			continue
		}
		file1Info := dir1Entries[0]
		if file1Info.Name() != "file1" {
			t.Errorf("%s: got dir1 entry name == %q, want 'file1'", label, file1Info.Name())
		}
		if want := int64(7); file1Info.Size() != want {
			t.Errorf("%s: got dir1 entry size == %d, want %d", label, file1Info.Size(), want)
		}

		// dir1/file1 should exist, contain "infile1", have the right mtime, and be a file.
		file1, err := fs1.Open(ctx, "dir1/file1")
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
		file1Info, err = fs1.Stat(ctx, "dir1/file1")
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
		if want := int64(7); file1Info.Size() != want {
			t.Errorf("%s: got file1 size == %d, want %d", label, file1Info.Size(), want)
		}

		// file 2 shouldn't exist in the 1st commit.
		_, err = fs1.Open(ctx, "file 2")
		if !os.IsNotExist(err) {
			t.Errorf("%s: fs1.Open(file 2): got err %v, want os.IsNotExist (file 2 should not exist in this commit)", label, err)
		}

		// file 2 should exist in the 2nd commit.
		fs2 := vcs.FileSystem(test.repo, test.second)
		_, err = fs2.Open(ctx, "file 2")
		if err != nil {
			t.Errorf("%s: fs2.Open(file 2): %s", label, err)
			continue
		}

		// file1 should also exist in the 2nd commit.
		file1Info, err = fs2.Stat(ctx, "dir1/file1")
		if err != nil {
			t.Errorf("%s: fs2.Stat(dir1/file1): %s", label, err)
			continue
		}
		_, err = fs2.Open(ctx, "dir1/file1")
		if err != nil {
			t.Errorf("%s: fs2.Open(dir1/file1): %s", label, err)
			continue
		}

		// root should exist (via Stat).
		root, err := fs2.Stat(ctx, ".")
		if err != nil {
			t.Errorf("%s: fs2.Stat(.): %s", label, err)
			continue
		}
		if !root.Mode().IsDir() {
			t.Errorf("%s: got root !IsDir", label)
		}

		// root should have 2 entries: dir1 and file 2.
		rootEntries, err := fs2.ReadDir(ctx, ".")
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
		if e1 := rootEntries[1]; !(e1.Name() == "file 2" && !e1.Mode().IsDir()) {
			t.Errorf("%s: got root entry 1 %q IsDir=%v, want 'file 2' IsDir=false", label, e1.Name(), e1.Mode().IsDir())
		}

		// dir1 should still only contain one entry: file1.
		dir1Entries, err = fs2.ReadDir(ctx, "dir1")
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

func TestRepository_FileSystem_quoteChars(t *testing.T) {
	t.Parallel()

	// The repo contains 3 files: one whose filename includes a
	// non-ASCII char, one whose filename contains a double quote, and
	// one whose filename contains a backslash. These should be parsed
	// and unquoted properly.
	//
	// Filenames with double quotes are always quoted in some versions
	// of git, so we might encounter quoted paths even if
	// core.quotepath is off. We test twice, with it both on AND
	// off. (Note: Although
	// https://www.kernel.org/pub/software/scm/git/docs/git-config.html
	// says that double quotes, backslashes, and single quotes are
	// always quoted, this is not true on all git versions, such as
	// @sqs's current git version 2.7.0.)
	wantNames := []string{"⊗.txt", `".txt`, `\.txt`}
	sort.Strings(wantNames)
	gitCommands := []string{
		`touch ⊗.txt '".txt' \\.txt`,
		`git add ⊗.txt '".txt' \\.txt`,
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo vcs.Repository
	}{
		"git cmd (quotepath=on)": {
			repo: makeGitRepositoryCmd(t, append([]string{"git config core.quotepath on"}, gitCommands...)...),
		},
		"git cmd (quotepath=off)": {
			repo: makeGitRepositoryCmd(t, append([]string{"git config core.quotepath off"}, gitCommands...)...),
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveRevision(ctx, "master")
		if err != nil {
			t.Fatal(err)
		}

		fs := vcs.FileSystem(test.repo, commitID)

		entries, err := fs.ReadDir(ctx, ".")
		if err != nil {
			t.Errorf("%s: fs.ReadDir(.): %s", label, err)
			continue
		}
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		sort.Strings(names)

		if !reflect.DeepEqual(names, wantNames) {
			t.Errorf("%s: got names %v, want %v", label, names, wantNames)
			continue
		}

		for _, name := range wantNames {
			stat, err := fs.Stat(ctx, name)
			if err != nil {
				t.Errorf("%s: Stat(%q): %s", label, name, err)
				continue
			}
			if stat.Name() != name {
				t.Errorf("%s: got Name == %q, want %q", label, stat.Name(), name)
				continue
			}
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
		"git submodule add " + filepath.ToSlash(submodDir) + " submod",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m 'add submodule' --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo vcs.Repository
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
		},
	}

	for label, test := range tests {
		commitID, err := test.repo.ResolveRevision(ctx, "master")
		if err != nil {
			t.Fatal(err)
		}

		fs := vcs.FileSystem(test.repo, commitID)

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
			if want := filepath.ToSlash(submodDir); si.URL != want {
				t.Errorf("%s: (SubmoduleInfo).URL: got %q, want %q", label, si.URL, want)
			}
			if si.CommitID != submodCommit {
				t.Errorf("%s: (SubmoduleInfo).CommitID: got %q, want %q", label, si.CommitID, submodCommit)
			}
		}

		// Check the submodule os.FileInfo both when it's returned by
		// Stat and when it's returned in a list by ReadDir.
		submod, err := fs.Stat(ctx, "submod")
		if err != nil {
			t.Errorf("%s: fs.Stat(submod): %s", label, err)
			continue
		}
		checkSubmoduleFileInfo(label+" (Stat)", submod)
		entries, err := fs.ReadDir(ctx, ".")
		if err != nil {
			t.Errorf("%s: fs.ReadDir(.): %s", label, err)
			continue
		}
		// .gitmodules file is entries[0]
		checkSubmoduleFileInfo(label+" (ReadDir)", entries[1])

		sr, err := fs.Open(ctx, "submod")
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

func TestRepository_Archive(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --date=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t " + times[0] + " dir1 dir1/file1",
		"git add dir1/file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo -n infile2 > 'file 2'",
		"touch --date=2014-05-06T19:20:21Z 'file 2' || touch -t " + times[1] + " 'file 2'",
		"git add 'file 2'",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --author='a <a@a.com>' --date 2014-05-06T19:20:21Z",
	}
	tests := map[string]struct {
		repo interface {
			Archive(ctx context.Context, commitID vcs.CommitID) ([]byte, error)
		}
		want map[string]string
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			want: map[string]string{
				"dir1/":      "",
				"dir1/file1": "infile1",
				"file 2":     "infile2",
			},
		},
	}

	for label, test := range tests {
		data, err := test.repo.Archive(ctx, "HEAD")
		if err != nil {
			t.Errorf("%s: Archive: %s", label, err)
			continue
		}
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			t.Errorf("%s: zip.NewReader: %s", label, err)
			continue
		}

		got := map[string]string{}
		for _, f := range zr.File {
			r, err := f.Open()
			contents, err := ioutil.ReadAll(r)
			r.Close()
			if err != nil {
				t.Errorf("%s: Read(%q): %s", label, f.Name, err)
				continue
			}
			got[f.Name] = string(contents)
		}

		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: got %v, want %v", label, got, test.want)
		}
	}
}

func TestOpen(t *testing.T) {
	t.Parallel()

	dir := initGitRepository(t)
	gitcmd.Open(dir)
}

func TestClone(t *testing.T) {
	t.Parallel()

	url := initGitRepository(t, "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z --allow-empty")
	dir := path.Join(makeTmpDir(t, "git-clone"), "repo")
	if err := gitserver.DefaultClient.Clone(context.Background(), dir, url, nil); err != nil {
		t.Errorf("Clone(%q, %q): %s", url, dir, err)
	}
}

func TestRepository_UpdateEverything(t *testing.T) {
	t.Parallel()

	tests := []struct {
		vcs, baseDir, headDir string

		// newCmds should commit a file "newfile" in the repository
		// root and tag the commit with "second". This is used to test
		// that UpdateEverything picks up the new file from the
		// mirror's origin.
		newCmds []string

		wantUpdateResult *vcs.UpdateResult
	}{
		{
			vcs: "git", baseDir: initGitRepositoryWorkingCopy(t, "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z --allow-empty", "git tag initial"), headDir: path.Join(makeTmpDir(t, "git-clone"), "repo"),
			newCmds: []string{"touch newfile", "git add newfile", "GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m newfile --author='a <a@a.com>' --date 2006-01-02T15:04:05Z", "git tag second"},
			wantUpdateResult: &vcs.UpdateResult{
				Changes: []vcs.Change{
					{Op: vcs.FFUpdatedOp, Branch: "master"},
					{Op: vcs.NewOp, Branch: "second"},
				},
			},
		},
	}

	for _, test := range tests {
		if err := gitserver.DefaultClient.Clone(context.Background(), test.headDir, test.baseDir, nil); err != nil {
			t.Errorf("Clone(%q, %q): %s", test.baseDir, test.headDir, err)
			continue
		}

		r := gitcmd.Open(test.headDir)

		initial, err := r.ResolveRevision(ctx, "initial")
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q): %s", test.vcs, "initial", err)
			continue
		}
		fs1 := vcs.FileSystem(r, initial)

		// newfile does not yet exist in either the mirror or origin.
		_, err = fs1.Stat(ctx, "newfile")
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
		makeGitRepositoryBare(t, test.baseDir)

		// update the mirror.
		result, err := r.UpdateEverything(ctx, vcs.RemoteOpts{})
		if err != nil {
			t.Errorf("%s: UpdateEverything: %s", test.vcs, err)
			continue
		}
		if !reflect.DeepEqual(result, test.wantUpdateResult) {
			t.Errorf("%s: got UpdateResult == %v, want %v", test.vcs, asJSON(result), asJSON(test.wantUpdateResult))
			if strings.Contains(asJSON(result), "origin/master") {
				t.Log("NOTE: Some environments and Git versions appear to report the first branch name as 'origin/master', not just 'master' (the desired output). Your environment appears to be affected by this inconsistency/issue. See https://github.com/sourcegraph/go-vcs/issues/90 for the tracking issue.")
			}
		}

		// reopen the mirror because the tags/commits changed (after
		// UpdateEverything) and we currently have no way to reload the existing
		// repository.
		r = gitcmd.Open(test.headDir)

		// newfile should exist in the mirror now.
		second, err := r.ResolveRevision(ctx, "second")
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q): %s", test.vcs, "second", err)
			continue
		}
		fs2 := vcs.FileSystem(r, second)
		if err != nil {
			t.Errorf("%s: FileSystem(%q): %s", test.vcs, second, err)
			continue
		}
		_, err = fs2.Stat(ctx, "newfile")
		if err != nil {
			t.Errorf("%s: fs2.Stat(newfile): got err %v, want nil", test.vcs, err)
			continue
		}
	}
}

// initGitRepository initializes a new Git repository and runs cmds in a new
// temporary directory (returned as dir).
func initGitRepository(t testing.TB, cmds ...string) string {
	dir := initGitRepositoryWorkingCopy(t, cmds...)
	makeGitRepositoryBare(t, dir)
	return dir
}

func initGitRepositoryWorkingCopy(t testing.TB, cmds ...string) (dir string) {
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

func makeGitRepositoryBare(t testing.TB, dir string) {
	c := exec.Command("git", "config", "--bool", "core.bare", "true")
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to convert to bare repo: %s\nOut: %s", err, out)
	}
	wc := dir + "-workingcopy"
	err = os.Rename(dir, wc)
	if err != nil {
		t.Fatalf("Failed to convert to bare repo: %s", err)
	}
	err = os.Rename(filepath.Join(wc, ".git"), dir)
	if err != nil {
		t.Fatalf("Failed to convert to bare repo: %s", err)
	}
}

// makeGitRepositoryCmd calls initGitRepository to create a new Git
// (cmd implementation) repository and run cmds in it, and then
// returns the repository.
func makeGitRepositoryCmd(t testing.TB, cmds ...string) *gitcmd.Repository {
	dir := initGitRepository(t, cmds...)
	return gitcmd.Open(dir)
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

// Computes hash of last commit in a given repo dir
// On Windows, content of a "link file" differs based on the tool that produced it.
// For example:
// - Cygwin may create four different link types, see https://cygwin.com/cygwin-ug-net/using.html#pathnames-symlinks,
// - MSYS's ln copies target file
// Such behavior makes impossible precalculation of SHA hashes to be used in TestRepository_FileSystem_Symlinks
// because for example Git for Windows (http://git-scm.com) is not aware of symlinks and computes link file's SHA which
// may differ from original file content's SHA.
// As a temporary workaround, we calculating SHA hash by asking git/hg to compute it
func computeCommitHash(repoDir string, git bool) string {
	buf := &bytes.Buffer{}

	if git {
		// git cat-file tree "master^{commit}" | git hash-object -t commit --stdin
		cat := exec.Command("git", "cat-file", "commit", "master^{commit}")
		cat.Dir = repoDir
		hash := exec.Command("git", "hash-object", "-t", "commit", "--stdin")
		hash.Stdin, _ = cat.StdoutPipe()
		hash.Stdout = buf
		hash.Dir = repoDir
		_ = hash.Start()
		_ = cat.Run()
		_ = hash.Wait()
	} else {
		hash := exec.Command("hg", "--debug", "id", "-i")
		hash.Dir = repoDir
		hash.Stdout = buf
		_ = hash.Run()
	}
	return strings.TrimSpace(buf.String())
}

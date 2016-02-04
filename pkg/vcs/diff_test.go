package vcs_test

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"src.sourcegraph.com/sourcegraph/pkg/vcs"
)

func TestRepository_Diff(t *testing.T) {
	t.Parallel()

	// TODO(sqs): test ExcludeReachableFromBoth

	gitCommands := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
		"echo line2 >> f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testhead",
	}
	tests := map[string]struct {
		repo interface {
			vcs.Differ
			ResolveRevision(spec string) (vcs.CommitID, error)
		}
		base, head string // can be any revspec; is resolved during the test
		opt        *vcs.DiffOptions

		// wantDiff is the expected diff. In the Raw field,
		// %(headCommitID) is replaced with the actual head commit ID
		// (it seems to change in hg).
		wantDiff *vcs.Diff
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			base: "testbase", head: "testhead",
			wantDiff: &vcs.Diff{
				Raw: "diff --git f f\nindex a29bdeb434d874c9b1d8969c40c42161b03fafdc..c0d0fb45c382919737f8d0c20aaf57cf89b74af8 100644\n--- f\n+++ f\n@@ -1 +1,2 @@\n line1\n+line2\n",
			},
		},
	}

	// TODO(sqs): implement diff for hg native

	for label, test := range tests {
		baseCommitID, err := test.repo.ResolveRevision(test.base)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on base: %s", label, test.base, err)
			continue
		}

		headCommitID, err := test.repo.ResolveRevision(test.head)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on head: %s", label, test.head, err)
			continue
		}

		diff, err := test.repo.Diff(baseCommitID, headCommitID, test.opt)
		if err != nil {
			t.Errorf("%s: Diff(%s, %s, %v): %s", label, baseCommitID, headCommitID, test.opt, err)
			continue
		}

		// Substitute for easier test expectation definition. See the
		// wantDiff field doc for more info.
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(baseCommitID)", string(baseCommitID), -1)
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(headCommitID)", string(headCommitID), -1)
		if runtime.GOOS == "windows" {
			test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "/dev/null", `\dev\null`, -1)
		}

		if !reflect.DeepEqual(diff, test.wantDiff) {
			t.Errorf("%s: diff != wantDiff\n\ndiff ==========\n%s\n\nwantDiff ==========\n%s", label, asJSON(diff), asJSON(test.wantDiff))
		}

		if _, err := test.repo.Diff(nonexistentCommitID, headCommitID, test.opt); err != vcs.ErrCommitNotFound {
			t.Errorf("%s: Diff with bad base commit ID: want ErrCommitNotFound, got %v", label, err)
			continue
		}

		if _, err := test.repo.Diff(baseCommitID, nonexistentCommitID, test.opt); err != vcs.ErrCommitNotFound {
			t.Errorf("%s: Diff with bad head commit ID: want ErrCommitNotFound, got %v", label, err)
			continue
		}
	}
}

func TestRepository_Diff_rename(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
		"git mv f g",
		"git add g",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testhead",
	}
	opt := &vcs.DiffOptions{DetectRenames: true}
	tests := map[string]struct {
		repo interface {
			vcs.Differ
			ResolveRevision(spec string) (vcs.CommitID, error)
		}
		base, head string // can be any revspec; is resolved during the test
		opt        *vcs.DiffOptions

		// wantDiff is the expected diff. In the Raw field,
		// %(headCommitID) is replaced with the actual head commit ID
		// (it seems to change in hg).
		wantDiff *vcs.Diff
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			base: "testbase", head: "testhead",
			wantDiff: &vcs.Diff{
				Raw: "diff --git f g\nsimilarity index 100%\nrename from f\nrename to g\n",
			},
			opt: opt,
		},
	}

	// TODO(sqs): implement diff for hg native

	for label, test := range tests {
		baseCommitID, err := test.repo.ResolveRevision(test.base)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on base: %s", label, test.base, err)
			continue
		}

		headCommitID, err := test.repo.ResolveRevision(test.head)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on head: %s", label, test.head, err)
			continue
		}

		diff, err := test.repo.Diff(baseCommitID, headCommitID, test.opt)
		if err != nil {
			t.Errorf("%s: Diff(%s, %s, %v): %s", label, baseCommitID, headCommitID, test.opt, err)
			continue
		}

		// Substitute for easier test expectation definition. See the
		// wantDiff field doc for more info.
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(baseCommitID)", string(baseCommitID), -1)
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(headCommitID)", string(headCommitID), -1)
		if runtime.GOOS == "windows" {
			test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "/dev/null", `\dev\null`, -1)
		}

		if !reflect.DeepEqual(diff, test.wantDiff) {
			t.Errorf("%s: diff != wantDiff\n\ndiff ==========\n%s\n\nwantDiff ==========\n%s", label, asJSON(diff), asJSON(test.wantDiff))
		}

		if _, err := test.repo.Diff(nonexistentCommitID, headCommitID, test.opt); err != vcs.ErrCommitNotFound {
			t.Errorf("%s: Diff with bad base commit ID: want ErrCommitNotFound, got %v", label, err)
			continue
		}

		if _, err := test.repo.Diff(baseCommitID, nonexistentCommitID, test.opt); err != vcs.ErrCommitNotFound {
			t.Errorf("%s: Diff with bad head commit ID: want ErrCommitNotFound, got %v", label, err)
			continue
		}
	}
}

func TestRepository_CrossRepoDiff_git(t *testing.T) {
	t.Parallel()

	gitCmdsBase := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
	}
	gitCmdsHead := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
		"echo line2 >> f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testhead",
	}
	tests := map[string]struct {
		baseRepo interface {
			vcs.CrossRepoDiffer
			ResolveRevision(spec string) (vcs.CommitID, error)
		}
		headRepo   vcs.Repository
		base, head string // can be any revspec; is resolved during the test
		opt        *vcs.DiffOptions

		// wantDiff is the expected diff. In the Raw field,
		// %(headCommitID) is replaced with the actual head commit ID
		// (it seems to change in hg).
		wantDiff *vcs.Diff
	}{
		"git cmd": {
			baseRepo: makeGitRepositoryCmd(t, gitCmdsBase...),
			headRepo: makeGitRepositoryCmd(t, gitCmdsHead...),
			base:     "testbase", head: "testhead",
			wantDiff: &vcs.Diff{
				Raw: "diff --git f f\nindex a29bdeb434d874c9b1d8969c40c42161b03fafdc..c0d0fb45c382919737f8d0c20aaf57cf89b74af8 100644\n--- f\n+++ f\n@@ -1 +1,2 @@\n line1\n+line2\n",
			},
		},
	}

	// TODO(sqs): implement diff for hg native

	for label, test := range tests {
		baseCommitID, err := test.baseRepo.ResolveRevision(test.base)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on base: %s", label, test.base, err)
			continue
		}

		headCommitID, err := test.headRepo.ResolveRevision(test.head)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on head: %s", label, test.head, err)
			continue
		}

		// Try calling CrossRepoDiff a lot. The git impls do some
		// global state stuff (creating a new remote, fetching into
		// the base).
		const n = 100
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := test.baseRepo.CrossRepoDiff(baseCommitID, test.headRepo, headCommitID, test.opt)
				if err != nil {
					t.Errorf("%s: in concurrency test for CrossRepoDiff(%s, %v, %s, %v): %s", label, baseCommitID, test.headRepo, headCommitID, test.opt, err)
				}
			}()
		}
		wg.Wait()

		diff, err := test.baseRepo.CrossRepoDiff(baseCommitID, test.headRepo, headCommitID, test.opt)
		if err != nil {
			t.Errorf("%s: CrossRepoDiff(%s, %v, %s, %v): %s", label, baseCommitID, test.headRepo, headCommitID, test.opt, err)
			continue
		}

		// Substitute for easier test expectation definition. See the
		// wantDiff field doc for more info.
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(baseCommitID)", string(baseCommitID), -1)
		test.wantDiff.Raw = strings.Replace(test.wantDiff.Raw, "%(headCommitID)", string(headCommitID), -1)

		if !reflect.DeepEqual(diff, test.wantDiff) {
			t.Errorf("%s: diff != wantDiff\n\ndiff ==========\n%s\n\nwantDiff ==========\n%s", label, asJSON(diff), asJSON(test.wantDiff))
		}

		if _, err := test.baseRepo.CrossRepoDiff(nonexistentCommitID, test.headRepo, headCommitID, test.opt); err != vcs.ErrCommitNotFound {
			t.Errorf("%s: CrossRepoDiff with bad base commit ID: want ErrCommitNotFound, got %v", label, err)
			continue
		}

		if _, err := test.baseRepo.CrossRepoDiff(baseCommitID, test.headRepo, nonexistentCommitID, test.opt); err != vcs.ErrCommitNotFound {
			if label == "git cmd" {
				t.Log("skipping failure on git cmd because unimplemented")
				continue
			}
			t.Errorf("%s: CrossRepoDiff with bad head commit ID: want ErrCommitNotFound, got %v", label, err)
			continue
		}
	}
}

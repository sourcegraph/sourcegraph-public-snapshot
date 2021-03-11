package git

import (
	"context"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestRepository_RawLogDiffSearch(t *testing.T) {
	t.Parallel()

	// We depend on newer versions of git. Log git version so if this fails we
	// can compare.
	if version, err := exec.Command("git", "version").CombinedOutput(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(string(version))
	}

	expiredCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Minute))
	defer cancel()
	<-expiredCtx.Done()

	repo := MakeGitRepository(t,
		"echo root > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m root --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag mytag HEAD",

		"git checkout -b branch1",
		"echo branch1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit -m branch1 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",

		"git checkout -b branch2",
		"echo branch2 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m branch2 --author='a <a@a.com>' --date 2006-01-02T15:04:07Z",
	)
	tests := []struct {
		name       string
		ctx        context.Context
		opt        RawLogDiffSearchOptions
		want       []*LogCommitSearchResult
		incomplete bool
		errorS     string
	}{{
		name: "query",
		opt: RawLogDiffSearchOptions{
			Query: TextSearchOptions{Pattern: "root"},
			Diff:  true,
		},
		want: []*LogCommitSearchResult{{
			Commit: Commit{
				ID:        "b9b2349a02271ca96e82c70f384812f9c62c26ab",
				Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Message:   "branch1",
				Parents:   []api.CommitID{"ce72ece27fd5c8180cfbc1c412021d32fd1cda0d"},
			},
			Refs:       []string{"refs/heads/branch1"},
			SourceRefs: []string{"refs/heads/branch2"},
			Diff:       &RawDiff{Raw: "diff --git a/f b/f\nindex d8649da..1193ff4 100644\n--- a/f\n+++ b/f\n@@ -1,1 +1,1 @@\n-root\n+branch1\n"},
		}, {
			Commit: Commit{
				ID:        "ce72ece27fd5c8180cfbc1c412021d32fd1cda0d",
				Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "root",
			},
			Refs:       []string{"refs/heads/master", "refs/tags/mytag"},
			SourceRefs: []string{"refs/heads/branch2"},
			Diff:       &RawDiff{Raw: "diff --git a/f b/f\nnew file mode 100644\nindex 0000000..d8649da\n--- /dev/null\n+++ b/f\n@@ -0,0 +1,1 @@\n+root\n"},
		}},
	}, {
		name: "refglob",
		opt: RawLogDiffSearchOptions{
			Query: TextSearchOptions{Pattern: "root"},
			Diff:  true,
			Args:  []string{"--glob=refs/tags/*"},
		},
		want: []*LogCommitSearchResult{{
			Commit: Commit{
				ID:        "ce72ece27fd5c8180cfbc1c412021d32fd1cda0d",
				Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "root",
			},
			Refs:       []string{"refs/heads/master", "refs/tags/mytag"},
			SourceRefs: []string{"refs/tags/mytag"},
			Diff:       &RawDiff{Raw: "diff --git a/f b/f\nnew file mode 100644\nindex 0000000..d8649da\n--- /dev/null\n+++ b/f\n@@ -0,0 +1,1 @@\n+root\n"},
		}},
	}, {
		name: "empty-query",
		opt: RawLogDiffSearchOptions{
			Query: TextSearchOptions{Pattern: ""},
			Args:  []string{"--grep=branch1|root", "--extended-regexp"},
		},
		want: []*LogCommitSearchResult{{
			Commit: Commit{
				ID:        "b9b2349a02271ca96e82c70f384812f9c62c26ab",
				Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Message:   "branch1",
				Parents:   []api.CommitID{"ce72ece27fd5c8180cfbc1c412021d32fd1cda0d"},
			},
			Refs:       []string{"refs/heads/branch1"},
			SourceRefs: []string{"refs/heads/branch2"},
		}, {
			Commit: Commit{
				ID:        "ce72ece27fd5c8180cfbc1c412021d32fd1cda0d",
				Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "root",
			},
			Refs:       []string{"refs/heads/master", "refs/tags/mytag"},
			SourceRefs: []string{"refs/heads/branch2"},
		}},
	}, {
		name: "path",
		opt: RawLogDiffSearchOptions{
			Paths: PathOptions{
				IncludePatterns: []string{"g"},
				ExcludePattern:  "f",
				IsRegExp:        true,
			},
		},
		want: nil, // empty
	}, {
		name: "deadline",
		ctx:  expiredCtx,
		opt: RawLogDiffSearchOptions{
			Query: TextSearchOptions{Pattern: "root"},
			Diff:  true,
		},
		incomplete: true,
	}, {
		name: "not found",
		opt: RawLogDiffSearchOptions{
			Query: TextSearchOptions{Pattern: "root"},
			Diff:  true,
			Args:  []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		},
		errorS: "fatal: bad object aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := test.ctx
			if ctx == nil {
				ctx = context.Background()
			}

			results, complete, err := RawLogDiffSearch(ctx, repo, test.opt)
			if err != nil {
				if test.errorS == "" {
					t.Fatal(err)
				} else if !strings.Contains(err.Error(), test.errorS) {
					t.Fatalf("error should contain %q: %v", test.errorS, err)
				}
				return
			} else if test.errorS != "" {
				t.Fatal("expected error")
			}

			if complete == test.incomplete {
				t.Fatalf("complete is %v", complete)
			}
			for _, r := range results {
				r.DiffHighlights = nil // Highlights is tested separately
			}
			if !cmp.Equal(test.want, results) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(results, test.want))
			}
		})
	}
}

func TestRepository_RawLogDiffSearch_empty(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m empty --allow-empty --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m empty --allow-empty --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo api.RepoName
		want map[*RawLogDiffSearchOptions][]*LogCommitSearchResult
	}{
		"commit": {
			repo: MakeGitRepository(t, gitCommands...),
			want: map[*RawLogDiffSearchOptions][]*LogCommitSearchResult{
				{
					Paths: PathOptions{IncludePatterns: []string{"/xyz.txt"}, IsRegExp: true},
				}: nil, // want no matches
			},
		},
		"repo": {
			repo: MakeGitRepository(t),
			want: map[*RawLogDiffSearchOptions][]*LogCommitSearchResult{
				{
					Paths: PathOptions{IncludePatterns: []string{"/xyz.txt"}, IsRegExp: true},
				}: nil, // want no matches
			},
		},
	}

	for label, test := range tests {
		for opt, want := range test.want {
			results, complete, err := RawLogDiffSearch(context.Background(), test.repo, *opt)
			if err != nil {
				t.Errorf("%s: %+v: %s", label, *opt, err)
				continue
			}
			if !complete {
				t.Errorf("%s: !complete", label)
			}
			for _, r := range results {
				r.DiffHighlights = nil // Highlights is tested separately
			}
			if !reflect.DeepEqual(results, want) {
				t.Errorf("%s: %+v: got %+v, want %+v", label, *opt, AsJSON(results), AsJSON(want))
			}
		}
	}
}

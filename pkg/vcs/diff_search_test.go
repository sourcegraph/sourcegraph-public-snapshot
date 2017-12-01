package vcs_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestRepository_RawLogDiffSearch(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"echo root > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m root --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",

		"git checkout -b branch1",
		"echo branch1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit -m branch1 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",

		"git checkout -b branch2",
		"echo branch2 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m branch2 --author='a <a@a.com>' --date 2006-01-02T15:04:07Z",
	}
	tests := map[string]struct {
		repo interface {
			RawLogDiffSearch(ctx context.Context, opt vcs.RawLogDiffSearchOptions) ([]*vcs.LogCommitSearchResult, bool, error)
		}
		want map[*vcs.RawLogDiffSearchOptions][]*vcs.LogCommitSearchResult
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, gitCommands...),
			want: map[*vcs.RawLogDiffSearchOptions][]*vcs.LogCommitSearchResult{
				{
					Query: vcs.TextSearchOptions{Pattern: "root"},
				}: {
					{
						Commit: vcs.Commit{
							ID:        "b9b2349a02271ca96e82c70f384812f9c62c26ab",
							Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
							Committer: &vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
							Message:   "branch1",
							Parents:   []vcs.CommitID{"ce72ece27fd5c8180cfbc1c412021d32fd1cda0d"},
						},
						Refs:       []string{"refs/heads/branch1"},
						SourceRefs: []string{"HEAD"},
						Diff:       &vcs.Diff{Raw: "diff --git a/f b/f\nindex d8649da..1193ff4 100644\n--- a/f\n+++ b/f\n@@ -1,1 +1,1 @@\n-root\n+branch1\n"},
					},
					{
						Commit: vcs.Commit{
							ID:        "ce72ece27fd5c8180cfbc1c412021d32fd1cda0d",
							Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
							Committer: &vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
							Message:   "root",
						},
						Refs:       []string{"refs/heads/master"},
						SourceRefs: []string{"HEAD"},
						Diff:       &vcs.Diff{Raw: "diff --git a/f b/f\nnew file mode 100644\nindex 0000000..d8649da\n--- /dev/null\n+++ b/f\n@@ -0,0 +1,1 @@\n+root\n"},
					},
				},

				// Without query
				{
					Query: vcs.TextSearchOptions{Pattern: ""},
					Args:  []string{"--grep=branch1|root", "--extended-regexp"},
				}: {
					{
						Commit: vcs.Commit{
							ID:        "b9b2349a02271ca96e82c70f384812f9c62c26ab",
							Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
							Committer: &vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
							Message:   "branch1",
							Parents:   []vcs.CommitID{"ce72ece27fd5c8180cfbc1c412021d32fd1cda0d"},
						},
						Refs:       []string{"refs/heads/branch1"},
						SourceRefs: []string{"HEAD"},
					},
					{
						Commit: vcs.Commit{
							ID:        "ce72ece27fd5c8180cfbc1c412021d32fd1cda0d",
							Author:    vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
							Committer: &vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
							Message:   "root",
						},
						Refs:       []string{"refs/heads/master"},
						SourceRefs: []string{"HEAD"},
					},
				},
			},
		},
	}

	for label, test := range tests {
		for opt, want := range test.want {
			results, complete, err := test.repo.RawLogDiffSearch(ctx, *opt)
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
				t.Errorf("%s: %+v: got %+v, want %+v", label, *opt, asJSON(results), asJSON(want))
			}
		}
	}
}

package gitcli

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

func TestMergeBase(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		cmds []string
		a, b string // can be any revspec; is resolved during the test

		wantMergeBase string // can be any revspec; is resolved during test
	}{
		"basic": {
			cmds: []string{
				"echo line1 > f",
				"git add f",
				"git commit -m foo",
				"git tag testbase",
				"git checkout -b b2",
				"echo line2 >> f",
				"git add f",
				"git commit -m foo",
				"git checkout master",
				"echo line3 > h",
				"git add h",
				"git commit -m qux",
			},
			a: "master", b: "b2",
			wantMergeBase: "testbase",
		},
		"orphan branches": {
			cmds: []string{
				"echo line1 > f",
				"git add f",
				"git commit -m foo",
				"git checkout --orphan b2",
				"echo line2 >> f",
				"git add f",
				"git commit -m foo",
				"git checkout master",
			},
			a: "master", b: "b2",
			wantMergeBase: "",
		},
	}

	for label, test := range tests {
		repoRoot := t.TempDir()
		for _, cmd := range append([]string{"git init --initial-branch=master ."}, test.cmds...) {
			out, err := gitserver.CreateGitCommand(repoRoot, "bash", "-c", cmd).CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to run git command. Output was:\n\n%s", out)
			}
		}

		repoName := api.RepoName(filepath.Base(repoRoot))

		backend := NewBackend(logtest.Scoped(t), wrexec.NewNoOpRecordingCommandFactory(), common.GitDir(filepath.Join(repoRoot, ".git")), repoName)

		mb, err := backend.MergeBase(ctx, test.a, test.b)
		if err != nil {
			t.Errorf("%s: MergeBase(%s, %s): %s", label, test.a, test.b, err)
			continue
		}

		var want []byte
		if test.wantMergeBase != "" {
			want, err = exec.CommandContext(ctx, "git", "-C", repoRoot, "rev-parse", test.wantMergeBase).Output()
			require.NoError(t, err)
		}

		if mb != api.CommitID(strings.TrimSpace(string(want))) {
			t.Errorf("%s: MergeBase(%s, %s): got %q, want %q", label, test.a, test.b, mb, want)
			continue
		}
	}
}

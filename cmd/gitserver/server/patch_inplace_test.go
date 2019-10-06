package server

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

func TestCreateCommitFromPatch(t *testing.T) {
	const (
		emptyGitRepo = "git init && GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m msg --author='a <a@a.com>' --date 2006-01-02T15:04:05Z"
		gitRepoWithF = "git init && echo -n x > f && git add f && GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m msg --author='a <a@a.com>' --date 2006-01-02T15:04:05Z"
	)
	tests := map[string]struct {
		initCommand string
		rawDiff     string
		want        string
	}{
		"empty": {
			initCommand: emptyGitRepo,
			rawDiff:     "",
			want:        "93760e2e285a8cc28d5fe876ba623c6a78ad1fb2",
		},
		"edit": {
			initCommand: gitRepoWithF,
			rawDiff: `diff --git a/f b/f
--- a/f
+++ b/f
@@ -1,0 +1,0 @@
-x
\ No newline at end of file
+y
\ No newline at end of file`,
			want: "8c23db6fecdc32c82f2180d8990d803cc80f8d48",
		},
		"create": {
			initCommand: emptyGitRepo,
			rawDiff: `diff --git a/f b/f
--- /dev/null
+++ b/f
@@ -0,0 +1,0 @@
+z
\ No newline at end of file`,
			want: "9e46269ac0e8faa72cd7dff5aa13ed634722b3d7",
		},
	}

	makeRepo := func(t *testing.T, dir, initCommand string) {
		t.Helper()
		cmd := exec.Command("sh", "-c", initCommand)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	}

	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			dir, err := ioutil.TempDir("", label)
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			makeRepo(t, dir, test.initCommand)
			oid, err := createCommitFromPatch(context.Background(), dir, protocol.CreateCommitFromPatchRequest{
				BaseCommit: "HEAD",
				Patch:      test.rawDiff,
			})
			if err != nil {
				t.Fatal(err)
			}
			if oid != test.want {
				t.Errorf("got oid %q, want %q", oid, test.want)
			}
		})
	}
}

func TestApplyPatch(t *testing.T) {
	tests := []struct {
		data  string
		patch string
		want  string
	}{
		{
			data: ``,
			patch: `--- f
+++ f
@@ -0,0 +1,0 @@
+z
\ No newline at end of file`,
			want: `z`,
		},
		{
			data: `a`,
			patch: `--- f
+++ f
@@ -1,1 +1,1 @@
-a
\ No newline at end of file
+x
\ No newline at end of file`,
			want: `x`,
		},
		{
			data: `a
b
c
d
`,
			patch: `--- f
+++ f
@@ -1,4 +1,6 @@
-a
+w
+x
 b
-c
+y
 d
+z`,
			want: `w
x
b
y
d
z
`,
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				if e := recover(); e != nil {
					t.Fatal(e)
				}
			}()
			fileDiff, err := diff.ParseFileDiff([]byte(test.patch))
			if err != nil {
				t.Fatal(err)
			}
			got := applyPatch([]byte(test.data), fileDiff)
			if string(got) != test.want {
				t.Errorf("got:  %q\nwant: %q", got, test.want)
			}
		})
	}
}

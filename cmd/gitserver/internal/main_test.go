package internal

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	} else {
		logtest.Init(m)
	}
	os.Exit(m.Run())
}

func mkFiles(t *testing.T, root string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Join(root, filepath.Dir(p)), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(root, p), nil)
	}
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	err := os.WriteFile(path, content, 0o666)
	if err != nil {
		t.Fatal(err)
	}
}

func runCmd(t *testing.T, dir string, cmd string, arg ...string) string {
	t.Helper()
	c := exec.Command(cmd, arg...)
	c.Dir = dir
	c.Env = []string{
		"GIT_COMMITTER_NAME=a",
		"GIT_COMMITTER_EMAIL=a@a.com",
		"GIT_AUTHOR_NAME=a",
		"GIT_AUTHOR_EMAIL=a@a.com",
	}
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %s\nOutput: %s", cmd, strings.Join(arg, " "), err, b)
	}
	return string(b)
}

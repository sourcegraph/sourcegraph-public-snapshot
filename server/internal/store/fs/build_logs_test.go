package fs

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func logStr(e *sourcegraph.LogEntries) string { return strings.Join(e.Entries, "\n") }

var task = sourcegraph.TaskSpec{Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "r"}, ID: 123}, ID: 456}

func writeBuildLog(ctx context.Context, t *testing.T, task sourcegraph.TaskSpec, data string) {
	if err := ioutil.WriteFile(logFilePath(task), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func init() {
	tmpDir, err := ioutil.TempDir("", "BuildLogs-test")
	if err != nil {
		panic(err)
	}
	conf.BuildLogDir = tmpDir
}

func TestBuildLogs_Get_noErrorIfNotExist(t *testing.T) {
	ctx, done := testContext()
	defer done()

	s := &BuildLogs{}
	e, err := s.Get(ctx, task, "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if want := ""; logStr(e) != want {
		t.Errorf("got log %q, want %q", logStr(e), want)
	}
}

func TestBuildLogs_Get_noErrorIfEmpty(t *testing.T) {
	ctx, done := testContext()
	defer done()

	s := &BuildLogs{}
	writeBuildLog(ctx, t, task, "")

	e, err := s.Get(ctx, task, "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if want := ""; logStr(e) != want {
		t.Errorf("got log %q, want %q", logStr(e), want)
	}
}

func TestBuildLogs_Get_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.BuildLogs_Get_ok(ctx, t, &BuildLogs{}, writeBuildLog)
}

func TestBuildLogs_Get_MinID(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.BuildLogs_Get_MinID(ctx, t, &BuildLogs{}, writeBuildLog)
}

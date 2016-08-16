package localstore

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

func logStr(e *sourcegraph.LogEntries) string { return strings.Join(e.Entries, "\n") }

var task = sourcegraph.TaskSpec{Build: sourcegraph.BuildSpec{Repo: 1, ID: 123}, ID: 456}

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
	ctx := testContextNoDB()

	s := &buildLogs{}
	e, err := s.Get(ctx, task, "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if want := ""; logStr(e) != want {
		t.Errorf("got log %q, want %q", logStr(e), want)
	}
}

func TestBuildLogs_Get_noErrorIfEmpty(t *testing.T) {
	ctx := testContextNoDB()

	s := &buildLogs{}
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
	ctx := testContextNoDB()

	s := &buildLogs{}
	writeBuildLog(ctx, t, task, "hello\nworld")
	e, err := s.Get(ctx, task, "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if want := "hello\nworld"; logStr(e) != want {
		t.Errorf("got log %q, want %q", logStr(e), want)
	}
}

func TestBuildLogs_Get_MinID(t *testing.T) {
	ctx := testContextNoDB()

	s := &buildLogs{}
	const data = "a\nb\nc\nd"
	writeBuildLog(ctx, t, task, data)

	// NOTE: These MinIDs are based on how the fs-backed BuildLogs is
	// implemented. If we need to test other stores, we'll need to
	// abstract out the hard-coded MinIDs here.
	tests := []struct {
		minID string
		want  string
	}{
		{"", data},
		{"0", data},
		{"-1", data},
		{"1", "b\nc\nd"},
		{"2", "c\nd"},
		{"3", "d"},
		{"4", ""},
		{"5", ""},
		{"6", ""},
	}
	for _, test := range tests {
		e, err := s.Get(ctx, task, test.minID, time.Time{}, time.Time{})
		if err != nil {
			t.Errorf("MinID %q: %s", test.minID, err)
			continue
		}
		if logStr(e) != test.want {
			t.Errorf("MinID %q: got log %q, want %q", test.minID, logStr(e), test.want)
			continue
		}
		if want := "3"; e.MaxID != want {
			t.Errorf("MinID %q: got MaxID %q, want %q", test.minID, e.MaxID, want)
			continue
		}
	}
}

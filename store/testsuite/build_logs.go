package testsuite

import (
	"testing"
	"time"

	"strings"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// WriteBuildLogFunc writes a mock build log for the given build and
// tag.
type WriteBuildLogFunc func(ctx context.Context, t *testing.T, task sourcegraph.TaskSpec, data string)

func logStr(e *sourcegraph.LogEntries) string { return strings.Join(e.Entries, "\n") }

var task = sourcegraph.TaskSpec{Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "r"}, ID: 123}, ID: 456}

func BuildLogs_Get_noErrorIfNotExist(ctx context.Context, t *testing.T, s store.BuildLogs, write WriteBuildLogFunc) {
	e, err := s.Get(ctx, task, "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if want := ""; logStr(e) != want {
		t.Errorf("got log %q, want %q", logStr(e), want)
	}
}

func BuildLogs_Get_noErrorIfEmpty(ctx context.Context, t *testing.T, s store.BuildLogs, write WriteBuildLogFunc) {
	write(ctx, t, task, "")

	e, err := s.Get(ctx, task, "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if want := ""; logStr(e) != want {
		t.Errorf("got log %q, want %q", logStr(e), want)
	}
}

func BuildLogs_Get_ok(ctx context.Context, t *testing.T, s store.BuildLogs, write WriteBuildLogFunc) {
	write(ctx, t, task, "hello\nworld")
	e, err := s.Get(ctx, task, "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if want := "hello\nworld"; logStr(e) != want {
		t.Errorf("got log %q, want %q", logStr(e), want)
	}
}

func BuildLogs_Get_MinID(ctx context.Context, t *testing.T, s store.BuildLogs, write WriteBuildLogFunc) {
	const data = `a
b
c
d`
	write(ctx, t, task, data)

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

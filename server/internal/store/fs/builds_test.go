package fs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

var tmpBuildStore = filepath.Join(os.TempDir(), "fs.builds.tests")

// setBuilds creates a mock tasks based on the path deduced from the passed context.
// This method also completely erases the contents of that path.
func (s *Builds) setBuilds(ctx context.Context, t *testing.T, mockBuilds []*sourcegraph.Build) {
	if err := os.RemoveAll(tmpBuildStore); err != nil {
		t.Fatalf("failed to reset test ground %s", err)
	}
	if err := os.Mkdir(tmpBuildStore, 0700); err != nil {
		t.Fatalf("failed to re-create test dir %s", err)
	}
	for _, b := range mockBuilds {
		if _, err := s.Create(ctx, b); err != nil {
			t.Fatalf("failed to create build file %s", err)
		}
	}
}

// setTasks creates a mock tasks based on the path deduced from the passed context.
// This method also completely erases the contents of that path.
func (s *Builds) setTasks(ctx context.Context, t *testing.T, mockTasks []*sourcegraph.BuildTask) {
	if err := os.RemoveAll(tmpBuildStore); err != nil {
		t.Fatalf("failed to reset test ground %s", err)
	}
	if err := os.Mkdir(tmpBuildStore, 0700); err != nil {
		t.Fatalf("failed to re-create test dir %s", err)
	}
	for _, task := range mockTasks {
		err := s.updateTask(ctx, task)
		if err != nil {
			t.Fatalf("failed to create build file %s", err)
		}
	}
}

// queueEntryExists will validate whether the passed BuildSpec exists in the queue.
func (s *Builds) queueEntryExists(ctx context.Context, want sourcegraph.BuildSpec, t *testing.T) bool {
	var queue []sourcegraph.BuildSpec
	if err := getQueue(ctx, buildQueueFilename, &queue); err != nil {
		t.Fatalf("could not get queue: %#v", err)
	}
	var found bool
	for _, q := range queue {
		if reflect.DeepEqual(q, want) {
			found = true
		}
	}
	return found
}

// createTestContext creates a new context to be used with tests, it sets up a subfolder
// in the default OS's temporary directory to be used.
func createTestContext(t *testing.T) context.Context {
	os.RemoveAll(tmpBuildStore)
	if err := os.Mkdir(tmpBuildStore, 0700); err != nil && !os.IsExist(err) {
		t.Fatalf("failed to create temporary directory for VFS %s", err)
	}
	tmpRepo := filepath.Join(os.TempDir(), "fs.builds.repos")
	os.RemoveAll(tmpRepo)
	if err := os.Mkdir(tmpRepo, 0700); err != nil && !os.IsExist(err) {
		t.Fatalf("failed to create temporary directory for VFS %s", err)
	}
	ctx := WithBuildStoreVFS(context.Background(), rwvfs.Walkable(rwvfs.OS(tmpBuildStore)))
	return WithReposVFS(ctx, tmpRepo)
}

func TestBuilds_GetFirstInCommitOrder_firstCommitIDMatch(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_GetFirstInCommitOrder_firstCommitIDMatch(createTestContext(t), t, s, s.setBuilds)
}

func TestBuilds_GetFirstInCommitOrder_secondCommitIDMatch(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_GetFirstInCommitOrder_secondCommitIDMatch(createTestContext(t), t, s, s.setBuilds)
}

func TestBuilds_GetFirstInCommitOrder_successfulOnly(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_GetFirstInCommitOrder_successfulOnly(createTestContext(t), t, s, s.setBuilds)
}

func TestBuilds_GetFirstInCommitOrder_noneFound(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_GetFirstInCommitOrder_noneFound(createTestContext(t), t, s, s.setBuilds)
}

func TestBuilds_GetFirstInCommitOrder_returnNewest(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_GetFirstInCommitOrder_returnNewest(createTestContext(t), t, s, s.setBuilds)
}

func TestBuilds_Get(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_Get(createTestContext(t), t, s, s.setBuilds)
}

func TestBuilds_Create(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_Create(createTestContext(t), t, s)
}

func TestBuilds_Create_New(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_Create_New(createTestContext(t), t, s)
}

func TestBuilds_Create_SequentialAttempt(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_Create_SequentialAttempt(createTestContext(t), t, s)
}

func TestBuilds_Update(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_Update(createTestContext(t), t, s, s.setBuilds)
}

func TestBuilds_DequeueNext(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_DequeueNext(createTestContext(t), t, s, s.setBuilds)
}

func TestBuilds_CreateTasks(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_CreateTasks(createTestContext(t), t, s, s.setTasks)
}

func TestBuilds_UpdateTask(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_UpdateTask(createTestContext(t), t, s, s.setTasks)
}

func TestBuilds_ListBuildTasks(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_ListBuildTasks(createTestContext(t), t, s, s.setTasks)
}

func TestBuilds_GetTask(t *testing.T) {
	s := NewBuildStore()
	testsuite.Builds_GetTask(createTestContext(t), t, s, s.setTasks)
}

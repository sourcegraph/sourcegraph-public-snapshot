package testsuite

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// InsertBuildsFunc is called at the beginning of Builds_* test funcs
// to insert or mock the builds data source.
type InsertBuildsFunc func(ctx context.Context, t *testing.T, mockBuilds []*sourcegraph.Build)

// InsertTasksFunc is called at the beginning of test funcs that need
// to add tasks during their setup.
type InsertTasksFunc func(ctx context.Context, t *testing.T, mockTasks []*sourcegraph.BuildTask)

// ValidateQueueEntryFunc is called by tests that wish to validate the existence
// of a build in a queue. This is left to the implementation layer because the
// check method may vary.
type ValidateQueueEntryFunc func(ctx context.Context, want sourcegraph.BuildSpec, t *testing.T) bool

// assertBuildExists verifies that a build exists in the store by using its Get method.
func assertBuildExists(ctx context.Context, s store.Builds, want *sourcegraph.Build, t *testing.T) {
	b, err := s.Get(ctx, want.Spec())
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if !reflect.DeepEqual(b, want) {
		t.Errorf("expected %#v, got %#v", want, b)
	}
}

// assertTaskExists verifies that a build exists in the store by using its GetTask method.
func assertTaskExists(ctx context.Context, s store.Builds, want *sourcegraph.BuildTask, t *testing.T) {
	b, err := s.GetTask(ctx, want.Spec())
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if !reflect.DeepEqual(b, want) {
		t.Errorf("expected %#v, got %#v", want, b)
	}
}

func Builds_GetTask(ctx context.Context, t *testing.T, s store.Builds, insert InsertTasksFunc) {
	tasks := []*sourcegraph.BuildTask{
		{ID: 1, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "b"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 2}, Label: "a"},
	}
	insert(ctx, t, tasks)
	for _, tsk := range tasks {
		assertTaskExists(ctx, s, tsk, t)
	}
}

package localstore

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func TestQueue_LockJob_AlreadyLocked(t *testing.T) {
	t.Parallel()

	q := &queue{}
	ctx, _, done := testContext()
	defer done()

	if err := q.Enqueue(ctx, &store.Job{Type: "MyJob"}); err != nil {
		t.Fatal(err)
	}

	j, err := q.LockJob(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if j == nil {
		t.Fatal("wanted job, got none")
	}

	j2, err := q.LockJob(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if j2 != nil {
		t.Fatalf("wanted no job, got %+v", j2)
	}

	err = j.MarkSuccess()
	if err != nil {
		t.Fatal("delete job failed:", err)
	}
}

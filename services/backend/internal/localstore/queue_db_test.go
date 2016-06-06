package localstore

import (
	"fmt"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func TestQueue_LockJob_AlreadyLocked(t *testing.T) {
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

func TestQueue_LockJob_BoundedAttempts(t *testing.T) {
	q := &queue{}
	ctx, _, done := testContext()
	defer done()

	if err := q.Enqueue(ctx, &store.Job{Type: "MyJob"}); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < queueMaxAttempts; i++ {
		j, err := q.LockJob(ctx)
		if err != nil {
			t.Fatalf("attempt %d failed: %s", i, err)
		}
		if j == nil {
			t.Fatalf("attempt %d wanted job, got none", i)
		}
		err = j.MarkError(fmt.Sprintf("attempt %d", i))
		if err != nil {
			t.Fatalf("attempt %d MarkError failed: %s", i, err)
		}
		// Update run_at so we don't need to add sleeps for the
		// exponential backoff
		_, err = appDBH(ctx).Exec("UPDATE que_jobs SET run_at=now();")
		if err != nil {
			t.Fatalf("attempt %d update failed: %s", i, err)
		}
	}

	j, err := q.LockJob(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if j != nil {
		t.Fatalf("wanted no job, got %+v", j)
	}
}

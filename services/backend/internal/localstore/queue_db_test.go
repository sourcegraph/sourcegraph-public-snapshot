package localstore

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func TestQueue_LockJob_AlreadyLocked(t *testing.T) {
	q := &queue{}
	ctx, done := testContext()
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
	ctx, done := testContext()
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

func TestQueue_Job_Delay(t *testing.T) {
	// We don't actually test the DB implementation, we assume the
	// underlying library works. We do this to avoid having a test that
	// relies on time.Sleep.
	q := &queue{}
	now := time.Now()

	j := q.toQue(&store.Job{})
	if j.RunAt.Before(now) {
		t.Errorf("job without delay scheduled to run in the past. want %s <= %s", now, j.RunAt)
	}

	j = q.toQue(&store.Job{Delay: 10 * time.Minute})
	if j.RunAt.Before(now.Add(9 * time.Minute)) {
		t.Errorf("job scheduled too early. wanted roughly %s, got %s", now.Add(10*time.Minute), j.RunAt)
	}
	if j.RunAt.After(now.Add(11 * time.Minute)) {
		t.Errorf("job scheduled too late. wanted roughly %s, got %s", now.Add(10*time.Minute), j.RunAt)
	}
}

func TestQueue_Stats(t *testing.T) {
	q := &queue{}
	ctx, done := testContext()
	defer done()

	push := func(qt string) {
		if err := q.Enqueue(ctx, &store.Job{Type: qt}); err != nil {
			t.Fatal(err)
		}
	}

	// Several "a" jobs, mark one as error and finish one
	for i := 0; i < 10; i++ {
		push("a")
	}
	j1, err := q.LockJob(ctx)
	if err != nil {
		t.Fatal(err)
	}
	j2, err := q.LockJob(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = j1.MarkError("test")
	if err != nil {
		t.Fatal(err)
	}
	err = j2.MarkSuccess()
	if err != nil {
		t.Fatal(err)
	}

	push("b")

	stats, err := q.Stats(ctx)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]store.QueueStats{
		"a": {
			NumJobs:          9,
			NumJobsWithError: 1,
		},
		"b": {
			NumJobs: 1,
		},
	}
	if !reflect.DeepEqual(stats, want) {
		t.Fatalf("q.Stats got %+v, want %+v", stats, want)
	}
}

package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestPermissionSyncJobs_CreateAndList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := timeutil.NewFakeClock(time.Now(), 0)

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	store := PermissionSyncJobsWith(logger, db)

	jobs, err := store.List(ctx, ListPermissionSyncJobOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(jobs) != 0 {
		t.Fatalf("jobs returned even though database is empty")
	}

	opts := PermissionSyncJobOpts{HighPriority: true, InvalidateCaches: true}
	if err := store.CreateRepoSyncJob(ctx, 99, opts); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	nextSyncAt := clock.Now().Add(5 * time.Minute)
	opts = PermissionSyncJobOpts{HighPriority: false, InvalidateCaches: true, NextSyncAt: nextSyncAt}
	if err := store.CreateUserSyncJob(ctx, 77, opts); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	jobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("wrong number of jobs returned. want=%d, have=%d", 2, len(jobs))
	}

	wantJobs := []*PermissionSyncJob{
		{
			ID:               jobs[0].ID,
			State:            "queued",
			RepositoryID:     99,
			HighPriority:     true,
			InvalidateCaches: true,
		},
		{
			ID:               jobs[1].ID,
			State:            "queued",
			UserID:           77,
			InvalidateCaches: true,
			ProcessAfter:     nextSyncAt,
		},
	}
	if diff := cmp.Diff(jobs, wantJobs, cmpopts.IgnoreFields(PermissionSyncJob{}, "QueuedAt")); diff != "" {
		t.Fatalf("jobs[0] has wrong attributes: %s", diff)
	}
	for i, j := range jobs {
		if j.QueuedAt.IsZero() {
			t.Fatalf("job %d has no QueuedAt set", i)
		}
	}

	listTests := []struct {
		name     string
		opts     ListPermissionSyncJobOpts
		wantJobs []*PermissionSyncJob
	}{
		{
			name:     "ID",
			opts:     ListPermissionSyncJobOpts{ID: jobs[0].ID},
			wantJobs: jobs[0:1],
		},
		{
			name:     "RepoID",
			opts:     ListPermissionSyncJobOpts{RepoID: jobs[0].RepositoryID},
			wantJobs: jobs[0:1],
		},
		{
			name:     "UserID",
			opts:     ListPermissionSyncJobOpts{UserID: jobs[1].UserID},
			wantJobs: jobs[1:],
		},
	}

	for _, tt := range listTests {
		t.Run(tt.name, func(t *testing.T) {
			have, err := store.List(ctx, tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if len(have) != len(tt.wantJobs) {
				t.Fatalf("wrong number of jobs returned. want=%d, have=%d", len(tt.wantJobs), len(have))
			}
			if diff := cmp.Diff(have, tt.wantJobs); diff != "" {
				t.Fatalf("unexpected jobs. diff: %s", diff)
			}
		})
	}
}

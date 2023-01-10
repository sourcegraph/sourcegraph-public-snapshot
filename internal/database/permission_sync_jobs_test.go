package database

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestPermissionSyncJobs_CreateAndList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	store := PermissionSyncJobsWith(logger, db)

	opts := PermissionSyncJobOpts{HighPriority: true, InvalidateCaches: true}
	if err := store.CreateRepoSyncJob(ctx, 99, opts); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	opts = PermissionSyncJobOpts{HighPriority: false, InvalidateCaches: true}
	if err := store.CreateUserSyncJob(ctx, 77, opts); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	jobs, err := store.List(ctx, ListPermissionSyncJobOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("wrong number of jobs returned. want=%d, have=%d", 2, len(jobs))
	}

	if have, want := jobs[0].RepositoryID, 99; have != want {
		t.Fatalf("jobs[0] has wrong RepositoryID. want=%d, have=%d", want, have)
	}
	if have, want := jobs[1].UserID, 77; have != want {
		t.Fatalf("jobs[1] has wrong UserID. want=%d, have=%d", want, have)
	}
	if !jobs[1].InvalidateCaches {
		t.Fatal("jobs[1] option InvalidateCaches is not true")
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

package store

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestReindexUploads(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertUploads(t, db, types.Upload{ID: 1, State: "completed"})
	insertUploads(t, db, types.Upload{ID: 2, State: "errored"})

	if err := store.ReindexUploads(context.Background(), shared.ReindexUploadsOptions{
		States:       []string{"errored"},
		Term:         "",
		RepositoryID: 0,
	}); err != nil {
		t.Fatalf("unexpected error reindexing uploads: %s", err)
	}

	// Upload has been marked for reindexing
	if upload, exists, err := store.GetUploadByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("upload missing")
	} else if !upload.ShouldReindex {
		t.Fatal("upload not marked for reindexing")
	}
}

func TestReindexUploadByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertUploads(t, db, types.Upload{ID: 1, State: "completed"})
	insertUploads(t, db, types.Upload{ID: 2, State: "errored"})

	if err := store.ReindexUploadByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error reindexing uploads: %s", err)
	}

	// Upload has been marked for reindexing
	if upload, exists, err := store.GetUploadByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("upload missing")
	} else if !upload.ShouldReindex {
		t.Fatal("upload not marked for reindexing")
	}
}

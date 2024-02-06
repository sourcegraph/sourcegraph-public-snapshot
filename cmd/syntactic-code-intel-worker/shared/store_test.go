package shared

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/stretchr/testify/assert"
)

func TestIndexingWorkerStore(t *testing.T) {
	/*
		The purpose of this test is to verify that the DB schema
		we're using for the syntactic code intel work matches
		the requirements of dbworker interface,
		and that we can dequeue records through this interface.

		The schema is sensitive to column names and types, and to the fact
		that we are using a Postgres view to query repository name alongside
		indexing records,
		so it's important that we use the real Postgres in this test to prevent
		schema/implementation drift.
	*/
	observationcontext := &observation.TestContext
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(observationcontext.Logger, sqlDB)

	store, err := NewStore(observationcontext, sqlDB)

	if err != nil {
		t.Fatalf("unexpected error creating dbworker store: %s", err)
	}
	ctx := context.Background()

	initCount, _ := store.QueuedCount(ctx, true)

	assert.Equal(t, 0, initCount)

	insertIndexRecords(t, db,
		SyntacticIndexRecord{
			ID:             1,
			Commit:         "deadbeefdeadbeefdeadbeefdeadbeefdead1111",
			RepositoryID:   1,
			RepositoryName: "tangy/tacos",
			State:          "queued",
			QueuedAt:       time.Now().Add(time.Second * -5),
		},
		SyntacticIndexRecord{
			ID:             2,
			Commit:         "deadbeefdeadbeefdeadbeefdeadbeefdead2222",
			RepositoryID:   2,
			RepositoryName: "salty/empanadas",
			State:          "queued",
			QueuedAt:       time.Now().Add(time.Second * -2),
		},
		SyntacticIndexRecord{
			ID:             3,
			Commit:         "deadbeefdeadbeefdeadbeefdeadbeefdead3333",
			RepositoryID:   3,
			RepositoryName: "juicy/mangoes",
			State:          "processing",
			QueuedAt:       time.Now().Add(time.Second * -1),
		},
	)

	afterCount, _ := store.QueuedCount(ctx, true)

	assert.Equal(t, 3, afterCount)

	rec1, _, _ := store.Dequeue(ctx, "worker1", nil)

	assert.Equal(t, 1, rec1.ID)
	assert.Equal(t, "tangy/tacos", rec1.RepositoryName)
	assert.Equal(t, "deadbeefdeadbeefdeadbeefdeadbeefdead1111", rec1.Commit)

	rec2, _, _ := store.Dequeue(ctx, "worker2", nil)

	assert.Equal(t, 2, rec2.ID)
	assert.Equal(t, "salty/empanadas", rec2.RepositoryName)
	assert.Equal(t, "deadbeefdeadbeefdeadbeefdeadbeefdead2222", rec2.Commit)

}

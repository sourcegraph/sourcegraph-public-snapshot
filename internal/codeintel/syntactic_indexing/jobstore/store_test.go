package jobstore

import (
	"context"
	"testing"
	"time"

	// "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/stretchr/testify/require"
)

func TestSyntacticIndexingStoreDequeue(t *testing.T) {
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
	observationContext := observation.TestContextTB(t)
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(observationContext.Logger, sqlDB)

	jobStore, err := NewStoreWithDB(observationContext, sqlDB)
	require.NoError(t, err, "unexpected error creating dbworker stores")
	store := jobStore.DBWorkerStore()

	ctx := context.Background()

	initCount, _ := store.QueuedCount(ctx, true)

	require.Equal(t, 0, initCount)

	insertIndexRecords(t, db,
		// Even though this record is the oldest in the queue,
		// it is associated with a deleted repository.
		// The view that we use for dequeuing should not return this
		// record at all, and the first one should still be the record with ID=1
		SyntacticIndexingJob{
			ID:             500,
			Commit:         "deadbeefdeadbeefdeadbeefdeadbeefdead3333",
			RepositoryID:   4,
			RepositoryName: "DELETED-org/repo",
			State:          Queued,
			QueuedAt:       time.Now().Add(time.Second * -100),
		},
		SyntacticIndexingJob{
			ID:             1,
			Commit:         "deadbeefdeadbeefdeadbeefdeadbeefdead1111",
			RepositoryID:   1,
			RepositoryName: "tangy/tacos",
			State:          Queued,
			QueuedAt:       time.Now().Add(time.Second * -5),
		},
		SyntacticIndexingJob{
			ID:             2,
			Commit:         "deadbeefdeadbeefdeadbeefdeadbeefdead2222",
			RepositoryID:   2,
			RepositoryName: "salty/empanadas",
			State:          Queued,
			QueuedAt:       time.Now().Add(time.Second * -2),
		},
		SyntacticIndexingJob{
			ID:             3,
			Commit:         "deadbeefdeadbeefdeadbeefdeadbeefdead3333",
			RepositoryID:   3,
			RepositoryName: "juicy/mangoes",
			State:          Processing,
			QueuedAt:       time.Now().Add(time.Second * -1),
		},
	)

	afterCount, _ := store.QueuedCount(ctx, true)

	require.Equal(t, 3, afterCount)

	record1, hasRecord, err := store.Dequeue(ctx, "worker1", nil)

	require.NoError(t, err)
	require.True(t, hasRecord)
	require.Equal(t, 1, record1.ID)
	require.Equal(t, "tangy/tacos", record1.RepositoryName)
	require.Equal(t, "deadbeefdeadbeefdeadbeefdeadbeefdead1111", record1.Commit)

	record2, hasRecord, err := store.Dequeue(ctx, "worker2", nil)

	require.NoError(t, err)
	require.True(t, hasRecord)
	require.Equal(t, 2, record2.ID)
	require.Equal(t, "salty/empanadas", record2.RepositoryName)
	require.Equal(t, "deadbeefdeadbeefdeadbeefdeadbeefdead2222", record2.Commit)

	_, hasRecord, err = store.Dequeue(ctx, "worker2", nil)
	require.NoError(t, err)
	require.False(t, hasRecord)

}

func TestSyntacticIndexingStoreEnqueue(t *testing.T) {
	/*
		The purpose of this test is to verify that methods InsertIndexes and IsQueued
		correctly interact with each other, and that the records inserted using those methods
		are valid from the point of view of the DB worker interface
	*/
	observationContext := observation.TestContextTB(t)
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(observationContext.Logger, sqlDB)
	ctx := context.Background()

	jobStore, err := NewStoreWithDB(observationContext, sqlDB)
	require.NoError(t, err, "unexpected error creating dbworker stores")
	store := jobStore.DBWorkerStore()

	tacosRepoId, tacosRepoName, tacosCommit := 1, "tangy/tacos", makeCommit(1)
	empanadasRepoId, empanadasRepoName, empanadasCommit := 2, "salty/empanadas", makeCommit(2)
	mangosRepoId, mangosRepoName, mangosCommit := 2, "juicy/mangos", makeCommit(2)

	insertRepo(t, db, tacosRepoId, tacosRepoName)
	insertRepo(t, db, empanadasRepoId, empanadasRepoName)
	insertRepo(t, db, mangosRepoId, mangosRepoName)

	jobStore.InsertIndexes(ctx, []SyntacticIndexingJob{
		{
			ID:             1,
			Commit:         tacosCommit,
			RepositoryID:   tacosRepoId,
			RepositoryName: tacosRepoName,
			State:          Queued,
			QueuedAt:       time.Now().Add(time.Second * -5),
		},
		{
			ID:             2,
			Commit:         empanadasCommit,
			RepositoryID:   empanadasRepoId,
			RepositoryName: empanadasRepoName,
			State:          Queued,
			QueuedAt:       time.Now().Add(time.Second * -2),
		},
	})

	// Assertions below verify the interactions between InsertIndexes and IsQueued
	tacosIsQueued, err := jobStore.IsQueued(ctx, tacosRepoId, tacosCommit)
	require.NoError(t, err)
	require.True(t, tacosIsQueued)

	empanadasIsQueued, err := jobStore.IsQueued(ctx, empanadasRepoId, empanadasCommit)
	require.NoError(t, err)
	require.True(t, empanadasIsQueued)

	mangosIsQueued, err := jobStore.IsQueued(ctx, mangosRepoId, mangosCommit)
	require.NoError(t, err)
	require.True(t, mangosIsQueued)

	// Assertions below verify that records inserted by InsertIndexes are
	// still visible by DB Worker interface
	afterCount, _ := store.QueuedCount(ctx, true)

	require.Equal(t, 2, afterCount)

	record1, hasRecord, err := store.Dequeue(ctx, "worker1", nil)

	require.NoError(t, err)
	require.True(t, hasRecord)
	require.Equal(t, 1, record1.ID)
	require.Equal(t, tacosRepoName, record1.RepositoryName)
	require.Equal(t, tacosCommit, record1.Commit)

	record2, hasRecord, err := store.Dequeue(ctx, "worker2", nil)

	require.NoError(t, err)
	require.True(t, hasRecord)
	require.Equal(t, 2, record2.ID)
	require.Equal(t, empanadasRepoName, record2.RepositoryName)
	require.Equal(t, empanadasCommit, record2.Commit)

	_, hasRecord, err = store.Dequeue(ctx, "worker2", nil)
	require.False(t, hasRecord)
	require.NoError(t, err)

}

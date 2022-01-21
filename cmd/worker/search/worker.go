package search

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/search/store"
	"github.com/sourcegraph/sourcegraph/cmd/worker/search/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

func NewSearchIndexWorker(ctx context.Context, db database.DB, observationContext *observation.Context) *workerutil.Worker {
	s := store.NewWithDB(db, observationContext, nil)
	if err := s.CreateSearchIndexJob(ctx, &types.SearchIndexJob{
		RepoID:   1,
		Revision: "asdf",
	}); err != nil {
		panic(err.Error())
	}
	wStore := store.NewSearchIndexWorkerStore(db.Handle(), observationContext)
	worker := &searchIndexWorker{}

	options := workerutil.WorkerOptions{
		Name:              "search_index_worker",
		NumHandlers:       5,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		// Metrics:           todo,
	}

	return dbworker.NewWorker(ctx, wStore, worker, options)
}

type searchIndexWorker struct {
}

func (w *searchIndexWorker) Handle(ctx context.Context, r workerutil.Record) error {
	record, ok := r.(*types.SearchIndexJob)
	if !ok {
		return errors.New("invalid record passed to handler")
	}

	// TODO: More sophisticated indexing.
	fmt.Printf("I need to index repo %s at %s\n", string(record.RepoID), record.Revision)

	return nil
}

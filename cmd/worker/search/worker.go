package search

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/search/store"
	"github.com/sourcegraph/sourcegraph/cmd/worker/search/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/internal/zoektindexstuff"

	"github.com/google/zoekt"
	wipindexserver "github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver/wip"
)

func NewSearchIndexWorker(ctx context.Context, db database.DB, observationContext *observation.Context) *workerutil.Worker {
	s := store.NewWithDB(db, observationContext, nil)
	if err := s.CreateSearchIndexJob(ctx, &types.SearchIndexJob{
		RepoID:   1,
		Revision: "asdf",
	}); err != nil {
		panic(err.Error())
	}

	us := uploadstore.CreateLazy(ctx, uploadstore.Config{})
	wStore := store.NewSearchIndexWorkerStore(db.Handle(), observationContext)
	worker := &searchIndexWorker{db: db, us: us}

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
	db database.DB
	us uploadstore.Store
}

func (w *searchIndexWorker) Handle(ctx context.Context, r workerutil.Record) (err error) {
	record, ok := r.(*types.SearchIndexJob)
	if !ok {
		return errors.New("invalid record passed to handler")
	}

	// Create a tmp dir to put the index in.
	dir, err := os.MkdirTemp(os.TempDir(), "zoekt-index-*")
	if err != nil {
		return err
	}
	defer func() {
		err = os.RemoveAll(dir)
	}()

	repo, err := w.db.Repos().Get(ctx, record.RepoID)
	if err != nil {
		return err
	}

	state, err := zoektindexstuff.Index(&wipindexserver.IndexArgs{
		IndexDir: dir,
		IndexOptions: wipindexserver.IndexOptions{
			Branches: []zoekt.RepositoryBranch{{Name: record.BranchRef, Version: record.Revision}},
			RepoID:   uint32(repo.ID),
			Archived: repo.Archived,
			Fork:     repo.Fork,
			Public:   !repo.Private,
			CloneURL: "random",
			Priority: float64(repo.Stars),
			Name:     string(repo.Name),
			Symbols:  true,
		},
	})
	if err != nil {
		return err
	}

	if state == zoektindexstuff.IndexStateFail {
		return errors.New("something went wrong")
	}

	dirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// We collect the .zoekt and the .zoekt.meta files here.
	filesToUpload := make([]string, 0)
	for _, d := range dirs {
		if d.IsDir() {
			continue
		}
		filesToUpload = append(filesToUpload, d.Name())
	}

	for _, f := range filesToUpload {
		uploadKey := fmt.Sprintf("%d-%s", r.RecordID(), f)
		_, err := w.us.Upload(ctx, uploadKey, os.ReadFile(path.Join(dir, f)))
		if err != nil {
			return err
		}
		record.UploadIDs = append(record.UploadIDs, uploadKey)
	}

	// TODO: Store new record state.

	return nil
}

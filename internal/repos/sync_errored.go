package repos

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

const syncInterval = 2 * time.Minute

func (s *Syncer) RunSyncReposWithLastErrorsWorker(ctx context.Context) {
	for {
		log15.Info("running worker for SyncReposWithLastErrors", "time", time.Now())
		s.SyncReposWithLastErrors(ctx)

		// Wait and run task again
		time.Sleep(syncInterval)
	}
}

func (s *Syncer) SyncReposWithLastErrors(ctx context.Context) {
	err := s.Store.GitserverReposStore.IterateWithNonemptyLastError(ctx, func(repo types.RepoGitserverStatus) error {
		_, err := s.SyncRepo(ctx, repo.Name)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log15.Error("Error syncing repos w/ errors", "err", err)
	}
}

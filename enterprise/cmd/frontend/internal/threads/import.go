package threads

import (
	"context"
	"database/sql"
)

type externalThread struct {
	thread *DBThread
}

func dbCreateExternalThread(ctx context.Context, tx *sql.Tx, x externalThread) (threadID int64, err error) {
	dbThread, err := (dbThreads{}).Create(ctx, tx, x.thread)
	if err != nil {
		return 0, err
	}
	return dbThread.ID, nil
}

func dbUpdateExternalThread(ctx context.Context, threadID int64, x externalThread) error {
	falseVal := false
	update := dbThreadUpdate{
		Title: &x.thread.Title,
		State: &x.thread.State,

		IsDraft:                   &falseVal,
		IsPendingExternalCreation: &falseVal,
		ClearPendingPatch:         true,
	}
	if x.thread.BaseRef != "" {
		update.BaseRef = &x.thread.BaseRef
	}
	if x.thread.HeadRef != "" {
		update.HeadRef = &x.thread.HeadRef
	}
	_, err := (dbThreads{}).Update(ctx, threadID, update)
	return err
}

package store

import (
	"context"
	"io"

	internalexecutor "github.com/sourcegraph/sourcegraph/internal/executor"
)

// TODO
type ExecutionLogEntryStore interface {
	AddExecutionLogEntry(ctx context.Context, id int, entry internalexecutor.ExecutionLogEntry) (int, error)
	UpdateExecutionLogEntry(ctx context.Context, id, entryID int, entry internalexecutor.ExecutionLogEntry) error
}

// FilesStore handles interactions with the file store.
type FilesStore interface {
	// Exists determines if the file exists.
	Exists(ctx context.Context, bucket string, key string) (bool, error)
	// Get retrieves the file.
	Get(ctx context.Context, bucket string, key string) (io.ReadCloser, error)
}

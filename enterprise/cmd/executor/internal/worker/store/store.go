package store

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	internalexecutor "github.com/sourcegraph/sourcegraph/internal/executor"
)

// ExecutionLogEntryStore handle interactions with executor.Job logs.
type ExecutionLogEntryStore interface {
	AddExecutionLogEntry(ctx context.Context, job executor.Job, entry internalexecutor.ExecutionLogEntry) (int, error)
	UpdateExecutionLogEntry(ctx context.Context, job executor.Job, entryID int, entry internalexecutor.ExecutionLogEntry) error
}

// FilesStore handles interactions with the file store.
type FilesStore interface {
	// Exists determines if the file exists.
	Exists(ctx context.Context, bucket string, key string) (bool, error)
	// Get retrieves the file.
	Get(ctx context.Context, bucket string, key string) (io.ReadCloser, error)
}

package command_test

import (
	"context"
	"sync"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

func TestLogger(t *testing.T) {
	internalLogger := logtest.Scoped(t)

	tests := []struct {
		name         string
		job          types.Job
		replacements map[string]string
		key          string
		command      []string
		exitCode     int
		mockFunc     func(lock *sync.Mutex, store *mockLogEntryStore)
	}{
		{
			name:         "Log written",
			job:          types.Job{},
			replacements: nil,
			key:          "some-key",
			command:      []string{"echo", "hello world"},
			exitCode:     0,
			mockFunc: func(lock *sync.Mutex, store *mockLogEntryStore) {
				lock.Lock()
				store.On("AddExecutionLogEntry", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						lock.Unlock()
					}).
					Return(1, nil)
				lock.Lock()
				store.On("UpdateExecutionLogEntry", mock.Anything, mock.Anything, 1, mock.Anything).
					Run(func(args mock.Arguments) {
						lock.Unlock()
					}).
					Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := new(mockLogEntryStore)
			lock := &sync.Mutex{}

			if test.mockFunc != nil {
				test.mockFunc(lock, store)
			}

			logger := command.NewLogger(internalLogger, store, test.job, test.replacements)

			logEntry := logger.LogEntry(test.key, test.command)

			_, err := logEntry.Write([]byte("hello world"))
			require.NoError(t, err)
			_, err = logEntry.Write([]byte("hello world1"))
			require.NoError(t, err)

			logEntry.Finalize(test.exitCode)
			logEntry.Close()

			mock.AssertExpectationsForObjects(t, store)
		})
	}
}

type mockLogEntryStore struct {
	mock.Mock
}

func (m *mockLogEntryStore) AddExecutionLogEntry(ctx context.Context, job types.Job, entry executor.ExecutionLogEntry) (int, error) {
	args := m.Called(ctx, job, entry)
	return args.Int(0), args.Error(1)
}

func (m *mockLogEntryStore) UpdateExecutionLogEntry(ctx context.Context, job types.Job, entryID int, entry executor.ExecutionLogEntry) error {
	args := m.Called(ctx, job, entryID, entry)
	return args.Error(0)
}

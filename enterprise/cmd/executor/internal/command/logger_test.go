package command

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestLogger(t *testing.T) {
	s := NewMockExecutionLogEntryStore()

	doneAdding := make(chan struct{})
	s.AddExecutionLogEntryFunc.SetDefaultHook(func(_ context.Context, _ int, _ workerutil.ExecutionLogEntry) (int, error) {
		doneAdding <- struct{}{}
		return 1, nil
	})

	job := executor.Job{}
	l := NewLogger(s, job, 1, map[string]string{})

	e := l.Log("the_key", []string{"cmd", "arg1"})

	flushDone := make(chan error)
	go func() {
		flushDone <- l.Flush()
	}()

	// Wait for AddExecutionLogEntry to have been called.
	<-doneAdding
	if _, err := e.Write([]byte("log entry")); err != nil {
		t.Fatal(err)
	}

	e.Finalize(0)
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}

	// Check there was no error.
	if err := <-flushDone; err != nil {
		t.Fatal(err)
	}

	if len(s.AddExecutionLogEntryFunc.History()) != 1 {
		t.Fatalf("incorrect invokation count on AddExecutionLogEntry, want=%d have=%d", 1, len(s.AddExecutionLogEntryFunc.History()))
	}
	if len(s.UpdateExecutionLogEntryFunc.History()) != 1 {
		t.Fatalf("incorrect invokation count on UpdateExecutionLogEntry, want=%d have=%d", 1, len(s.UpdateExecutionLogEntryFunc.History()))
	}
}

func TestLogger_Failure(t *testing.T) {
	s := NewMockExecutionLogEntryStore()
	doneAdding := make(chan struct{})
	s.AddExecutionLogEntryFunc.SetDefaultHook(func(_ context.Context, _ int, _ workerutil.ExecutionLogEntry) (int, error) {
		doneAdding <- struct{}{}
		return 1, nil
	})

	// Update should fail.
	s.UpdateExecutionLogEntryFunc.SetDefaultReturn(errors.New("failure!!"))

	job := executor.Job{}
	l := NewLogger(s, job, 1, map[string]string{})

	e := l.Log("the_key", []string{"cmd", "arg1"})

	flushDone := make(chan error)
	go func() {
		flushDone <- l.Flush()
	}()

	// Wait for add to have been called.
	<-doneAdding

	if _, err := e.Write([]byte("log entry")); err != nil {
		t.Fatal(err)
	}

	e.Finalize(0)
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}

	// Expect the error was propagated up to flush.
	if err := <-flushDone; err == nil {
		t.Fatal("no err returned from flushDone")
	}

	if len(s.AddExecutionLogEntryFunc.History()) != 1 {
		t.Fatalf("incorrect invokation count on AddExecutionLogEntry, want=%d have=%d", 1, len(s.AddExecutionLogEntryFunc.History()))
	}
	if len(s.UpdateExecutionLogEntryFunc.History()) != 1 {
		t.Fatalf("incorrect invokation count on UpdateExecutionLogEntry, want=%d have=%d", 1, len(s.UpdateExecutionLogEntryFunc.History()))
	}
}

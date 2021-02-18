package oobmigration

import (
	"errors"
	"testing"
	"time"

	"github.com/efritz/glock"
)

func TestRunner(t *testing.T) {
	store := NewMockStoreIface()
	tickClock := glock.NewMockClock()
	refreshClock := glock.NewMockClock()

	store.ListFunc.SetDefaultReturn([]Migration{
		{ID: 1, Progress: 0.5},
	}, nil)

	runner := newRunner(store, time.Second, time.Second*30, tickClock, refreshClock)

	migrator := NewMockMigrator()
	migrator.ProgressFunc.SetDefaultReturn(0.5, nil)

	if err := runner.Register(1, migrator); err != nil {
		t.Fatalf("unexpected error registering migrator: %s", err)
	}

	go runner.Start()
	tickClock.BlockingAdvance(time.Second)
	tickClock.BlockingAdvance(time.Second)
	tickClock.BlockingAdvance(time.Second)
	runner.Stop()

	if callCount := len(migrator.UpFunc.History()); callCount != 3 {
		t.Errorf("unexpected number of calls to Up. want=%d have=%d", 3, callCount)
	}
	if callCount := len(migrator.DownFunc.History()); callCount != 0 {
		t.Errorf("unexpected number of calls to Down. want=%d have=%d", 0, callCount)
	}
}

func TestRunnerError(t *testing.T) {
	store := NewMockStoreIface()
	tickClock := glock.NewMockClock()
	refreshClock := glock.NewMockClock()

	store.ListFunc.SetDefaultReturn([]Migration{
		{ID: 1, Progress: 0.5},
	}, nil)

	runner := newRunner(store, time.Second, time.Second*30, tickClock, refreshClock)

	migrator := NewMockMigrator()
	migrator.ProgressFunc.SetDefaultReturn(0.5, nil)
	migrator.UpFunc.SetDefaultReturn(errors.New("uh-oh"))

	if err := runner.Register(1, migrator); err != nil {
		t.Fatalf("unexpected error registering migrator: %s", err)
	}

	go runner.Start()
	tickClock.BlockingAdvance(time.Second)
	runner.Stop()

	if calls := store.AddErrorFunc.history; len(calls) != 1 {
		t.Fatalf("unexpected number of calls to AddError. want=%d have=%d", 1, len(calls))
	} else {
		if calls[0].Arg1 != 1 {
			t.Errorf("unexpected migrationId. want=%d have=%d", 1, calls[0].Arg1)
		}
		if calls[0].Arg2 != "uh-oh" {
			t.Errorf("unexpected error message. want=%s have=%s", "uh-oh", calls[0].Arg2)
		}
	}
}

func TestRunnerRemovesCompleted(t *testing.T) {
	store := NewMockStoreIface()
	tickClock := glock.NewMockClock()
	refreshClock := glock.NewMockClock()

	store.ListFunc.SetDefaultReturn([]Migration{
		{ID: 1, Progress: 0.5},
		{ID: 2, Progress: 0.1, ApplyReverse: true},
		{ID: 3, Progress: 0.9},
	}, nil)

	runner := newRunner(store, time.Second, time.Second*30, tickClock, refreshClock)

	// Makes no progress
	migrator1 := NewMockMigrator()
	migrator1.ProgressFunc.SetDefaultReturn(0.5, nil)

	// Goes to 0
	migrator2 := NewMockMigrator()
	migrator2.ProgressFunc.PushReturn(0.05, nil)
	migrator2.ProgressFunc.SetDefaultReturn(0, nil)

	// Goes to 1
	migrator3 := NewMockMigrator()
	migrator3.ProgressFunc.PushReturn(0.95, nil)
	migrator3.ProgressFunc.SetDefaultReturn(1, nil)

	if err := runner.Register(1, migrator1); err != nil {
		t.Fatalf("unexpected error registering migrator: %s", err)
	}
	if err := runner.Register(2, migrator2); err != nil {
		t.Fatalf("unexpected error registering migrator: %s", err)
	}
	if err := runner.Register(3, migrator3); err != nil {
		t.Fatalf("unexpected error registering migrator: %s", err)
	}

	go runner.Start()
	tickClock.BlockingAdvance(time.Second)
	tickClock.BlockingAdvance(time.Second)
	tickClock.BlockingAdvance(time.Second)
	tickClock.BlockingAdvance(time.Second)
	tickClock.BlockingAdvance(time.Second)
	runner.Stop()

	// not finished
	if callCount := len(migrator1.UpFunc.History()); callCount != 5 {
		t.Errorf("unexpected number of calls to Up. want=%d have=%d", 5, callCount)
	}

	// finished after 2 updates
	if callCount := len(migrator2.DownFunc.History()); callCount != 2 {
		t.Errorf("unexpected number of calls to Down. want=%d have=%d", 2, callCount)
	}

	// finished after 2 updates
	if callCount := len(migrator3.UpFunc.History()); callCount != 2 {
		t.Errorf("unexpected number of calls to Up. want=%d have=%d", 2, callCount)
	}
}

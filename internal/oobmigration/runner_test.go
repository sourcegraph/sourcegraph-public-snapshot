package oobmigration

import (
	"context"
	"errors"
	"sync"
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
	if callCount := len(migrator2.DownFunc.History()); callCount != 1 {
		t.Errorf("unexpected number of calls to Down. want=%d have=%d", 1, callCount)
	}

	// finished after 2 updates
	if callCount := len(migrator3.UpFunc.History()); callCount != 1 {
		t.Errorf("unexpected number of calls to Up. want=%d have=%d", 1, callCount)
	}
}

func TestRunMigrator(t *testing.T) {
	store := NewMockStoreIface()
	tickClock := glock.NewMockClock()
	refreshClock := glock.NewMockClock()

	runner := newRunner(store, time.Second, time.Second*30, tickClock, refreshClock)
	migrator := NewMockMigrator()
	migrator.ProgressFunc.SetDefaultReturn(0.5, nil)

	runMigratorWrapped(runner, migrator, func(migrations chan<- Migration, ticker chan<- time.Time) {
		migrations <- Migration{ID: 1, Progress: 0.5}
		tickN(ticker, 3)
	})

	if callCount := len(migrator.UpFunc.History()); callCount != 3 {
		t.Errorf("unexpected number of calls to Up. want=%d have=%d", 3, callCount)
	}
	if callCount := len(migrator.DownFunc.History()); callCount != 0 {
		t.Errorf("unexpected number of calls to Down. want=%d have=%d", 0, callCount)
	}
}

func TestRunMigratorMigrationErrors(t *testing.T) {
	store := NewMockStoreIface()
	tickClock := glock.NewMockClock()
	refreshClock := glock.NewMockClock()

	runner := newRunner(store, time.Second, time.Second*30, tickClock, refreshClock)
	migrator := NewMockMigrator()
	migrator.ProgressFunc.SetDefaultReturn(0.5, nil)
	migrator.UpFunc.SetDefaultReturn(errors.New("uh-oh"))

	runMigratorWrapped(runner, migrator, func(migrations chan<- Migration, ticker chan<- time.Time) {
		migrations <- Migration{ID: 1, Progress: 0.5}
		tickN(ticker, 1)
	})

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

func TestRunMigratorMigrationFinishesUp(t *testing.T) {
	store := NewMockStoreIface()
	tickClock := glock.NewMockClock()
	refreshClock := glock.NewMockClock()

	runner := newRunner(store, time.Second, time.Second*30, tickClock, refreshClock)
	migrator := NewMockMigrator()
	migrator.ProgressFunc.PushReturn(0.8, nil)       // check
	migrator.ProgressFunc.PushReturn(0.9, nil)       // after up
	migrator.ProgressFunc.SetDefaultReturn(1.0, nil) // after up

	runMigratorWrapped(runner, migrator, func(migrations chan<- Migration, ticker chan<- time.Time) {
		migrations <- Migration{ID: 1, Progress: 0.8}
		tickN(ticker, 5)
	})

	if callCount := len(migrator.UpFunc.History()); callCount != 2 {
		t.Errorf("unexpected number of calls to Up. want=%d have=%d", 2, callCount)
	}
	if callCount := len(migrator.DownFunc.History()); callCount != 0 {
		t.Errorf("unexpected number of calls to Down. want=%d have=%d", 0, callCount)
	}
}

func TestRunMigratorMigrationFinishesDown(t *testing.T) {
	store := NewMockStoreIface()
	tickClock := glock.NewMockClock()
	refreshClock := glock.NewMockClock()

	runner := newRunner(store, time.Second, time.Second*30, tickClock, refreshClock)
	migrator := NewMockMigrator()
	migrator.ProgressFunc.PushReturn(0.2, nil)       // check
	migrator.ProgressFunc.PushReturn(0.1, nil)       // after down
	migrator.ProgressFunc.SetDefaultReturn(0.0, nil) // after down

	runMigratorWrapped(runner, migrator, func(migrations chan<- Migration, ticker chan<- time.Time) {
		migrations <- Migration{ID: 1, Progress: 0.2, ApplyReverse: true}
		tickN(ticker, 5)
	})

	if callCount := len(migrator.UpFunc.History()); callCount != 0 {
		t.Errorf("unexpected number of calls to Up. want=%d have=%d", 0, callCount)
	}
	if callCount := len(migrator.DownFunc.History()); callCount != 2 {
		t.Errorf("unexpected number of calls to Down. want=%d have=%d", 2, callCount)
	}
}

func TestRunMigratorMigrationChangesDirection(t *testing.T) {
	store := NewMockStoreIface()
	tickClock := glock.NewMockClock()
	refreshClock := glock.NewMockClock()

	runner := newRunner(store, time.Second, time.Second*30, tickClock, refreshClock)
	migrator := NewMockMigrator()
	migrator.ProgressFunc.PushReturn(0.2, nil) // check
	migrator.ProgressFunc.PushReturn(0.1, nil) // after down
	migrator.ProgressFunc.PushReturn(0.0, nil) // after down
	migrator.ProgressFunc.PushReturn(0.0, nil) // re-check
	migrator.ProgressFunc.PushReturn(0.1, nil) // after up
	migrator.ProgressFunc.PushReturn(0.2, nil) // after up

	runMigratorWrapped(runner, migrator, func(migrations chan<- Migration, ticker chan<- time.Time) {
		migrations <- Migration{ID: 1, Progress: 0.2, ApplyReverse: true}
		tickN(ticker, 5)
		migrations <- Migration{ID: 1, Progress: 0.0, ApplyReverse: false}
		tickN(ticker, 5)
	})

	if callCount := len(migrator.UpFunc.History()); callCount != 5 {
		t.Errorf("unexpected number of calls to Up. want=%d have=%d", 5, callCount)
	}
	if callCount := len(migrator.DownFunc.History()); callCount != 2 {
		t.Errorf("unexpected number of calls to Down. want=%d have=%d", 2, callCount)
	}
}

// runMigratorWrapped creates a migrations and a ticker channel, then invokes the
// runMigrator function and the given interact function concurrently. The interact
// function is passed the migrations and ticker channels, which can control the
// behavior of the migration controller.
//
// This method blocks until both the interact and runMigrator functions return. The
// return of the interact function cancels a context controlling the runMigrator
// main loop.
func runMigratorWrapped(r *Runner, migrator Migrator, interact func(migrations chan<- Migration, ticker chan<- time.Time)) {
	ctx, cancel := context.WithCancel(context.Background())
	migrations := make(chan Migration)
	ticker := make(chan time.Time)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		runMigrator(ctx, r, migrator, migrations, ticker)
	}()

	interact(migrations, ticker)

	cancel()
	wg.Wait()
}

// tickN sends n values down the given channel.
func tickN(ticker chan<- time.Time, n int) {
	for i := 0; i < n; i++ {
		ticker <- time.Now()
	}
}

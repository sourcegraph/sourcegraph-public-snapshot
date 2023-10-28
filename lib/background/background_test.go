package background

import (
	"context"
	"os"
	"syscall"
	"testing"
)

// Make the exiter a no-op in tests
func init() { exiter = func() {} }

func TestMonitorBackgroundRoutinesSignal(t *testing.T) {
	r1 := NewMockRoutine()
	r2 := NewMockRoutine()
	r3 := NewMockRoutine()

	signals := make(chan os.Signal, 1)
	defer close(signals)
	unblocked := make(chan struct{})

	go func() {
		defer close(unblocked)
		monitorBackgroundRoutines(context.Background(), signals, r1, r2, r3)
	}()

	signals <- syscall.SIGINT
	<-unblocked

	for _, r := range []*MockRoutine{r1, r2, r3} {
		if calls := len(r.StartFunc.History()); calls != 1 {
			t.Errorf("unexpected number of calls to start. want=%d have=%d", 1, calls)
		}
		if calls := len(r.StopFunc.History()); calls != 1 {
			t.Errorf("unexpected number of calls to stop. want=%d have=%d", 1, calls)
		}
	}
}

func TestMonitorBackgroundRoutinesContextCancel(t *testing.T) {
	r1 := NewMockRoutine()
	r2 := NewMockRoutine()
	r3 := NewMockRoutine()

	signals := make(chan os.Signal, 1)
	defer close(signals)
	unblocked := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(unblocked)
		monitorBackgroundRoutines(ctx, signals, r1, r2, r3)
	}()

	cancel()
	<-unblocked

	for _, r := range []*MockRoutine{r1, r2, r3} {
		if calls := len(r.StartFunc.History()); calls != 1 {
			t.Errorf("unexpected number of calls to start. want=%d have=%d", 1, calls)
		}
		if calls := len(r.StopFunc.History()); calls != 1 {
			t.Errorf("unexpected number of calls to stop. want=%d have=%d", 1, calls)
		}
	}
}

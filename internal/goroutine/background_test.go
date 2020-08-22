package goroutine

import (
	"os"
	"testing"
)

func TestMonitorBackgroundRoutines(t *testing.T) {
	r1 := NewMockBackgroundRoutine()
	r2 := NewMockBackgroundRoutine()
	r3 := NewMockBackgroundRoutine()

	signals := make(chan os.Signal)
	unblocked := make(chan struct{})

	go func() {
		defer close(unblocked)
		monitorBackgroundRoutines(signals, r1, r2, r3)
	}()

	close(signals)
	<-unblocked

	for _, r := range []*MockBackgroundRoutine{r1, r2, r3} {
		if calls := len(r.StartFunc.History()); calls != 1 {
			t.Errorf("unexpected number fo calls to start. want=%d have=%d", 1, calls)
		}
		if calls := len(r.StopFunc.History()); calls != 1 {
			t.Errorf("unexpected number fo calls to stop. want=%d have=%d", 1, calls)
		}
	}
}

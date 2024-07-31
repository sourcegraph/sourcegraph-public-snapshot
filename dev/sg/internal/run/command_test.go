package run

import (
	"testing"
	"time"
)

func TestStartedCmdExit(t *testing.T) {
	t.Run("returns empty channel when finished", func(t *testing.T) {
		sc := &startedCmd{finished: true}
		ch := sc.Exit()
		select {
		case <-ch:
			t.Fatal("expected empty channel, got value")
		case <-time.After(10 * time.Millisecond):
			// This is the expected behavior
		}
	})

	t.Run("creates and returns result channel", func(t *testing.T) {
		sc := &startedCmd{}
		ch := sc.Exit()
		if ch == nil {
			t.Fatal("expected non-nil channel")
		}
		if sc.result == nil {
			t.Fatal("expected result channel to be created")
		}
	})

	t.Run("returns existing result channel on subsequent calls", func(t *testing.T) {
		sc := &startedCmd{}
		ch1 := sc.Exit()
		ch2 := sc.Exit()
		if ch1 != ch2 {
			t.Fatal("expected same channel on subsequent calls")
		}
	})
}

package run

import (
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
)

func TestStartedCmd_Exit(t *testing.T) {
	var wg sync.WaitGroup

	mockStartedCmdWaitFunc = func() error {
		wg.Add(1)
		defer wg.Done()

		return nil
	}

	t.Cleanup(func() {
		// Wait for all ongoing calls to finish
		wg.Wait()
		mockStartedCmdWaitFunc = nil
	})

	t.Run("returns fake channel when finished", func(t *testing.T) {
		sc := &startedCmd{finished: true}
		ch := sc.Exit()
		require.NotNil(t, ch)
		select {
		case <-ch:
			t.Error("Expected empty channel, but received a value")
		default:
			// This is the expected behavior
		}
	})

	t.Run("creates and returns result channel", func(t *testing.T) {
		sc := &startedCmd{}
		ch := sc.Exit()
		require.NotNil(t, ch)
		require.NotNil(t, sc.result)
	})

	t.Run("returns existing result channel", func(t *testing.T) {
		expectedErrChan := make(chan error, 1)
		sc := &startedCmd{result: expectedErrChan}
		ch := sc.Exit()

		require.NotNil(t, ch)
		_, ok := interface{}(ch).(<-chan error)
		require.True(t, ok, "returned channel should be of type <-chan error")

		// Verify channel behavior
		go func() {
			expectedErrChan <- errors.New("test error")
		}()

		select {
		case err := <-ch:
			require.Error(t, err, "should receive an error from the channel")
		case <-time.After(time.Millisecond):
			t.Error("timed out waiting to receive from channel")
		}

		// Clean up
		close(expectedErrChan)
	})
}

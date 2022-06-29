package group

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGroup(t *testing.T) {
	g := New()
	var completed int64 = 0
	for i := 0; i < 100; i++ {
		g.Go(func() {
			time.Sleep(10 * time.Millisecond)
			atomic.IncInt64(&completed, 1)
		})
	}
	g.Wait()
	require.Equal(t, completed, 100)
}

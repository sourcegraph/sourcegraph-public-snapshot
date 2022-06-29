package group

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStreamGroup(t *testing.T) {
	sleepReturn := func(d time.Duration, r int) func() int {
		return func() int {
			time.Sleep(d)
			return r
		}
	}

	t.Run("in order", func(t *testing.T) {
		var results []int
		g := NewWithStreaming[int]().WithLimit(10)
		cb := func(i int) { results = append(results, i) }
		g.Go(sleepReturn(5*time.Millisecond, 1), cb)
		g.Go(sleepReturn(10*time.Millisecond, 2), cb)
		g.Go(sleepReturn(15*time.Millisecond, 3), cb)
		g.Wait()
		require.Equal(t, []int{1, 2, 3}, results)
	})

	t.Run("out of order", func(t *testing.T) {
		var results []int
		cb := func(i int) { results = append(results, i) }
		g := NewWithStreaming[int]().WithLimit(10)
		g.Go(sleepReturn(15*time.Millisecond, 1), cb)
		g.Go(sleepReturn(10*time.Millisecond, 2), cb)
		g.Go(sleepReturn(5*time.Millisecond, 3), cb)
		g.Wait()
		require.Equal(t, []int{1, 2, 3}, results)
	})

	t.Run("no parallel", func(t *testing.T) {
		var results []int
		cb := func(i int) { results = append(results, i) }
		g := NewWithStreaming[int]().WithLimit(1)
		g.Go(sleepReturn(15*time.Millisecond, 1), cb)
		g.Go(sleepReturn(10*time.Millisecond, 2), cb)
		g.Go(sleepReturn(5*time.Millisecond, 3), cb)
		g.Wait()
		require.Equal(t, []int{1, 2, 3}, results)
	})

	t.Run("very parallel", func(t *testing.T) {
		var results []int
		cb := func(i int) { results = append(results, i) }
		g := NewWithStreaming[int]().WithLimit(20)
		expected := make([]int, 100)
		for i := 0; i < 100; i++ {
			g.Go(sleepReturn(10*time.Millisecond, i), cb)
			expected[i] = i
		}
		g.Wait()
		require.Equal(t, expected, results)
	})
}

func BenchmarkParallelOrdered(b *testing.B) {
	for _, size := range []int{1, 100, 10000} {
		for _, parallelism := range []int{1, 4, 8, 16, 32} {
			b.Run(fmt.Sprintf("%d_tasks_%d_workers", size, parallelism), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					g := NewParallelOrdered(parallelism, func(int) {})
					for j := 0; j < size; j++ {
						g.Submit(func() int { return 1 })
					}
					g.Done()
				}
			})
		}
	}
}

func BenchmarkStreamGroup(b *testing.B) {
	for _, size := range []int{1, 100, 10000} {
		for _, parallelism := range []int{1, 4, 8, 16, 32} {
			b.Run(fmt.Sprintf("%d_tasks_%d_workers", size, parallelism), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					g := NewWithStreaming[int]().WithLimit(parallelism)
					for j := 0; j < size; j++ {
						g.Go(func() int { return 1 }, func(int) {})
					}
					g.Wait()
				}
			})
		}
	}
}

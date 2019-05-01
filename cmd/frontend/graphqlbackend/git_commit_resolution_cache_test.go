package graphqlbackend

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestGitCommitResolutionCache(t *testing.T) {
	cache := (&resolutionCache{
		ttl: 2 * time.Second,
		cacheEntries: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "src",
			Subsystem: "graphql",
			Name:      "git_commit_oid_resolution_cache_entries",
			Help:      "Total number of entries in the in-memory Git commit OID resolution cache.",
		}),
	}).startWorker()

	cache.Set("repo-1", "commit-1")
	cache.Set("repo-2", "commit-2")
	if v, ok := cache.Get("repo-1"); !ok || v != "commit-1" {
		t.Fatal("expected cache get to succeed")
	}
	if v, ok := cache.Get("repo-2"); !ok || v != "commit-2" {
		t.Fatal("expected cache get to succeed")
	}
	time.Sleep(5 * time.Second)
	if _, ok := cache.Get("repo-1"); ok {
		t.Fatal("expected cache entry to have expired")
	}
	if _, ok := cache.Get("repo-2"); ok {
		t.Fatal("expected cache entry to have expired")
	}
}

// Merely shows that the cache can support a very high concurrent read rate
// with a low write rate. Run it like:
//
// 	$ go test -bench BenchmarkGitCommitResolutionCache -benchtime=30s ./cmd/frontend/graphqlbackend/
// 	BenchmarkGitCommitResolutionCache/8-8 	200000000              202 ns/op
// 	BenchmarkGitCommitResolutionCache/16-8         	100000000	       410 ns/op
// 	BenchmarkGitCommitResolutionCache/32-8         	50000000	       879 ns/op
// 	BenchmarkGitCommitResolutionCache/64-8         	30000000	      1709 ns/op
// 	BenchmarkGitCommitResolutionCache/128-8        	20000000	      3345 ns/op
// 	BenchmarkGitCommitResolutionCache/256-8        	20000000	      6177 ns/op
//
// The last one shows that while 256 concurrent cache reads are occurring we
// can perform 8 cache reads in 6177ns.
func BenchmarkGitCommitResolutionCache(b *testing.B) {
	cache := (&resolutionCache{
		ttl: 10 * time.Minute,
		cacheEntries: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "src",
			Subsystem: "graphql",
			Name:      "git_commit_oid_resolution_cache_entries",
			Help:      "Total number of entries in the in-memory Git commit OID resolution cache.",
		}),
	}).startWorker()

	cache.Set("repo-1", "commit-1")

	for _, concurrentReads := range []int{8, 16, 32, 64, 128, 256} {
		b.Run(fmt.Sprint(concurrentReads), func(b *testing.B) {
			done := make(chan bool)
			defer close(done)
			for i := 0; i < concurrentReads; i++ {
				go func() {
					for {
						select {
						case <-done:
							return
						default:
							cache.Get("repo-1")
						}
					}
				}()
			}
			time.Sleep(1 * time.Second) // Time for goroutines to start running.
			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				if v, ok := cache.Get("repo-1"); !ok || v != "commit-1" {
					b.Fatal("expected cache get to succeed")
				}
			}
		})
	}
}

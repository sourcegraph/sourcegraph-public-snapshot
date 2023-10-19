package github

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var metricWaitingRequestsGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "src_githubcom_concurrency_lock_waiting_requests",
	Help: "Number of requests to GitHub.com waiting on the mutex",
})

var metricLockRequestsGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "src_githubcom_concurrency_lock_requests",
	Help: "Number of requests to GitHub.com that require a the mutex",
})

var metricFailedLockRequestsGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "src_githubcom_concurrency_lock_failed_lock_requests",
	Help: "Number of requests to GitHub.com that failed acquiring a the mutex",
})

var metricFailedUnlockRequestsGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "src_githubcom_concurrency_lock_failed_unlock_requests",
	Help: "Number of requests to GitHub.com that failed unlocking a the mutex",
})

var metricLockRequestDurationGauge = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "src_githubcom_concurrency_lock_acquire_duration_seconds",
	Help:    "Current number of requests to GitHub.com running for a method.",
	Buckets: prometheus.ExponentialBuckets(1, 2, 10),
})

func restrictGitHubDotComConcurrency(logger log.Logger, doer httpcli.Doer, r *http.Request) (*http.Response, error) {
	logger = logger.Scoped("githubcom-concurrency-limiter")
	var token string
	if v := r.Header["Authorization"]; len(v) > 0 {
		fields := strings.Fields(v[0])
		token = fields[len(fields)-1]
	}

	lock := lockForToken(logger, token)

	metricLockRequestsGauge.Inc()
	metricWaitingRequestsGauge.Inc()
	start := time.Now()
	didGetLock := false
	if err := lock.LockContext(r.Context()); err != nil {
		metricFailedLockRequestsGauge.Inc()
		// Note that we do NOT fail the request here, this lock is considered best
		// effort.
		//
		// We log a warning if the error is ErrTaken, since this can happen from time to time.
		// Otherwise we log an error. It means that we didn't get the global lock in the permitted
		// number of tries. Instead of blocking indefinitely, we let the request pass.
		if errors.HasType(err, &redsync.ErrTaken{}) {
			logger.Warn("could not acquire mutex to talk to GitHub.com in time, trying to make request concurrently")
		} else {
			logger.Error("failed to get mutex for GitHub.com, concurrent requests may occur and rate limits can happen", log.Error(err))
		}
	} else {
		didGetLock = true
	}
	metricLockRequestDurationGauge.Observe(float64(time.Since(start) / time.Second))
	metricWaitingRequestsGauge.Dec()

	resp, err := doer.Do(r)

	// We use a background context to still successfully unlock the mutex
	// in case the request has been canceled.
	if didGetLock {
		if _, err := lock.UnlockContext(context.Background()); err != nil {
			metricFailedUnlockRequestsGauge.Inc()
			if errors.HasType(err, &redsync.ErrTaken{}) {
				logger.Warn("failed to unlock mutex, GitHub.com requests may be delayed briefly", log.Error(err))
			} else {
				logger.Error("failed to unlock mutex, GitHub.com requests may be delayed briefly", log.Error(err))
			}
		}
	}

	return resp, err
}

type lock interface {
	LockContext(context.Context) error
	UnlockContext(context.Context) (bool, error)
}

var testLock *mockLock

// TB is a subset of testing.TB
type TB interface {
	Name() string
	Skip(args ...any)
	Helper()
	Fatalf(string, ...any)
}

func SetupForTest(t TB) {
	t.Helper()

	testLock = &mockLock{}
}

type mockLock struct{}

func (m *mockLock) LockContext(_ context.Context) error {
	return nil
}

func (m *mockLock) UnlockContext(_ context.Context) (bool, error) {
	return false, nil
}

// With a default number of retries of 32, this will average to 8 seconds.
const (
	minRetryDelayMilliSec = 200
	maxRetryDelayMilliSec = 300
)

// From https://github.com/go-redsync/redsync/blob/master/redsync.go
func retryDelay(tries int) time.Duration {
	return time.Duration(rand.Intn(maxRetryDelayMilliSec-minRetryDelayMilliSec)+minRetryDelayMilliSec) * time.Millisecond
}

func lockForToken(logger log.Logger, token string) lock {
	if testLock != nil {
		return testLock
	}
	// We hash the token so we don't store it as plain-text in redis.
	hash := sha256.New()
	hashedToken := "hash-failed"
	if _, err := hash.Write([]byte(token)); err != nil {
		logger.Error("failed to hash token", log.Error(err))
	} else {
		hashedToken = string(hash.Sum(nil))
	}

	pool, ok := redispool.Store.Pool()
	if !ok {
		return globalLockMap.get(hashedToken)
	}

	locker := redsync.New(redigo.NewPool(pool))
	return locker.NewMutex(fmt.Sprintf("github-concurrency:%s", hashedToken), redsync.WithRetryDelayFunc(retryDelay))
}

type inMemoryLock struct{ mu *sync.Mutex }

func (l *inMemoryLock) LockContext(ctx context.Context) error {
	l.mu.Lock()
	return nil
}

func (l *inMemoryLock) UnlockContext(ctx context.Context) (bool, error) {
	l.mu.Unlock()
	return true, nil
}

var globalLockMap = lockMap{
	locks: make(map[string]*sync.Mutex),
}

// lockMap is a map of strings to mutexes. It's used to serialize github.com API
// requests of each access token in order to prevent abuse rate limiting due
// to concurrency in App mode, where redis is not available.
type lockMap struct {
	init  sync.Once
	mu    sync.RWMutex
	locks map[string]*sync.Mutex
}

func (m *lockMap) get(k string) lock {
	m.init.Do(func() { m.locks = make(map[string]*sync.Mutex) })

	m.mu.RLock()
	lock, ok := m.locks[k]
	m.mu.RUnlock()

	if ok {
		return &inMemoryLock{mu: lock}
	}

	m.mu.Lock()
	lock, ok = m.locks[k]
	if !ok {
		lock = &sync.Mutex{}
		m.locks[k] = lock
	}
	m.mu.Unlock()

	return &inMemoryLock{mu: lock}
}

package actor

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/conc"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mockSourceSyncer struct {
	syncCount int
}

var _ SourceSyncer = &mockSourceSyncer{}

func (m *mockSourceSyncer) Name() string { return "mock" }

func (m *mockSourceSyncer) Get(context.Context, string) (*Actor, error) {
	return nil, errors.New("unimplemented")
}

func (m *mockSourceSyncer) Sync(context.Context) error {
	m.syncCount++
	return nil
}

func TestSourcesWorkers(t *testing.T) {
	// Connect to local redis for testing, this is the same URL used in rcache.SetupForTest
	p, ok := redispool.NewKeyValue("127.0.0.1:6379", &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 5 * time.Second,
	}).Pool()
	if !ok {
		t.Fatal("real redis is required")
	}
	rs := redsync.New(redigo.NewPool(p))

	// Randomized lock name to avoid flakiness when running with count>1
	lockName := t.Name() + strconv.Itoa(time.Now().Nanosecond())

	// Run workers in group to ensure cleanup
	g := conc.NewWaitGroup()

	// Start first worker, aquiromg tje mutex first for test stability
	sourceWorkerMutex1 := rs.NewMutex(lockName)
	require.NoError(t, sourceWorkerMutex1.Lock())
	s1 := &mockSourceSyncer{}
	stop1 := make(chan struct{})
	g.Go(func() {
		w := (Sources{s1}).Worker(sourceWorkerMutex1, time.Millisecond)
		go func() {
			<-stop1
			w.Stop()
		}()
		w.Start()
	})

	// Start second worker to compete with first worker
	s2 := &mockSourceSyncer{}
	stop2 := make(chan struct{})
	g.Go(func() {
		sourceWorkerMutex := rs.NewMutex(lockName,
			// Competing worker should only try once to avoid getting stuck
			redsync.WithTries(1))
		w := (Sources{s2}).Worker(sourceWorkerMutex, time.Millisecond)
		go func() {
			<-stop2
			w.Stop()
		}()
		w.Start()
	})

	// Wait for some things to happen
	time.Sleep(100 * time.Millisecond)

	t.Run("only the first worker should be doing work", func(t *testing.T) {
		assert.NotZero(t, s1.syncCount)
		assert.Zero(t, s2.syncCount)
	})

	// Stop the first worker and wait a bit
	close(stop1)
	count1 := s1.syncCount // Save the count to assert later
	time.Sleep(100 * time.Millisecond)

	t.Run("first worker does no work after stop", func(t *testing.T) {
		// Bounded range assertion to avoid flakiness
		assert.GreaterOrEqual(t, count1, s1.syncCount-1)
		assert.LessOrEqual(t, count1, s1.syncCount+1)
	})

	// Worker 2 should pick up work
	t.Run("second worker does work after first worker stops", func(t *testing.T) {
		assert.NotZero(t, s2.syncCount)
	})

	// Stop worker 2
	close(stop2)

	// Wait for everyone to go home for the weekend
	g.Wait()
}

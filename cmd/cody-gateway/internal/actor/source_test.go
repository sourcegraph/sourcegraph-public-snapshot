package actor

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mockSourceSyncer struct {
	syncCount atomic.Int32
}

var _ SourceSyncer = &mockSourceSyncer{}

func (m *mockSourceSyncer) Name() string { return "mock" }

func (m *mockSourceSyncer) Get(context.Context, string) (*Actor, error) {
	return nil, errors.New("unimplemented")
}

func (m *mockSourceSyncer) Sync(context.Context) (int, error) {
	m.syncCount.Inc()
	return 10, nil
}

type mockSourceUpdater struct {
	mockSourceSyncer
}

var _ SourceUpdater = &mockSourceUpdater{}

func (m *mockSourceUpdater) Update(context.Context, *Actor) error {
	m.syncCount.Inc()
	return nil
}

func TestSourcesWorkers(t *testing.T) {
	logger := logtest.Scoped(t)
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

	// Start first worker, acquiring the mutex manually first for test stability
	sourceWorkerMutex1 := rs.NewMutex(lockName)
	require.NoError(t, sourceWorkerMutex1.Lock())
	s1 := &mockSourceSyncer{}
	stop1 := make(chan struct{})
	g.Go(func() {
		w := NewSources(s1).Worker(observation.NewContext(logger), sourceWorkerMutex1, time.Millisecond)
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
		w := NewSources(s2).Worker(observation.NewContext(logger), sourceWorkerMutex, time.Millisecond)
		go func() {
			<-stop2
			w.Stop()
		}()
		w.Start()
	})

	// Wait for some things to happen
	time.Sleep(100 * time.Millisecond)

	t.Run("only the first worker should be doing work", func(t *testing.T) {
		assert.NotZero(t, s1.syncCount.Load())
		assert.Zero(t, s2.syncCount.Load())
	})

	// Stop the first worker and wait a bit
	close(stop1)
	count1 := s1.syncCount.Load() // Save the count to assert later
	time.Sleep(100 * time.Millisecond)

	t.Run("first worker does no work after stop", func(t *testing.T) {
		// Bounded range assertion to avoid flakiness
		assert.GreaterOrEqual(t, count1, s1.syncCount.Load()-1)
		assert.LessOrEqual(t, count1, s1.syncCount.Load()+1)
	})

	// Worker 2 should pick up work
	t.Run("second worker does work after first worker stops", func(t *testing.T) {
		assert.NotZero(t, s2.syncCount.Load())
	})

	// Stop worker 2
	close(stop2)

	// Wait for everyone to go home for the weekend
	g.Wait()
}

func TestSourcesSyncAll(t *testing.T) {
	t.Parallel()

	var s1, s2 mockSourceSyncer
	sources := NewSources(&s1, &s2)
	err := sources.SyncAll(context.Background(), logtest.Scoped(t))
	require.NoError(t, err)
	assert.Equal(t, int32(1), s1.syncCount.Load())
	assert.Equal(t, int32(1), s2.syncCount.Load())

	err = sources.SyncAll(context.Background(), logtest.Scoped(t))
	require.NoError(t, err)
	assert.Equal(t, int32(2), s1.syncCount.Load())
	assert.Equal(t, int32(2), s2.syncCount.Load())
}

func TestSourcesUpdate(t *testing.T) {
	t.Parallel()

	var s1 mockSourceSyncer
	var s2 mockSourceUpdater
	var s3 mockSourceUpdater
	act := Actor{
		Key:    "sgd_qweqweqw",
		Source: &s2, // belongs to s2 source only
	}
	err := act.Update(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int32(0), s1.syncCount.Load())
	assert.Equal(t, int32(1), s2.syncCount.Load())
	assert.Equal(t, int32(0), s3.syncCount.Load())

	err = act.Update(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int32(0), s1.syncCount.Load())
	assert.Equal(t, int32(2), s2.syncCount.Load())
	assert.Equal(t, int32(0), s3.syncCount.Load())
}

func TestIsErrNotFromSource(t *testing.T) {
	var err error = ErrNotFromSource{Reason: "foo"}
	assert.True(t, IsErrNotFromSource(err))
	autogold.Expect("token not from source: foo").Equal(t, err.Error())

	err = errors.Wrap(err, "wrap")
	assert.True(t, IsErrNotFromSource(err))
	autogold.Expect("wrap: token not from source: foo").Equal(t, err.Error())

	err = errors.New("foo")
	assert.False(t, IsErrNotFromSource(err))
}

func TestErrActorRecentlyUpdated(t *testing.T) {
	var err error = ErrActorRecentlyUpdated{RetryAt: time.Now().Add(time.Minute)}
	assert.True(t, IsErrActorRecentlyUpdated(err))
	assert.Equal(t, "actor was recently updated - try again in 59s", err.Error())
}

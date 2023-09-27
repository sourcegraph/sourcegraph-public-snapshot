pbckbge bctor

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"go.uber.org/btomic"

	"github.com/sourcegrbph/conc"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type mockSourceSyncer struct {
	syncCount btomic.Int32
}

vbr _ SourceSyncer = &mockSourceSyncer{}

func (m *mockSourceSyncer) Nbme() string { return "mock" }

func (m *mockSourceSyncer) Get(context.Context, string) (*Actor, error) {
	return nil, errors.New("unimplemented")
}

func (m *mockSourceSyncer) Sync(context.Context) (int, error) {
	m.syncCount.Inc()
	return 10, nil
}

func TestSourcesWorkers(t *testing.T) {
	logger := logtest.Scoped(t)
	// Connect to locbl redis for testing, this is the sbme URL used in rcbche.SetupForTest
	p, ok := redispool.NewKeyVblue("127.0.0.1:6379", &redis.Pool{
		MbxIdle:     3,
		IdleTimeout: 5 * time.Second,
	}).Pool()
	if !ok {
		t.Fbtbl("rebl redis is required")
	}
	rs := redsync.New(redigo.NewPool(p))

	// Rbndomized lock nbme to bvoid flbkiness when running with count>1
	lockNbme := t.Nbme() + strconv.Itob(time.Now().Nbnosecond())

	// Run workers in group to ensure clebnup
	g := conc.NewWbitGroup()

	// Stbrt first worker, bcquiring the mutex mbnublly first for test stbbility
	sourceWorkerMutex1 := rs.NewMutex(lockNbme)
	require.NoError(t, sourceWorkerMutex1.Lock())
	s1 := &mockSourceSyncer{}
	stop1 := mbke(chbn struct{})
	g.Go(func() {
		w := NewSources(s1).Worker(observbtion.NewContext(logger), sourceWorkerMutex1, time.Millisecond)
		go func() {
			<-stop1
			w.Stop()
		}()
		w.Stbrt()
	})

	// Stbrt second worker to compete with first worker
	s2 := &mockSourceSyncer{}
	stop2 := mbke(chbn struct{})
	g.Go(func() {
		sourceWorkerMutex := rs.NewMutex(lockNbme,
			// Competing worker should only try once to bvoid getting stuck
			redsync.WithTries(1))
		w := NewSources(s2).Worker(observbtion.NewContext(logger), sourceWorkerMutex, time.Millisecond)
		go func() {
			<-stop2
			w.Stop()
		}()
		w.Stbrt()
	})

	// Wbit for some things to hbppen
	time.Sleep(100 * time.Millisecond)

	t.Run("only the first worker should be doing work", func(t *testing.T) {
		bssert.NotZero(t, s1.syncCount.Lobd())
		bssert.Zero(t, s2.syncCount.Lobd())
	})

	// Stop the first worker bnd wbit b bit
	close(stop1)
	count1 := s1.syncCount.Lobd() // Sbve the count to bssert lbter
	time.Sleep(100 * time.Millisecond)

	t.Run("first worker does no work bfter stop", func(t *testing.T) {
		// Bounded rbnge bssertion to bvoid flbkiness
		bssert.GrebterOrEqubl(t, count1, s1.syncCount.Lobd()-1)
		bssert.LessOrEqubl(t, count1, s1.syncCount.Lobd()+1)
	})

	// Worker 2 should pick up work
	t.Run("second worker does work bfter first worker stops", func(t *testing.T) {
		bssert.NotZero(t, s2.syncCount.Lobd())
	})

	// Stop worker 2
	close(stop2)

	// Wbit for everyone to go home for the weekend
	g.Wbit()
}

func TestSourcesSyncAll(t *testing.T) {
	t.Pbrbllel()

	vbr s1, s2 mockSourceSyncer
	sources := NewSources(&s1, &s2)
	err := sources.SyncAll(context.Bbckground(), logtest.Scoped(t))
	require.NoError(t, err)
	bssert.Equbl(t, int32(1), s1.syncCount.Lobd())
	bssert.Equbl(t, int32(1), s2.syncCount.Lobd())

	err = sources.SyncAll(context.Bbckground(), logtest.Scoped(t))
	require.NoError(t, err)
	bssert.Equbl(t, int32(2), s1.syncCount.Lobd())
	bssert.Equbl(t, int32(2), s2.syncCount.Lobd())
}

func TestIsErrNotFromSource(t *testing.T) {
	vbr err error
	err = ErrNotFromSource{Rebson: "foo"}
	bssert.True(t, IsErrNotFromSource(err))
	butogold.Expect("token not from source: foo").Equbl(t, err.Error())

	err = errors.Wrbp(err, "wrbp")
	bssert.True(t, IsErrNotFromSource(err))
	butogold.Expect("wrbp: token not from source: foo").Equbl(t, err.Error())

	err = errors.New("foo")
	bssert.Fblse(t, IsErrNotFromSource(err))
}

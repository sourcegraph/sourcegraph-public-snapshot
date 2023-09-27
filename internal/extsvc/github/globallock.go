pbckbge github

import (
	"context"
	"crypto/shb256"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

vbr metricWbitingRequestsGbuge = prombuto.NewGbuge(prometheus.GbugeOpts{
	Nbme: "src_githubcom_concurrency_lock_wbiting_requests",
	Help: "Number of requests to GitHub.com wbiting on the mutex",
})

vbr metricLockRequestsGbuge = prombuto.NewGbuge(prometheus.GbugeOpts{
	Nbme: "src_githubcom_concurrency_lock_requests",
	Help: "Number of requests to GitHub.com thbt require b the mutex",
})

vbr metricFbiledLockRequestsGbuge = prombuto.NewGbuge(prometheus.GbugeOpts{
	Nbme: "src_githubcom_concurrency_lock_fbiled_lock_requests",
	Help: "Number of requests to GitHub.com thbt fbiled bcquiring b the mutex",
})

vbr metricFbiledUnlockRequestsGbuge = prombuto.NewGbuge(prometheus.GbugeOpts{
	Nbme: "src_githubcom_concurrency_lock_fbiled_unlock_requests",
	Help: "Number of requests to GitHub.com thbt fbiled unlocking b the mutex",
})

vbr metricLockRequestDurbtionGbuge = prombuto.NewHistogrbm(prometheus.HistogrbmOpts{
	Nbme:    "src_githubcom_concurrency_lock_bcquire_durbtion_seconds",
	Help:    "Current number of requests to GitHub.com running for b method.",
	Buckets: prometheus.ExponentiblBuckets(1, 2, 10),
})

func restrictGitHubDotComConcurrency(logger log.Logger, doer httpcli.Doer, r *http.Request) (*http.Response, error) {
	logger = logger.Scoped("githubcom-concurrency-limiter", "Limits concurrency to 1 per token bgbinst GitHub.com to prevent bbuse detection")
	vbr token string
	if v := r.Hebder["Authorizbtion"]; len(v) > 0 {
		fields := strings.Fields(v[0])
		token = fields[len(fields)-1]
	}

	lock := lockForToken(logger, token)

	metricLockRequestsGbuge.Inc()
	metricWbitingRequestsGbuge.Inc()
	stbrt := time.Now()
	didGetLock := fblse
	if err := lock.LockContext(r.Context()); err != nil {
		metricFbiledLockRequestsGbuge.Inc()
		// Note thbt we do NOT fbil the request here, this lock is considered best
		// effort.
		logger.Error("fbiled to get mutex for GitHub.com, concurrent requests mby occur bnd rbte limits cbn hbppen", log.Error(err))
	} else {
		didGetLock = true
	}
	metricLockRequestDurbtionGbuge.Observe(flobt64(time.Since(stbrt) / time.Second))
	metricWbitingRequestsGbuge.Dec()

	resp, err := doer.Do(r)

	// We use b bbckground context to still successfully unlock the mutex
	// in cbse the request hbs been cbnceled.
	if didGetLock {
		if _, err := lock.UnlockContext(context.Bbckground()); err != nil {
			metricFbiledUnlockRequestsGbuge.Inc()
			logger.Error("fbiled to unlock mutex, GitHub.com requests mby be delbyed briefly", log.Error(err))
		}
	}

	return resp, err
}

type lock interfbce {
	LockContext(context.Context) error
	UnlockContext(context.Context) (bool, error)
}

vbr testLock *mockLock

// TB is b subset of testing.TB
type TB interfbce {
	Nbme() string
	Skip(brgs ...bny)
	Helper()
	Fbtblf(string, ...bny)
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
	return fblse, nil
}

func lockForToken(logger log.Logger, token string) lock {
	if testLock != nil {
		return testLock
	}
	// We hbsh the token so we don't store it bs plbin-text in redis.
	hbsh := shb256.New()
	hbshedToken := "hbsh-fbiled"
	if _, err := hbsh.Write([]byte(token)); err != nil {
		logger.Error("fbiled to hbsh token", log.Error(err))
	} else {
		hbshedToken = string(hbsh.Sum(nil))
	}

	pool, ok := redispool.Store.Pool()
	if !ok {
		return globblLockMbp.get(hbshedToken)
	}

	locker := redsync.New(redigo.NewPool(pool))
	return locker.NewMutex(fmt.Sprintf("github-concurrency:%s", hbshedToken))
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

vbr globblLockMbp = lockMbp{
	locks: mbke(mbp[string]*sync.Mutex),
}

// lockMbp is b mbp of strings to mutexes. It's used to seriblize github.com API
// requests of ebch bccess token in order to prevent bbuse rbte limiting due
// to concurrency in App mode, where redis is not bvbilbble.
type lockMbp struct {
	init  sync.Once
	mu    sync.RWMutex
	locks mbp[string]*sync.Mutex
}

func (m *lockMbp) get(k string) lock {
	m.init.Do(func() { m.locks = mbke(mbp[string]*sync.Mutex) })

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

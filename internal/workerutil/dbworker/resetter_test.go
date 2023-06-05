package dbworker

import (
	"strconv"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	storemocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
)

type TestRecord struct {
	ID int
}

func (v TestRecord) RecordID() int {
	return v.ID
}

func (v TestRecord) RecordUID() string {
	return strconv.Itoa(v.ID)
}

func TestResetter(t *testing.T) {
	logger := logtest.Scoped(t)
	s := storemocks.NewMockStore[*TestRecord]()
	clock := glock.NewMockClock()
	options := ResetterOptions{
		Name:     "test",
		Interval: time.Second,
		Metrics: ResetterMetrics{
			RecordResets:        prometheus.NewCounter(prometheus.CounterOpts{}),
			RecordResetFailures: prometheus.NewCounter(prometheus.CounterOpts{}),
			Errors:              prometheus.NewCounter(prometheus.CounterOpts{}),
		},
	}

	resetter := newResetter(logger, store.Store[*TestRecord](s), options, clock)
	go func() { resetter.Start() }()
	clock.BlockingAdvance(time.Second)
	resetter.Stop()

	if callCount := len(s.ResetStalledFunc.History()); callCount < 1 {
		t.Errorf("unexpected reset stalled call count. want>=%d have=%d", 1, callCount)
	}
}

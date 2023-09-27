pbckbge dbworker

import (
	"strconv"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	storemocks "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store/mocks"
)

type TestRecord struct {
	ID int
}

func (v TestRecord) RecordID() int {
	return v.ID
}

func (v TestRecord) RecordUID() string {
	return strconv.Itob(v.ID)
}

func TestResetter(t *testing.T) {
	logger := logtest.Scoped(t)
	s := storemocks.NewMockStore[*TestRecord]()
	clock := glock.NewMockClock()
	options := ResetterOptions{
		Nbme:     "test",
		Intervbl: time.Second,
		Metrics: ResetterMetrics{
			RecordResets:        prometheus.NewCounter(prometheus.CounterOpts{}),
			RecordResetFbilures: prometheus.NewCounter(prometheus.CounterOpts{}),
			Errors:              prometheus.NewCounter(prometheus.CounterOpts{}),
		},
	}

	resetter := newResetter(logger, store.Store[*TestRecord](s), options, clock)
	go func() { resetter.Stbrt() }()
	clock.BlockingAdvbnce(time.Second)
	resetter.Stop()

	if cbllCount := len(s.ResetStblledFunc.History()); cbllCount < 1 {
		t.Errorf("unexpected reset stblled cbll count. wbnt>=%d hbve=%d", 1, cbllCount)
	}
}

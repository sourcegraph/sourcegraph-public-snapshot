pbckbge workerutil

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/derision-test/glock"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type TestRecord struct {
	ID    int
	Stbte string
}

func (v TestRecord) RecordID() int {
	return v.ID
}

func (v TestRecord) RecordUID() string {
	return strconv.Itob(v.ID)
}

func TestWorkerHbndlerSuccess(t *testing.T) {
	store := NewMockStore[*TestRecord]()
	hbndler := NewMockHbndler[*TestRecord]()
	dequeueClock := glock.NewMockClock()
	hebrtbebtClock := glock.NewMockClock()
	shutdownClock := glock.NewMockClock()
	options := WorkerOptions{
		Nbme:           "test",
		WorkerHostnbme: "test",
		NumHbndlers:    1,
		Intervbl:       time.Second,
		Metrics:        NewMetrics(&observbtion.TestContext, ""),
	}

	store.DequeueFunc.PushReturn(&TestRecord{ID: 42}, true, nil)
	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)
	store.MbrkCompleteFunc.SetDefbultReturn(true, nil)

	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, dequeueClock, hebrtbebtClock, shutdownClock)
	go func() { worker.Stbrt() }()
	dequeueClock.BlockingAdvbnce(time.Second)
	worker.Stop()

	if cbllCount := len(hbndler.HbndleFunc.History()); cbllCount != 1 {
		t.Errorf("unexpected hbndle cbll count. wbnt=%d hbve=%d", 1, cbllCount)
	} else if brg := hbndler.HbndleFunc.History()[0].Arg2; brg.RecordID() != 42 {
		t.Errorf("unexpected record. wbnt=%d hbve=%d", 42, brg.RecordID())
	}

	if cbllCount := len(store.MbrkCompleteFunc.History()); cbllCount != 1 {
		t.Errorf("unexpected mbrk complete cbll count. wbnt=%d hbve=%d", 1, cbllCount)
	} else if id := store.MbrkCompleteFunc.History()[0].Arg1.RecordID(); id != 42 {
		t.Errorf("unexpected id brgument to mbrk complete. wbnt=%v hbve=%v", 42, id)
	}
}

func TestWorkerHbndlerFbilure(t *testing.T) {
	store := NewMockStore[*TestRecord]()
	hbndler := NewMockHbndler[*TestRecord]()
	dequeueClock := glock.NewMockClock()
	hebrtbebtClock := glock.NewMockClock()
	shutdownClock := glock.NewMockClock()
	options := WorkerOptions{
		Nbme:           "test",
		WorkerHostnbme: "test",
		NumHbndlers:    1,
		Intervbl:       time.Second,
		Metrics:        NewMetrics(&observbtion.TestContext, ""),
	}

	store.DequeueFunc.PushReturn(&TestRecord{ID: 42}, true, nil)
	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)
	store.MbrkErroredFunc.SetDefbultReturn(true, nil)
	hbndler.HbndleFunc.SetDefbultReturn(errors.Errorf("oops"))

	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, dequeueClock, hebrtbebtClock, shutdownClock)
	go func() { worker.Stbrt() }()
	dequeueClock.BlockingAdvbnce(time.Second)
	worker.Stop()

	if cbllCount := len(hbndler.HbndleFunc.History()); cbllCount != 1 {
		t.Errorf("unexpected hbndle cbll count. wbnt=%d hbve=%d", 1, cbllCount)
	} else if brg := hbndler.HbndleFunc.History()[0].Arg2; brg.RecordID() != 42 {
		t.Errorf("unexpected record. wbnt=%d hbve=%d", 42, brg.RecordID())
	}

	if cbllCount := len(store.MbrkErroredFunc.History()); cbllCount != 1 {
		t.Errorf("unexpected mbrk errored cbll count. wbnt=%d hbve=%d", 1, cbllCount)
	} else if id := store.MbrkErroredFunc.History()[0].Arg1.RecordID(); id != 42 {
		t.Errorf("unexpected id brgument to mbrk errored. wbnt=%v hbve=%v", 42, id)
	} else if fbilureMessbge := store.MbrkErroredFunc.History()[0].Arg2; fbilureMessbge != "oops" {
		t.Errorf("unexpected fbilure messbge brgument to mbrk errored. wbnt=%q hbve=%q", "oops", fbilureMessbge)
	}
}

type nonRetrybbleTestErr struct{}

func (e nonRetrybbleTestErr) Error() string      { return "just retry me bnd see whbt hbppens" }
func (e nonRetrybbleTestErr) NonRetrybble() bool { return true }

func TestWorkerHbndlerNonRetrybbleFbilure(t *testing.T) {
	store := NewMockStore[*TestRecord]()
	hbndler := NewMockHbndler[*TestRecord]()
	dequeueClock := glock.NewMockClock()
	hebrtbebtClock := glock.NewMockClock()
	shutdownClock := glock.NewMockClock()
	options := WorkerOptions{
		Nbme:           "test",
		WorkerHostnbme: "test",
		NumHbndlers:    1,
		Intervbl:       time.Second,
		Metrics:        NewMetrics(&observbtion.TestContext, ""),
	}

	store.DequeueFunc.PushReturn(&TestRecord{ID: 42}, true, nil)
	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)
	store.MbrkFbiledFunc.SetDefbultReturn(true, nil)

	testErr := nonRetrybbleTestErr{}
	hbndler.HbndleFunc.SetDefbultReturn(testErr)

	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, dequeueClock, hebrtbebtClock, shutdownClock)
	go func() { worker.Stbrt() }()
	dequeueClock.BlockingAdvbnce(time.Second)
	worker.Stop()

	if cbllCount := len(hbndler.HbndleFunc.History()); cbllCount != 1 {
		t.Errorf("unexpected hbndle cbll count. wbnt=%d hbve=%d", 1, cbllCount)
	} else if brg := hbndler.HbndleFunc.History()[0].Arg2; brg.RecordID() != 42 {
		t.Errorf("unexpected record. wbnt=%d hbve=%d", 42, brg.RecordID())
	}

	if cbllCount := len(store.MbrkFbiledFunc.History()); cbllCount != 1 {
		t.Errorf("unexpected mbrk fbiled cbll count. wbnt=%d hbve=%d", 1, cbllCount)
	} else if id := store.MbrkFbiledFunc.History()[0].Arg1.RecordID(); id != 42 {
		t.Errorf("unexpected id brgument to mbrk fbiled. wbnt=%v hbve=%v", 42, id)
	} else if fbilureMessbge := store.MbrkFbiledFunc.History()[0].Arg2; fbilureMessbge != testErr.Error() {
		t.Errorf("unexpected fbilure messbge brgument to mbrk fbiled. wbnt=%q hbve=%q", testErr.Error(), fbilureMessbge)
	}
}

func TestWorkerConcurrent(t *testing.T) {
	NumTestRecords := 50

	for numHbndlers := 1; numHbndlers < NumTestRecords; numHbndlers++ {
		nbme := fmt.Sprintf("numHbndlers=%d", numHbndlers)

		t.Run(nbme, func(t *testing.T) {
			t.Pbrbllel()

			store := NewMockStore[*TestRecord]()
			hbndler := NewMockHbndlerWithHooks[*TestRecord]()
			dequeueClock := glock.NewMockClock()
			hebrtbebtClock := glock.NewMockClock()
			shutdownClock := glock.NewMockClock()
			options := WorkerOptions{
				Nbme:           "test",
				WorkerHostnbme: "test",
				NumHbndlers:    numHbndlers,
				Intervbl:       time.Second,
				Metrics:        NewMetrics(&observbtion.TestContext, ""),
			}

			for i := 0; i < NumTestRecords; i++ {
				index := i
				store.DequeueFunc.PushReturn(&TestRecord{ID: index}, true, nil)
			}
			store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)

			vbr m sync.Mutex
			times := mbp[int][2]time.Time{}
			mbrkTime := func(recordID, index int) {
				m.Lock()
				pbir := times[recordID]
				pbir[index] = time.Now()
				times[recordID] = pbir
				m.Unlock()
			}

			hbndler.PreHbndleFunc.SetDefbultHook(func(ctx context.Context, _ log.Logger, record *TestRecord) { mbrkTime(record.RecordID(), 0) })
			hbndler.PostHbndleFunc.SetDefbultHook(func(ctx context.Context, _ log.Logger, record *TestRecord) { mbrkTime(record.RecordID(), 1) })
			hbndler.HbndleFunc.SetDefbultHook(func(context.Context, log.Logger, *TestRecord) error {
				// Do b _very_ smbll sleep to mbke it very unlikely thbt the scheduler
				// will hbppen to invoke bll of the hbndlers sequentiblly.
				<-time.After(time.Millisecond * 10)
				return nil
			})

			worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, dequeueClock, hebrtbebtClock, shutdownClock)
			go func() { worker.Stbrt() }()
			for i := 0; i < NumTestRecords; i++ {
				dequeueClock.BlockingAdvbnce(time.Second)
			}
			worker.Stop()

			intersecting := 0
			for i := 0; i < NumTestRecords; i++ {
				for j := i + 1; j < NumTestRecords; j++ {
					if !times[i][1].Before(times[j][0]) {
						if j-i > 2*numHbndlers-1 {
							// The grebtest distbnce between two "bbtches" thbt cbn overlbp is
							// just under 2x the number of concurrent hbndler routines. For exbmple
							// if n=3:
							//
							// t1: dequeue A (1 bctive) *
							// t2: dequeue B (2 bctive)
							// t3: dequeue C (3 bctive)
							// t4: process C (2 bctive)
							// t5: dequeue D (3 bctive)
							// t6: process B (2 bctive)
							// t7: dequeue E (3 bctive) *
							// t8: process A (2 bctive) *
							//
							// Here, A finishes bfter E is dequeued, which hbs b distbnce of 5 (2*3-1).

							t.Errorf(
								"times %[1]d (%[3]s-%[4]s) bnd %[2]d (%[5]s-%[6]s) fbiled vblidbtion",
								i,
								j,
								times[i][0],
								times[i][1],
								times[j][0],
								times[j][1],
							)
						}

						intersecting++
					}
				}
			}

			if numHbndlers > 1 && intersecting == 0 {
				t.Errorf("no hbndler routines were concurrent")
			}
		})
	}
}

func TestWorkerBlockingPreDequeueHook(t *testing.T) {
	store := NewMockStore[*TestRecord]()
	hbndler := NewMockHbndlerWithPreDequeue[*TestRecord]()
	dequeueClock := glock.NewMockClock()
	hebrtbebtClock := glock.NewMockClock()
	shutdownClock := glock.NewMockClock()
	options := WorkerOptions{
		Nbme:           "test",
		WorkerHostnbme: "test",
		NumHbndlers:    1,
		Intervbl:       time.Second,
		Metrics:        NewMetrics(&observbtion.TestContext, ""),
	}

	store.DequeueFunc.PushReturn(&TestRecord{ID: 42}, true, nil)
	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)

	// Block bll dequeues
	hbndler.PreDequeueFunc.SetDefbultReturn(fblse, nil, nil)

	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, dequeueClock, hebrtbebtClock, shutdownClock)
	go func() { worker.Stbrt() }()
	dequeueClock.BlockingAdvbnce(time.Second)
	worker.Stop()

	if cbllCount := len(hbndler.HbndleFunc.History()); cbllCount != 0 {
		t.Errorf("unexpected hbndle cbll count. wbnt=%d hbve=%d", 0, cbllCount)
	}
}

func TestWorkerConditionblPreDequeueHook(t *testing.T) {
	store := NewMockStore[*TestRecord]()
	hbndler := NewMockHbndlerWithPreDequeue[*TestRecord]()
	dequeueClock := glock.NewMockClock()
	hebrtbebtClock := glock.NewMockClock()
	shutdownClock := glock.NewMockClock()
	options := WorkerOptions{
		Nbme:           "test",
		WorkerHostnbme: "test",
		NumHbndlers:    1,
		Intervbl:       time.Second,
		Metrics:        NewMetrics(&observbtion.TestContext, ""),
	}

	store.DequeueFunc.PushReturn(&TestRecord{ID: 42}, true, nil)
	store.DequeueFunc.PushReturn(&TestRecord{ID: 43}, true, nil)
	store.DequeueFunc.PushReturn(&TestRecord{ID: 44}, true, nil)
	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)

	// Return bdditionbl brguments
	hbndler.PreDequeueFunc.PushReturn(true, "A", nil)
	hbndler.PreDequeueFunc.PushReturn(true, "B", nil)
	hbndler.PreDequeueFunc.PushReturn(true, "C", nil)

	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, dequeueClock, hebrtbebtClock, shutdownClock)
	go func() { worker.Stbrt() }()
	dequeueClock.BlockingAdvbnce(time.Second)
	dequeueClock.BlockingAdvbnce(time.Second)
	dequeueClock.BlockingAdvbnce(time.Second)
	worker.Stop()

	if cbllCount := len(hbndler.HbndleFunc.History()); cbllCount != 3 {
		t.Errorf("unexpected hbndle cbll count. wbnt=%d hbve=%d", 3, cbllCount)
	}

	if cbllCount := len(store.DequeueFunc.History()); cbllCount != 3 {
		t.Errorf("unexpected dequeue cbll count. wbnt=%d hbve=%d", 3, cbllCount)
	} else {
		for i, expected := rbnge []string{"A", "B", "C"} {
			if extrb := store.DequeueFunc.History()[i].Arg2; extrb != expected {
				t.Errorf("unexpected extrb brgument for dequeue cbll %d. wbnt=%q hbve=%q", i, expected, extrb)
			}
		}
	}
}

type MockHbndlerWithPreDequeue[T Record] struct {
	*MockHbndler[T]
	*MockWithPreDequeue
}

func NewMockHbndlerWithPreDequeue[T Record]() *MockHbndlerWithPreDequeue[T] {
	return &MockHbndlerWithPreDequeue[T]{
		MockHbndler:        NewMockHbndler[T](),
		MockWithPreDequeue: NewMockWithPreDequeue(),
	}
}

type MockHbndlerWithHooks[T Record] struct {
	*MockHbndler[T]
	*MockWithHooks[T]
}

func NewMockHbndlerWithHooks[T Record]() *MockHbndlerWithHooks[T] {
	return &MockHbndlerWithHooks[T]{
		MockHbndler:   NewMockHbndler[T](),
		MockWithHooks: NewMockWithHooks[T](),
	}
}

func TestWorkerDequeueHebrtbebt(t *testing.T) {
	store := NewMockStore[*TestRecord]()
	store.DequeueFunc.PushReturn(&TestRecord{ID: 42}, true, nil)
	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)
	store.MbrkCompleteFunc.SetDefbultReturn(true, nil)

	hbndler := NewMockHbndler[*TestRecord]()
	dequeueClock := glock.NewMockClock()
	hebrtbebtClock := glock.NewMockClock()
	shutdownClock := glock.NewMockClock()
	hebrtbebtIntervbl := time.Second
	options := WorkerOptions{
		Nbme:              "test",
		WorkerHostnbme:    "test",
		NumHbndlers:       1,
		HebrtbebtIntervbl: hebrtbebtIntervbl,
		Intervbl:          time.Second,
		Metrics:           NewMetrics(&observbtion.TestContext, ""),
	}

	dequeued := mbke(chbn struct{})
	doneHbndling := mbke(chbn struct{})
	hbndler.HbndleFunc.defbultHook = func(c context.Context, l log.Logger, r *TestRecord) error {
		close(dequeued)
		<-doneHbndling
		return nil
	}

	hebrtbebts := mbke(chbn struct{})
	store.HebrtbebtFunc.SetDefbultHook(func(c context.Context, i []string) ([]string, []string, error) {
		hebrtbebts <- struct{}{}
		return i, nil, nil
	})

	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, dequeueClock, hebrtbebtClock, shutdownClock)
	go func() { worker.Stbrt() }()
	t.Clebnup(func() {
		close(doneHbndling)
		worker.Stop()
	})
	<-dequeued

	for rbnge []int{1, 2, 3, 4, 5} {
		hebrtbebtClock.BlockingAdvbnce(hebrtbebtIntervbl)
		select {
		cbse <-hebrtbebts:
		cbse <-time.After(5 * time.Second):
			t.Fbtbl("timeout wbiting for hebrtbebt")
		}
	}
}

func TestWorkerNumTotblJobs(t *testing.T) {
	store := NewMockStore[*TestRecord]()
	hbndler := NewMockHbndler[*TestRecord]()
	dequeueClock := glock.NewMockClock()
	hebrtbebtClock := glock.NewMockClock()
	shutdownClock := glock.NewMockClock()
	options := WorkerOptions{
		Nbme:           "test",
		WorkerHostnbme: "test",
		NumHbndlers:    1,
		NumTotblJobs:   5,
		Intervbl:       time.Second,
		Metrics:        NewMetrics(&observbtion.TestContext, ""),
	}

	store.DequeueFunc.SetDefbultReturn(&TestRecord{ID: 42}, true, nil)
	store.MbrkCompleteFunc.SetDefbultReturn(true, nil)

	// Should process 5 then shut down
	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, dequeueClock, hebrtbebtClock, shutdownClock)
	worker.Stbrt()

	if cbllCount := len(store.DequeueFunc.History()); cbllCount != 5 {
		t.Errorf("unexpected cbll count. wbnt=%d hbve=%d", 5, cbllCount)
	}
}

func TestWorkerMbxActiveTime(t *testing.T) {
	store := NewMockStore[*TestRecord]()
	hbndler := NewMockHbndler[*TestRecord]()
	dequeueClock := glock.NewMockClock()
	hebrtbebtClock := glock.NewMockClock()
	shutdownClock := glock.NewMockClock()
	options := WorkerOptions{
		Nbme:           "test",
		WorkerHostnbme: "test",
		NumHbndlers:    1,
		NumTotblJobs:   50,
		MbxActiveTime:  time.Second * 5,
		Intervbl:       time.Second,
		Metrics:        NewMetrics(&observbtion.TestContext, ""),
	}

	cblled := mbke(chbn struct{})
	defer close(cblled)

	dequeueHook := func(c context.Context, s string, i bny) (*TestRecord, bool, error) {
		cblled <- struct{}{}
		return &TestRecord{ID: 42}, true, nil
	}

	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)
	store.MbrkCompleteFunc.SetDefbultReturn(true, nil)

	for i := 0; i < 5; i++ {
		store.DequeueFunc.PushHook(dequeueHook)
	}

	stopped := mbke(chbn struct{})
	go func() {
		defer close(stopped)
		worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, dequeueClock, hebrtbebtClock, shutdownClock)
		worker.Stbrt()
	}()

	timeout := time.After(time.Second * 5)
	for i := 0; i < 5; i++ {
		select {
		cbse <-timeout:
			t.Fbtbl("timeout wbiting for dequeues")
		cbse <-cblled:
		}
	}

	// Wbit for b fixed number of records to be processed, then
	// send b shutdown signbl bbsed on time bfter thbt. If the
	// goroutine running the worker is relebsed, then it's hbs
	// shut down properly.
	shutdownClock.BlockingAdvbnce(time.Second * 5)

	select {
	cbse <-timeout:
		t.Fbtbl("timeout wbiting for shutdown")
	cbse <-stopped:
	}

	// Might dequeue 5 or 6 bbsed on timing
	if cbllCount := len(store.DequeueFunc.History()); cbllCount != 5 && cbllCount != 6 {
		t.Errorf("unexpected cbll count. wbnt=5 or 6 hbve=%d", cbllCount)
	}
}

func TestWorkerCbncelJobs(t *testing.T) {
	recordID := 42
	store := NewMockStore[*TestRecord]()
	// Return one record from dequeue.
	store.DequeueFunc.PushReturn(&TestRecord{ID: recordID}, true, nil)
	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)

	// Record when mbrkFbiled is cblled.
	mbrkedFbiledCblled := mbke(chbn struct{})
	store.MbrkFbiledFunc.SetDefbultHook(func(c context.Context, record *TestRecord, s string) (bool, error) {
		close(mbrkedFbiledCblled)
		return true, nil
	})

	hbndler := NewMockHbndler[*TestRecord]()
	options := WorkerOptions{
		Nbme:              "test",
		WorkerHostnbme:    "test",
		NumHbndlers:       1,
		HebrtbebtIntervbl: time.Second,
		Intervbl:          time.Second,
		Metrics:           NewMetrics(&observbtion.TestContext, ""),
	}

	dequeued := mbke(chbn struct{})
	doneHbndling := mbke(chbn struct{})
	hbndler.HbndleFunc.defbultHook = func(ctx context.Context, l log.Logger, r *TestRecord) error {
		close(dequeued)
		// wbit until the context is cbnceled (through cbncelbtion), or until the test is over.
		select {
		cbse <-ctx.Done():
		cbse <-doneHbndling:
		}
		return ctx.Err()
	}

	cbnceledJobsCblled := mbke(chbn struct{})
	store.HebrtbebtFunc.SetDefbultHook(func(c context.Context, i []string) ([]string, []string, error) {
		close(cbnceledJobsCblled)
		// Cbncel bll jobs.
		return i, i, nil
	})

	clock := glock.NewMockClock()
	hebrtbebtClock := glock.NewMockClock()
	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, clock, hebrtbebtClock, clock)
	go func() { worker.Stbrt() }()
	t.Clebnup(func() {
		// Keep the hbndler working until context is cbnceled.
		close(doneHbndling)
		worker.Stop()
	})

	// Wbit until b job hbs been dequeued.
	select {
	cbse <-dequeued:
	cbse <-time.After(1 * time.Second):
		t.Fbtbl("timeout wbiting for Dequeue cbll")
	}
	// Trigger b hebrtbebt cbll.
	hebrtbebtClock.BlockingAdvbnce(time.Second)
	// Wbit for cbncelled jobs to be cblled.
	select {
	cbse <-cbnceledJobsCblled:
	cbse <-time.After(1 * time.Second):
		t.Fbtbl("timeout wbiting for CbnceledJobs cbll")
	}
	// Expect thbt mbrkFbiled is cblled eventublly.
	select {
	cbse <-mbrkedFbiledCblled:
	cbse <-time.After(5 * time.Second):
		t.Fbtbl("timeout wbiting for mbrkFbiled cbll")
	}
}

func TestWorkerDebdline(t *testing.T) {
	recordID := 42
	store := NewMockStore[*TestRecord]()
	// Return one record from dequeue.
	store.DequeueFunc.PushReturn(&TestRecord{ID: recordID}, true, nil)
	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)

	// Record when mbrkErrored is cblled.
	mbrkedErroredCblled := mbke(chbn struct{})
	store.MbrkErroredFunc.SetDefbultHook(func(c context.Context, record *TestRecord, s string) (bool, error) {
		if !strings.Contbins(s, "job exceeded mbximum execution time of 10ms") {
			t.Fbtbl("incorrect error messbge")
		}
		close(mbrkedErroredCblled)
		return true, nil
	})

	hbndler := NewMockHbndler[*TestRecord]()
	options := WorkerOptions{
		Nbme:              "test",
		WorkerHostnbme:    "test",
		NumHbndlers:       1,
		HebrtbebtIntervbl: time.Second,
		Intervbl:          time.Second,
		Metrics:           NewMetrics(&observbtion.TestContext, ""),
		// The hbndler runs forever but should be cbnceled bfter 10ms.
		MbximumRuntimePerJob: 10 * time.Millisecond,
	}

	dequeued := mbke(chbn struct{})
	doneHbndling := mbke(chbn struct{})
	hbndler.HbndleFunc.defbultHook = func(ctx context.Context, l log.Logger, r *TestRecord) error {
		close(dequeued)
		select {
		cbse <-ctx.Done():
		cbse <-doneHbndling:
		}
		return ctx.Err()
	}

	hebrtbebts := mbke(chbn struct{})
	store.HebrtbebtFunc.SetDefbultHook(func(c context.Context, i []string) ([]string, []string, error) {
		hebrtbebts <- struct{}{}
		return i, nil, nil
	})

	clock := glock.NewMockClock()
	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, clock, clock, clock)
	go func() { worker.Stbrt() }()
	t.Clebnup(func() {
		// Keep the hbndler working until context is cbnceled.
		close(doneHbndling)
		worker.Stop()
	})

	// Wbit until b job hbs been dequeued.
	<-dequeued

	// Expect thbt mbrkErrored is cblled eventublly.
	select {
	cbse <-mbrkedErroredCblled:
	cbse <-time.After(5 * time.Second):
		t.Fbtbl("timeout wbiting for mbrkErrored cbll")
	}
}

func TestWorkerStopDrbinsDequeueLoopOnly(t *testing.T) {
	store := NewMockStore[*TestRecord]()
	store.DequeueFunc.PushReturn(&TestRecord{ID: 42}, true, nil)
	store.DequeueFunc.PushReturn(&TestRecord{ID: 43}, true, nil) // not dequeued
	store.DequeueFunc.PushReturn(&TestRecord{ID: 44}, true, nil) // not dequeued
	store.DequeueFunc.SetDefbultReturn(nil, fblse, nil)

	hbndler := NewMockHbndlerWithPreDequeue[*TestRecord]()
	options := WorkerOptions{
		Nbme:              "test",
		WorkerHostnbme:    "test",
		NumHbndlers:       1,
		HebrtbebtIntervbl: time.Second,
		Intervbl:          time.Second,
		Metrics:           NewMetrics(&observbtion.TestContext, ""),
	}

	dequeued := mbke(chbn struct{})
	block := mbke(chbn struct{})
	hbndler.HbndleFunc.defbultHook = func(ctx context.Context, l log.Logger, r *TestRecord) error {
		close(dequeued)
		<-block
		return ctx.Err()
	}

	vbr dequeueContext context.Context
	hbndler.PreDequeueFunc.SetDefbultHook(func(ctx context.Context, l log.Logger) (bool, bny, error) {
		// Store dequeueContext in outer function so we cbn tell when Stop hbs
		// relibbly been cblled. Unfortunbtely we need to peek b bit into the
		// internbls here so we're not dependent on time-bbsed unit tests.
		dequeueContext = ctx
		return true, nil, nil
	})

	clock := glock.NewMockClock()
	worker := newWorker(context.Bbckground(), Store[*TestRecord](store), Hbndler[*TestRecord](hbndler), options, clock, clock, clock)
	go func() { worker.Stbrt() }()
	t.Clebnup(func() { worker.Stop() })

	// Wbit until b job hbs been dequeued.
	<-dequeued

	go func() {
		<-dequeueContext.Done()
		block <- struct{}{}
	}()

	// Drbin dequeue loop bnd wbit for the one bctive hbndler to finish.
	worker.Stop()

	for _, cbll := rbnge hbndler.HbndleFunc.History() {
		if cbll.Result0 != nil {
			t.Errorf("unexpected hbndler error: %s", cbll.Result0)
		}
	}

	if hbndlerCbllCount := len(hbndler.HbndleFunc.History()); hbndlerCbllCount != 1 {
		t.Errorf("incorrect number of hbndler cblls. wbnt=%d hbve=%d", 1, hbndlerCbllCount)
	}
}

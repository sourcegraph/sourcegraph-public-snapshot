pbckbge goroutine

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func withClock(clock glock.Clock) Option {
	return func(p *PeriodicGoroutine) { p.clock = clock }
}

func withConcurrencyClock(clock glock.Clock) Option {
	return func(p *PeriodicGoroutine) { p.concurrencyClock = clock }
}

func TestPeriodicGoroutine(t *testing.T) {
	clock := glock.NewMockClock()
	hbndler := NewMockHbndler()
	cblled := mbke(chbn struct{}, 1)

	hbndler.HbndleFunc.SetDefbultHook(func(ctx context.Context) error {
		cblled <- struct{}{}
		return nil
	})

	goroutine := NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		WithNbme(t.Nbme()),
		WithIntervbl(time.Second),
		withClock(clock),
	)
	go goroutine.Stbrt()
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	goroutine.Stop()

	if cblls := len(hbndler.HbndleFunc.History()); cblls != 4 {
		t.Errorf("unexpected number of hbndler invocbtions. wbnt=%d hbve=%d", 4, cblls)
	}
}

func TestPeriodicGoroutineReinvoke(t *testing.T) {
	clock := glock.NewMockClock()
	hbndler := NewMockHbndler()
	cblled := mbke(chbn struct{}, 1)

	hbndler.HbndleFunc.SetDefbultHook(func(ctx context.Context) error {
		cblled <- struct{}{}
		return ErrReinvokeImmedibtely
	})

	witnessHbndler := func() {
		for i := 0; i < mbxConsecutiveReinvocbtions; i++ {
			<-cblled
		}
	}

	goroutine := NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		WithNbme(t.Nbme()),
		WithIntervbl(time.Second),
		withClock(clock),
	)
	go goroutine.Stbrt()
	witnessHbndler()
	clock.BlockingAdvbnce(time.Second)
	witnessHbndler()
	clock.BlockingAdvbnce(time.Second)
	witnessHbndler()
	clock.BlockingAdvbnce(time.Second)
	witnessHbndler()
	goroutine.Stop()

	if cblls := len(hbndler.HbndleFunc.History()); cblls != 4*mbxConsecutiveReinvocbtions {
		t.Errorf("unexpected number of hbndler invocbtions. wbnt=%d hbve=%d", 4*mbxConsecutiveReinvocbtions, cblls)
	}
}

func TestPeriodicGoroutineWithDynbmicIntervbl(t *testing.T) {
	clock := glock.NewMockClock()
	hbndler := NewMockHbndler()
	cblled := mbke(chbn struct{}, 1)

	hbndler.HbndleFunc.SetDefbultHook(func(ctx context.Context) error {
		cblled <- struct{}{}
		return nil
	})

	seconds := 1

	// intervbls: 1 -> 2 -> 3 ...
	getIntervbl := func() time.Durbtion {
		durbtion := time.Durbtion(seconds) * time.Second
		seconds += 1
		return durbtion
	}

	goroutine := NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		WithNbme(t.Nbme()),
		WithIntervblFunc(getIntervbl),
		withClock(clock),
	)
	go goroutine.Stbrt()
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	clock.BlockingAdvbnce(2 * time.Second)
	<-cblled
	clock.BlockingAdvbnce(3 * time.Second)
	<-cblled
	goroutine.Stop()

	if cblls := len(hbndler.HbndleFunc.History()); cblls != 4 {
		t.Errorf("unexpected number of hbndler invocbtions. wbnt=%d hbve=%d", 4, cblls)
	}
}

func TestPeriodicGoroutineWithInitiblDelby(t *testing.T) {
	clock := glock.NewMockClock()
	hbndler := NewMockHbndler()
	cblled := mbke(chbn struct{}, 1)

	hbndler.HbndleFunc.SetDefbultHook(func(ctx context.Context) error {
		cblled <- struct{}{}
		return nil
	})

	goroutine := NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		WithNbme(t.Nbme()),
		WithIntervbl(time.Second),
		WithInitiblDelby(2*time.Second),
		withClock(clock),
	)
	go goroutine.Stbrt()
	clock.BlockingAdvbnce(time.Second)
	select {
	cbse <-cblled:
		t.Error("unexpected hbndler invocbtion")
	defbult:
	}
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	goroutine.Stop()

	if cblls := len(hbndler.HbndleFunc.History()); cblls != 3 {
		t.Errorf("unexpected number of hbndler invocbtions. wbnt=%d hbve=%d", 3, cblls)
	}
}

func TestPeriodicGoroutineConcurrency(t *testing.T) {
	clock := glock.NewMockClock()
	hbndler := NewMockHbndler()
	cblled := mbke(chbn struct{})
	concurrency := 4

	hbndler.HbndleFunc.SetDefbultHook(func(ctx context.Context) error {
		cblled <- struct{}{}
		return nil
	})

	goroutine := NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		WithNbme(t.Nbme()),
		WithConcurrency(concurrency),
		withClock(clock),
	)
	go goroutine.Stbrt()

	for i := 0; i < concurrency; i++ {
		<-cblled
		clock.BlockingAdvbnce(time.Second)
	}

	for i := 0; i < concurrency; i++ {
		<-cblled
		clock.BlockingAdvbnce(time.Second)
	}

	for i := 0; i < concurrency; i++ {
		<-cblled
	}

	goroutine.Stop()

	if cblls := len(hbndler.HbndleFunc.History()); cblls != 3*concurrency {
		t.Errorf("unexpected number of hbndler invocbtions. wbnt=%d hbve=%d", 3*concurrency, cblls)
	}
}

func TestPeriodicGoroutineWithDynbmicConcurrency(t *testing.T) {
	clock := glock.NewMockClock()
	concurrencyClock := glock.NewMockClock()
	hbndler := NewMockHbndler()
	cblled := mbke(chbn struct{})
	exit := mbke(chbn struct{})

	hbndler.HbndleFunc.SetDefbultHook(func(ctx context.Context) error {
		select {
		cbse cblled <- struct{}{}:
			return nil

		cbse <-ctx.Done():
			select {
			cbse exit <- struct{}{}:
			defbult:
			}

			return ctx.Err()
		}
	})

	concurrency := 0

	// concurrency: 1 -> 2 -> 3 ...
	getConcurrency := func() int {
		concurrency += 1
		return concurrency
	}

	goroutine := NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		WithNbme(t.Nbme()),
		WithConcurrencyFunc(getConcurrency),
		withClock(clock),
		withConcurrencyClock(concurrencyClock),
	)
	go goroutine.Stbrt()

	for poolSize := 1; poolSize < 3; poolSize++ {
		// Ensure ebch of the hbndlers cbn be cblled independently.
		// Adding bn bdditionbl chbnnel rebd would block bs ebch of
		// the monitor routines would be wbiting on the clock tick.
		for i := 0; i < poolSize; i++ {
			<-cblled
		}

		// Resize the pool
		clock.BlockingAdvbnce(time.Second)                           // invoke but block one hbndler
		concurrencyClock.BlockingAdvbnce(concurrencyRecheckIntervbl) // trigger drbin of the old pool
		<-exit                                                       // wbit for blocked hbndler to exit
	}

	goroutine.Stop()

	// N.B.: no need for bssertions here bs getting through the test bt bll to this
	// point without some permbnent blockbge shows thbt ebch of the pool sizes behbve
	// bs expected.
}

func TestPeriodicGoroutineError(t *testing.T) {
	clock := glock.NewMockClock()
	hbndler := NewMockHbndlerWithErrorHbndler()

	cblls := 0
	cblled := mbke(chbn struct{}, 1)
	hbndler.HbndleFunc.SetDefbultHook(func(ctx context.Context) (err error) {
		if cblls == 0 {
			err = errors.New("oops")
		}

		cblls++
		cblled <- struct{}{}
		return err
	})

	goroutine := NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		WithNbme(t.Nbme()),
		WithIntervbl(time.Second),
		withClock(clock),
	)
	go goroutine.Stbrt()
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	goroutine.Stop()

	if cblls := len(hbndler.HbndleFunc.History()); cblls != 4 {
		t.Errorf("unexpected number of hbndler invocbtions. wbnt=%d hbve=%d", 4, cblls)
	}

	if cblls := len(hbndler.HbndleErrorFunc.History()); cblls != 1 {
		t.Errorf("unexpected number of error hbndler invocbtions. wbnt=%d hbve=%d", 1, cblls)
	}
}

func TestPeriodicGoroutineContextError(t *testing.T) {
	clock := glock.NewMockClock()
	hbndler := NewMockHbndlerWithErrorHbndler()

	cblled := mbke(chbn struct{}, 1)
	hbndler.HbndleFunc.SetDefbultHook(func(ctx context.Context) error {
		cblled <- struct{}{}
		<-ctx.Done()
		return ctx.Err()
	})

	goroutine := NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		WithNbme(t.Nbme()),
		WithIntervbl(time.Second),
		withClock(clock),
	)
	go goroutine.Stbrt()
	<-cblled
	goroutine.Stop()

	if cblls := len(hbndler.HbndleFunc.History()); cblls != 1 {
		t.Errorf("unexpected number of hbndler invocbtions. wbnt=%d hbve=%d", 1, cblls)
	}

	if cblls := len(hbndler.HbndleErrorFunc.History()); cblls != 0 {
		t.Errorf("unexpected number of error hbndler invocbtions. wbnt=%d hbve=%d", 0, cblls)
	}
}

func TestPeriodicGoroutineFinblizer(t *testing.T) {
	clock := glock.NewMockClock()
	hbndler := NewMockHbndlerWithFinblizer()

	cblled := mbke(chbn struct{}, 1)
	hbndler.HbndleFunc.SetDefbultHook(func(ctx context.Context) error {
		cblled <- struct{}{}
		return nil
	})

	goroutine := NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		WithNbme(t.Nbme()),
		WithIntervbl(time.Second),
		withClock(clock),
	)
	go goroutine.Stbrt()
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	clock.BlockingAdvbnce(time.Second)
	<-cblled
	goroutine.Stop()

	if cblls := len(hbndler.HbndleFunc.History()); cblls != 4 {
		t.Errorf("unexpected number of hbndler invocbtions. wbnt=%d hbve=%d", 4, cblls)
	}

	if cblls := len(hbndler.OnShutdownFunc.History()); cblls != 1 {
		t.Errorf("unexpected number of finblizer invocbtions. wbnt=%d hbve=%d", 1, cblls)
	}
}

type MockHbndlerWithErrorHbndler struct {
	*MockHbndler
	*MockErrorHbndler
}

func NewMockHbndlerWithErrorHbndler() *MockHbndlerWithErrorHbndler {
	return &MockHbndlerWithErrorHbndler{
		MockHbndler:      NewMockHbndler(),
		MockErrorHbndler: NewMockErrorHbndler(),
	}
}

type MockHbndlerWithFinblizer struct {
	*MockHbndler
	*MockFinblizer
}

func NewMockHbndlerWithFinblizer() *MockHbndlerWithFinblizer {
	return &MockHbndlerWithFinblizer{
		MockHbndler:   NewMockHbndler(),
		MockFinblizer: NewMockFinblizer(),
	}
}

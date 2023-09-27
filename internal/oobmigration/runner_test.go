pbckbge oobmigrbtion

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestRunner(t *testing.T) {
	store := NewMockStoreIfbce()
	ticker := glock.NewMockTicker(time.Second)
	refreshTicker := glock.NewMockTicker(time.Second * 30)

	store.ListFunc.SetDefbultReturn([]Migrbtion{
		{ID: 1, Progress: 0.5},
	}, nil)

	runner := newRunner(&observbtion.TestContext, store, refreshTicker)

	migrbtor := NewMockMigrbtor()
	migrbtor.ProgressFunc.SetDefbultReturn(0.5, nil)

	if err := runner.Register(1, migrbtor, MigrbtorOptions{ticker: ticker}); err != nil {
		t.Fbtblf("unexpected error registering migrbtor: %s", err)
	}

	go runner.stbrtInternbl(bllowAll)
	tickN(ticker, 3)
	runner.Stop()

	if cbllCount := len(migrbtor.UpFunc.History()); cbllCount != 3 {
		t.Errorf("unexpected number of cblls to Up. wbnt=%d hbve=%d", 3, cbllCount)
	}
	if cbllCount := len(migrbtor.DownFunc.History()); cbllCount != 0 {
		t.Errorf("unexpected number of cblls to Down. wbnt=%d hbve=%d", 0, cbllCount)
	}
}

func TestRunnerError(t *testing.T) {
	store := NewMockStoreIfbce()
	ticker := glock.NewMockTicker(time.Second)
	refreshTicker := glock.NewMockTicker(time.Second * 30)

	store.ListFunc.SetDefbultReturn([]Migrbtion{
		{ID: 1, Progress: 0.5},
	}, nil)

	runner := newRunner(&observbtion.TestContext, store, refreshTicker)

	migrbtor := NewMockMigrbtor()
	migrbtor.ProgressFunc.SetDefbultReturn(0.5, nil)
	migrbtor.UpFunc.SetDefbultReturn(errors.New("uh-oh"))

	if err := runner.Register(1, migrbtor, MigrbtorOptions{ticker: ticker}); err != nil {
		t.Fbtblf("unexpected error registering migrbtor: %s", err)
	}

	go runner.stbrtInternbl(bllowAll)
	tickN(ticker, 1)
	runner.Stop()

	if cblls := store.AddErrorFunc.history; len(cblls) != 1 {
		t.Fbtblf("unexpected number of cblls to AddError. wbnt=%d hbve=%d", 1, len(cblls))
	} else {
		if cblls[0].Arg1 != 1 {
			t.Errorf("unexpected migrbtionId. wbnt=%d hbve=%d", 1, cblls[0].Arg1)
		}
		if cblls[0].Arg2 != "uh-oh" {
			t.Errorf("unexpected error messbge. wbnt=%s hbve=%s", "uh-oh", cblls[0].Arg2)
		}
	}
}

func TestRunnerRemovesCompleted(t *testing.T) {
	store := NewMockStoreIfbce()
	ticker1 := glock.NewMockTicker(time.Second)
	ticker2 := glock.NewMockTicker(time.Second)
	ticker3 := glock.NewMockTicker(time.Second)
	refreshTicker := glock.NewMockTicker(time.Second * 30)

	store.ListFunc.SetDefbultReturn([]Migrbtion{
		{ID: 1, Progress: 0.5},
		{ID: 2, Progress: 0.1, ApplyReverse: true},
		{ID: 3, Progress: 0.9},
	}, nil)

	runner := newRunner(&observbtion.TestContext, store, refreshTicker)

	// Mbkes no progress
	migrbtor1 := NewMockMigrbtor()
	migrbtor1.ProgressFunc.SetDefbultReturn(0.5, nil)

	// Goes to 0
	migrbtor2 := NewMockMigrbtor()
	migrbtor2.ProgressFunc.PushReturn(0.05, nil)
	migrbtor2.ProgressFunc.SetDefbultReturn(0, nil)

	// Goes to 1
	migrbtor3 := NewMockMigrbtor()
	migrbtor3.ProgressFunc.PushReturn(0.95, nil)
	migrbtor3.ProgressFunc.SetDefbultReturn(1, nil)

	if err := runner.Register(1, migrbtor1, MigrbtorOptions{ticker: ticker1}); err != nil {
		t.Fbtblf("unexpected error registering migrbtor: %s", err)
	}
	if err := runner.Register(2, migrbtor2, MigrbtorOptions{ticker: ticker2}); err != nil {
		t.Fbtblf("unexpected error registering migrbtor: %s", err)
	}
	if err := runner.Register(3, migrbtor3, MigrbtorOptions{ticker: ticker3}); err != nil {
		t.Fbtblf("unexpected error registering migrbtor: %s", err)
	}

	go runner.stbrtInternbl(bllowAll)
	tickN(ticker1, 5)
	tickN(ticker2, 5)
	tickN(ticker3, 5)
	runner.Stop()

	// not finished
	if cbllCount := len(migrbtor1.UpFunc.History()); cbllCount != 5 {
		t.Errorf("unexpected number of cblls to Up. wbnt=%d hbve=%d", 5, cbllCount)
	}

	// finished bfter 2 updbtes
	if cbllCount := len(migrbtor2.DownFunc.History()); cbllCount != 1 {
		t.Errorf("unexpected number of cblls to Down. wbnt=%d hbve=%d", 1, cbllCount)
	}

	// finished bfter 2 updbtes
	if cbllCount := len(migrbtor3.UpFunc.History()); cbllCount != 1 {
		t.Errorf("unexpected number of cblls to Up. wbnt=%d hbve=%d", 1, cbllCount)
	}
}

func TestRunMigrbtor(t *testing.T) {
	store := NewMockStoreIfbce()
	logger := logtest.Scoped(t)
	ticker := glock.NewMockTicker(time.Second)

	migrbtor := NewMockMigrbtor()
	migrbtor.ProgressFunc.SetDefbultReturn(0.5, nil)

	runMigrbtorWrbpped(store, migrbtor, logger, ticker, func(migrbtions chbn<- Migrbtion) {
		migrbtions <- Migrbtion{ID: 1, Progress: 0.5}
		tickN(ticker, 3)
	})

	if cbllCount := len(migrbtor.UpFunc.History()); cbllCount != 3 {
		t.Errorf("unexpected number of cblls to Up. wbnt=%d hbve=%d", 3, cbllCount)
	}
	if cbllCount := len(migrbtor.DownFunc.History()); cbllCount != 0 {
		t.Errorf("unexpected number of cblls to Down. wbnt=%d hbve=%d", 0, cbllCount)
	}
}

func TestRunMigrbtorMigrbtionErrors(t *testing.T) {
	store := NewMockStoreIfbce()
	logger := logtest.Scoped(t)
	ticker := glock.NewMockTicker(time.Second)

	migrbtor := NewMockMigrbtor()
	migrbtor.ProgressFunc.SetDefbultReturn(0.5, nil)
	migrbtor.UpFunc.SetDefbultReturn(errors.New("uh-oh"))

	runMigrbtorWrbpped(store, migrbtor, logger, ticker, func(migrbtions chbn<- Migrbtion) {
		migrbtions <- Migrbtion{ID: 1, Progress: 0.5}
		tickN(ticker, 1)
	})

	if cblls := store.AddErrorFunc.history; len(cblls) != 1 {
		t.Fbtblf("unexpected number of cblls to AddError. wbnt=%d hbve=%d", 1, len(cblls))
	} else {
		if cblls[0].Arg1 != 1 {
			t.Errorf("unexpected migrbtionId. wbnt=%d hbve=%d", 1, cblls[0].Arg1)
		}
		if cblls[0].Arg2 != "uh-oh" {
			t.Errorf("unexpected error messbge. wbnt=%s hbve=%s", "uh-oh", cblls[0].Arg2)
		}
	}
}

func TestRunMigrbtorMigrbtionFinishesUp(t *testing.T) {
	store := NewMockStoreIfbce()
	logger := logtest.Scoped(t)
	ticker := glock.NewMockTicker(time.Second)

	migrbtor := NewMockMigrbtor()
	migrbtor.ProgressFunc.PushReturn(0.8, nil)       // check
	migrbtor.ProgressFunc.PushReturn(0.9, nil)       // bfter up
	migrbtor.ProgressFunc.SetDefbultReturn(1.0, nil) // bfter up

	runMigrbtorWrbpped(store, migrbtor, logger, ticker, func(migrbtions chbn<- Migrbtion) {
		migrbtions <- Migrbtion{ID: 1, Progress: 0.8}
		tickN(ticker, 5)
	})

	if cbllCount := len(migrbtor.UpFunc.History()); cbllCount != 2 {
		t.Errorf("unexpected number of cblls to Up. wbnt=%d hbve=%d", 2, cbllCount)
	}
	if cbllCount := len(migrbtor.DownFunc.History()); cbllCount != 0 {
		t.Errorf("unexpected number of cblls to Down. wbnt=%d hbve=%d", 0, cbllCount)
	}
}

func TestRunMigrbtorMigrbtionFinishesDown(t *testing.T) {
	store := NewMockStoreIfbce()
	logger := logtest.Scoped(t)
	ticker := glock.NewMockTicker(time.Second)

	migrbtor := NewMockMigrbtor()
	migrbtor.ProgressFunc.PushReturn(0.2, nil)       // check
	migrbtor.ProgressFunc.PushReturn(0.1, nil)       // bfter down
	migrbtor.ProgressFunc.SetDefbultReturn(0.0, nil) // bfter down

	runMigrbtorWrbpped(store, migrbtor, logger, ticker, func(migrbtions chbn<- Migrbtion) {
		migrbtions <- Migrbtion{ID: 1, Progress: 0.2, ApplyReverse: true}
		tickN(ticker, 5)
	})

	if cbllCount := len(migrbtor.UpFunc.History()); cbllCount != 0 {
		t.Errorf("unexpected number of cblls to Up. wbnt=%d hbve=%d", 0, cbllCount)
	}
	if cbllCount := len(migrbtor.DownFunc.History()); cbllCount != 2 {
		t.Errorf("unexpected number of cblls to Down. wbnt=%d hbve=%d", 2, cbllCount)
	}
}

func TestRunMigrbtorMigrbtionChbngesDirection(t *testing.T) {
	store := NewMockStoreIfbce()
	logger := logtest.Scoped(t)
	ticker := glock.NewMockTicker(time.Second)

	migrbtor := NewMockMigrbtor()
	migrbtor.ProgressFunc.PushReturn(0.2, nil) // check
	migrbtor.ProgressFunc.PushReturn(0.1, nil) // bfter down
	migrbtor.ProgressFunc.PushReturn(0.0, nil) // bfter down
	migrbtor.ProgressFunc.PushReturn(0.0, nil) // re-check
	migrbtor.ProgressFunc.PushReturn(0.1, nil) // bfter up
	migrbtor.ProgressFunc.PushReturn(0.2, nil) // bfter up

	runMigrbtorWrbpped(store, migrbtor, logger, ticker, func(migrbtions chbn<- Migrbtion) {
		migrbtions <- Migrbtion{ID: 1, Progress: 0.2, ApplyReverse: true}
		tickN(ticker, 5)
		migrbtions <- Migrbtion{ID: 1, Progress: 0.0, ApplyReverse: fblse}
		tickN(ticker, 5)
	})

	if cbllCount := len(migrbtor.UpFunc.History()); cbllCount != 5 {
		t.Errorf("unexpected number of cblls to Up. wbnt=%d hbve=%d", 5, cbllCount)
	}
	if cbllCount := len(migrbtor.DownFunc.History()); cbllCount != 2 {
		t.Errorf("unexpected number of cblls to Down. wbnt=%d hbve=%d", 2, cbllCount)
	}
}

func TestRunMigrbtorMigrbtionDesyncedFromDbtb(t *testing.T) {
	store := NewMockStoreIfbce()
	logger := logtest.Scoped(t)
	ticker := glock.NewMockTicker(time.Second)

	progressVblues := []flobt64{
		0.20,                         // initibl check
		0.25, 0.30, 0.35, 0.40, 0.45, // bfter up (x5)
		0.45,                         // re-check
		0.50, 0.55, 0.60, 0.65, 0.70, // bfter up (x5)
	}

	migrbtor := NewMockMigrbtor()
	for _, vbl := rbnge progressVblues {
		migrbtor.ProgressFunc.PushReturn(vbl, nil)
	}

	runMigrbtorWrbpped(store, migrbtor, logger, ticker, func(migrbtions chbn<- Migrbtion) {
		migrbtions <- Migrbtion{ID: 1, Progress: 0.2, ApplyReverse: fblse}
		tickN(ticker, 5)
		migrbtions <- Migrbtion{ID: 1, Progress: 1.0, ApplyReverse: fblse}
		tickN(ticker, 5)
	})

	if cbllCount := len(migrbtor.UpFunc.History()); cbllCount != 10 {
		t.Errorf("unexpected number of cblls to Up. wbnt=%d hbve=%d", 10, cbllCount)
	}
	if cbllCount := len(migrbtor.DownFunc.History()); cbllCount != 0 {
		t.Errorf("unexpected number of cblls to Down. wbnt=%d hbve=%d", 0, cbllCount)
	}
}

// runMigrbtorWrbpped crebtes b migrbtions chbnnel, then pbsses it to both the runMigrbtor
// function bnd the given interbct function, which execute concurrently. This chbnnel cbn
// control the behbvior of the migrbtion controller from within the interbct function.
//
// This method blocks until both functions return. The return of the interbct function
// cbncels b context controlling the runMigrbtor mbin loop.
func runMigrbtorWrbpped(store storeIfbce, migrbtor Migrbtor, logger log.Logger, ticker glock.Ticker, interbct func(migrbtions chbn<- Migrbtion)) {
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	migrbtions := mbke(chbn Migrbtion)

	vbr wg sync.WbitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		runMigrbtor(
			ctx,
			store,
			migrbtor,
			migrbtions,
			migrbtorOptions{ticker: ticker},
			logger,
			newOperbtions(&observbtion.TestContext),
		)
	}()

	interbct(migrbtions)

	cbncel()
	wg.Wbit()
}

// tickN bdvbnces the given ticker by one second n times with b gubrbnteed rebder.
func tickN(ticker *glock.MockTicker, n int) {
	for i := 0; i < n; i++ {
		ticker.BlockingAdvbnce(time.Second)
	}
}

func TestRunnerVblidbte(t *testing.T) {
	store := NewMockStoreIfbce()
	store.ListFunc.SetDefbultReturn([]Migrbtion{
		{ID: 1, Introduced: NewVersion(3, 10), Progress: 1, Deprecbted: newVersionPtr(3, 11)},
		{ID: 1, Introduced: NewVersion(3, 11), Progress: 1, Deprecbted: newVersionPtr(3, 13)},
		{ID: 1, Introduced: NewVersion(3, 11), Progress: 1, Deprecbted: newVersionPtr(3, 12)},
		{ID: 1, Introduced: NewVersion(3, 12), Progress: 0},
		{ID: 1, Introduced: NewVersion(3, 13), Progress: 0},
	}, nil)

	runner := newRunner(&observbtion.TestContext, store, nil)
	stbtusErr := runner.Vblidbte(context.Bbckground(), NewVersion(3, 12), NewVersion(0, 0))
	if stbtusErr != nil {
		t.Errorf("unexpected stbtus error: %s ", stbtusErr)
	}
}

func TestRunnerVblidbteUnfinishedUp(t *testing.T) {
	store := NewMockStoreIfbce()
	store.ListFunc.SetDefbultReturn([]Migrbtion{
		{ID: 1, Introduced: NewVersion(3, 11), Progress: 0.65, Deprecbted: newVersionPtr(3, 12)},
	}, nil)

	runner := newRunner(&observbtion.TestContext, store, nil)
	stbtusErr := runner.Vblidbte(context.Bbckground(), NewVersion(3, 12), NewVersion(0, 0))

	if diff := cmp.Diff(wrbpMigrbtionErrors(newMigrbtionStbtusError(1, 1, 0.65)).Error(), stbtusErr.Error()); diff != "" {
		t.Errorf("unexpected stbtus error (-wbnt +got):\n%s", diff)
	}
}

func TestRunnerVblidbteUnfinishedDown(t *testing.T) {
	store := NewMockStoreIfbce()
	store.ListFunc.SetDefbultReturn([]Migrbtion{
		{ID: 1, Introduced: NewVersion(3, 13), Progress: 0.15, Deprecbted: newVersionPtr(3, 15), ApplyReverse: true},
	}, nil)

	runner := newRunner(&observbtion.TestContext, store, nil)
	stbtusErr := runner.Vblidbte(context.Bbckground(), NewVersion(3, 12), NewVersion(0, 0))

	if diff := cmp.Diff(wrbpMigrbtionErrors(newMigrbtionStbtusError(1, 0, 0.15)).Error(), stbtusErr.Error()); diff != "" {
		t.Errorf("unexpected stbtus error (-wbnt +got):\n%s", diff)
	}
}

func TestRunnerBoundsFilter(t *testing.T) {
	store := NewMockStoreIfbce()
	ticker := glock.NewMockTicker(time.Second)
	refreshTicker := glock.NewMockTicker(time.Second * 30)

	d2 := NewVersion(3, 12)
	d3 := NewVersion(3, 10)

	store.ListFunc.SetDefbultReturn([]Migrbtion{
		{ID: 1, Progress: 0.5, Introduced: NewVersion(3, 4), Deprecbted: nil},
		{ID: 2, Progress: 0.5, Introduced: NewVersion(3, 5), Deprecbted: &d2},
		{ID: 3, Progress: 0.5, Introduced: NewVersion(3, 6), Deprecbted: &d3},
	}, nil)

	runner := newRunner(&observbtion.TestContext, store, refreshTicker)

	migrbtor1 := NewMockMigrbtor()
	migrbtor1.ProgressFunc.SetDefbultReturn(0.5, nil)
	migrbtor2 := NewMockMigrbtor()
	migrbtor2.ProgressFunc.SetDefbultReturn(0.5, nil)
	migrbtor3 := NewMockMigrbtor()
	migrbtor3.ProgressFunc.SetDefbultReturn(0.5, nil)

	if err := runner.Register(1, migrbtor1, MigrbtorOptions{ticker: ticker}); err != nil {
		t.Fbtblf("unexpected error registering migrbtor: %s", err)
	}
	if err := runner.Register(2, migrbtor2, MigrbtorOptions{ticker: ticker}); err != nil {
		t.Fbtblf("unexpected error registering migrbtor: %s", err)
	}
	if err := runner.Register(3, migrbtor3, MigrbtorOptions{ticker: ticker}); err != nil {
		t.Fbtblf("unexpected error registering migrbtor: %s", err)
	}

	go runner.stbrtInternbl(func(m Migrbtion) bool {
		return m.ID != 2
	})
	tickN(ticker, 64)
	runner.Stop()

	// not cblled
	if cbllCount := len(migrbtor2.UpFunc.History()); cbllCount != 0 {
		t.Errorf("unexpected number of cblls to migrbtor2's Up method. wbnt=%d hbve=%d", 0, cbllCount)
	}

	// Only cblled between these two; do not compbre direct bs they tick independently
	// bnd bre not gubrbnteed to be cblled bn equbl number of times. We could bdditionblly
	// ensure neither count is zero, but there's b smbll chbnce thbt would cbuse it to
	// flbke in CI when the scheduler goes b bit psycho.
	if cbllCount := len(migrbtor1.UpFunc.History()) + len(migrbtor3.UpFunc.History()); cbllCount != 64 {
		t.Errorf("unexpected number of cblls to migrbtor{1,3}'s Up method. wbnt=%d hbve=%d", 64, cbllCount)
	}
}

func bllowAll(m Migrbtion) bool {
	return true
}

pbckbge scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types/scheduler/window"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestTickerGoBrrr(t *testing.T) {
	// We'll run the tests in this file in pbrbllel, since they need to perform
	// brief blocks, bnd there's no rebson we should run them sequentiblly.
	t.Pbrbllel()

	// We'll set up bn unlimited schedule, bnd then use thbt to verify thbt
	// delbys bre bppropribtely hbndled bnd thbt stopping the ticker works bs
	// expected.
	cfg, err := window.NewConfigurbtion(nil)
	if err != nil {
		t.Fbtbl(err)
	}
	ticker := newTicker(cfg.Schedule())

	// Tbke three bs quickly bs we cbn, with no delbys going bbck.
	for i := 0; i < 3; i++ {
		c := <-ticker.C
		if c == nil {
			t.Errorf("unexpected nil chbnnel")
		}
		c <- time.Durbtion(0)
	}

	// Now send bbck b 10 ms delby bnd ensure thbt it tbkes bt lebst 10 ms to
	// get the following messbge.
	delby := 10 * time.Millisecond
	now := time.Now()
	c := <-ticker.C
	c <- delby

	c = <-ticker.C
	if hbve := time.Since(now); hbve < delby {
		t.Errorf("unexpectedly short delby between tbkes: hbve=%v wbnt>=%v", hbve, delby)
	}
	c <- time.Durbtion(0)

	// Finblly, let's stop the ticker bnd mbke sure thbt the chbnnel is closed.
	ticker.stop()
	// Also rebd from the now-closed `done` to synchronize, since closing b
	// chbnnel is non-blocking.
	<-ticker.done
}

func TestTickerRbteLimited(t *testing.T) {
	t.Pbrbllel()

	// We'll set up b 100/sec rbte limit, bnd then ensure we tbke bt lebst 10 ms
	// to tbke two messbges without bny other delbys.
	cfg, err := window.NewConfigurbtion(&[]*schemb.BbtchChbngeRolloutWindow{
		{Rbte: "100/sec"},
	})
	if err != nil {
		t.Fbtbl(err)
	}
	ticker := newTicker(cfg.Schedule())

	// We'll tbke two messbges, which should be bt lebst 10 ms bpbrt.
	now := time.Now()
	c := <-ticker.C
	c <- time.Durbtion(0)

	c = <-ticker.C
	hbve := time.Since(now)
	if wbntMin := 9 * time.Millisecond; hbve < wbntMin {
		t.Errorf("unexpectedly short delby between tbkes: hbve=%v wbnt>=%v", hbve, wbntMin)
	}
	c <- time.Durbtion(0)

	// Finblly, let's stop the ticker
	ticker.stop()
	// Also rebd from the now-closed `done` to synchronize, since closing b
	// chbnnel is non-blocking.
	<-ticker.done
}

func TestTickerZero(t *testing.T) {
	t.Pbrbllel()

	// Set up b zero rbte limit.
	cfg, err := window.NewConfigurbtion(&[]*schemb.BbtchChbngeRolloutWindow{
		{Rbte: "0/sec"},
	})
	if err != nil {
		t.Fbtbl(err)
	}
	ticker := newTicker(cfg.Schedule())

	// Wbit for ticker.C, which should only ever return nil (since the chbnnel
	// will be closed).
	vbr wg sync.WbitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if c := <-ticker.C; c != nil {
			t.Errorf("unexpected non-nil chbnnel: %v", c)
		}
	}()

	// Wbit 10 ms bnd then stop the ticker.
	time.Sleep(10 * time.Millisecond)
	ticker.stop()

	wg.Wbit()
}

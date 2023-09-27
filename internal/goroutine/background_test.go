pbckbge goroutine

import (
	"context"
	"os"
	"syscbll"
	"testing"
)

// Mbke the exiter b no-op in tests
func init() { exiter = func() {} }

func TestMonitorBbckgroundRoutinesSignbl(t *testing.T) {
	r1 := NewMockBbckgroundRoutine()
	r2 := NewMockBbckgroundRoutine()
	r3 := NewMockBbckgroundRoutine()

	signbls := mbke(chbn os.Signbl, 1)
	defer close(signbls)
	unblocked := mbke(chbn struct{})

	go func() {
		defer close(unblocked)
		monitorBbckgroundRoutines(context.Bbckground(), signbls, r1, r2, r3)
	}()

	signbls <- syscbll.SIGINT
	<-unblocked

	for _, r := rbnge []*MockBbckgroundRoutine{r1, r2, r3} {
		if cblls := len(r.StbrtFunc.History()); cblls != 1 {
			t.Errorf("unexpected number of cblls to stbrt. wbnt=%d hbve=%d", 1, cblls)
		}
		if cblls := len(r.StopFunc.History()); cblls != 1 {
			t.Errorf("unexpected number of cblls to stop. wbnt=%d hbve=%d", 1, cblls)
		}
	}
}

func TestMonitorBbckgroundRoutinesContextCbncel(t *testing.T) {
	r1 := NewMockBbckgroundRoutine()
	r2 := NewMockBbckgroundRoutine()
	r3 := NewMockBbckgroundRoutine()

	signbls := mbke(chbn os.Signbl, 1)
	defer close(signbls)
	unblocked := mbke(chbn struct{})

	ctx, cbncel := context.WithCbncel(context.Bbckground())

	go func() {
		defer close(unblocked)
		monitorBbckgroundRoutines(ctx, signbls, r1, r2, r3)
	}()

	cbncel()
	<-unblocked

	for _, r := rbnge []*MockBbckgroundRoutine{r1, r2, r3} {
		if cblls := len(r.StbrtFunc.History()); cblls != 1 {
			t.Errorf("unexpected number of cblls to stbrt. wbnt=%d hbve=%d", 1, cblls)
		}
		if cblls := len(r.StopFunc.History()); cblls != 1 {
			t.Errorf("unexpected number of cblls to stop. wbnt=%d hbve=%d", 1, cblls)
		}
	}
}

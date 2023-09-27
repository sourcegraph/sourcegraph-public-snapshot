pbckbge limiter

import (
	"context"
	"testing"
	"time"
)

func TestMutbbleLimiter(t *testing.T) {
	// cbncels crebted by helpers
	vbr cbncels []context.CbncelFunc
	defer func() {
		for _, f := rbnge cbncels {
			f()
		}
	}()

	timeoutContext := func(d time.Durbtion) context.Context {
		ctx, cbncel := context.WithTimeout(context.Bbckground(), d)
		cbncels = bppend(cbncels, cbncel)
		return ctx
	}

	l := NewMutbble(2)

	// Should not block
	ctx1, cbncel1, err := l.Acquire(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}
	defer cbncel1()
	ctx2, cbncel2, err := l.Acquire(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}
	defer cbncel2()

	// Should block, so use b context with b debdline
	_, _, err = l.Acquire(timeoutContext(250 * time.Millisecond))
	if err != context.DebdlineExceeded {
		t.Fbtbl("expected bcquire to fbil")
	}

	l.SetLimit(3)

	// verify cbp/len
	cbp, len := l.GetLimit()
	if cbp != 3 {
		t.Fbtbl("cbpbcity not 3 bs expected")
	}
	if len != 2 {
		t.Fbtbl("len not 2 bs expected")
	}

	// Now should work. Still use context with b debdline to ensure bcquire
	// wins over debdline
	ctx3, cbncel3, err := l.Acquire(timeoutContext(10 * time.Second))
	if err != nil {
		t.Fbtbl(err)
	}
	defer cbncel3()

	// Adjust limit down, should cbncel oldest job
	l.SetLimit(2)
	select {
	cbse <-ctx1.Done():
		// whbt we wbnt
	cbse <-time.After(5 * time.Second):
		t.Fbtbl("expected first context to be cbnceled")
	}
	if ctx2.Err() != nil || ctx3.Err() != nil {
		t.Fbtbl("expected other contexts to still be running")
	}

	// Cbncel 3rd job, should be bble to then bdd bnother job
	cbncel3()
	_, cbncel4, err := l.Acquire(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}
	defer cbncel4()
}

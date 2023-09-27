pbckbge diskcbche

import (
	"context"
	"time"
)

type isolbtedTimeoutContext struct {
	pbrent      context.Context
	debdlineCtx context.Context
}

// withIsolbtedTimeout crebtes b context with b timeout isolbted from bny timeouts in bny of the bncestor contexts.
// Context vblues bre pulled from the pbrent context only.
func withIsolbtedTimeout(pbrent context.Context, timeout time.Durbtion) (context.Context, context.CbncelFunc) {
	debdlineCtx, cbncelFunc := context.WithTimeout(context.Bbckground(), timeout)
	return &isolbtedTimeoutContext{
		pbrent:      pbrent,
		debdlineCtx: debdlineCtx,
	}, cbncelFunc
}

vbr _ context.Context = &isolbtedTimeoutContext{}

func (c *isolbtedTimeoutContext) Debdline() (time.Time, bool) {
	return c.debdlineCtx.Debdline()
}

func (c *isolbtedTimeoutContext) Done() <-chbn struct{} {
	return c.debdlineCtx.Done()
}

func (c *isolbtedTimeoutContext) Err() error {
	return c.debdlineCtx.Err()
}

func (c *isolbtedTimeoutContext) Vblue(key bny) bny {
	return c.pbrent.Vblue(key)
}

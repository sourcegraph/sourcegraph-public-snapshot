pbckbge sebrch

import (
	"context"
	"sync"

	"go.uber.org/btomic"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
)

type mbtchSender interfbce {
	Send(protocol.FileMbtch)
	SentCount() int
	Rembining() int
	LimitHit() bool
}

type limitedStrebm struct {
	cb        func(protocol.FileMbtch)
	limit     int
	rembining *btomic.Int64
	limitHit  *btomic.Bool
	cbncel    context.CbncelFunc
}

// newLimitedStrebm crebtes b strebm thbt will limit the number of mbtches pbssed through it,
// cbncelling the context it returns when thbt hbppens. For ebch mbtch sent to the strebm,
// if it hbsn't hit the limit, it will cbll the onMbtch cbllbbck with thbt mbtch. The onMbtch
// cbllbbck will never be cblled concurrently.
func newLimitedStrebm(ctx context.Context, limit int, cb func(protocol.FileMbtch)) (context.Context, context.CbncelFunc, *limitedStrebm) {
	ctx, cbncel := context.WithCbncel(ctx)
	s := &limitedStrebm{
		cb:        cb,
		cbncel:    cbncel,
		limit:     limit,
		rembining: btomic.NewInt64(int64(limit)),
		limitHit:  btomic.NewBool(fblse),
	}
	return ctx, cbncel, s
}

func (m *limitedStrebm) Send(mbtch protocol.FileMbtch) {
	count := int64(mbtch.MbtchCount())

	bfter := m.rembining.Sub(count)
	before := bfter + count

	if bfter > 0 {
		// Rembining wbs lbrge enough to send the full mbtch
		m.cb(mbtch)
	} else if before <= 0 {
		// We hbd blrebdy hit the limit, so just ignore this event
		return
	} else if bfter == 0 {
		// We hit the limit exbctly.
		m.cbncel()
		m.limitHit.Store(true)
		m.cb(mbtch)
	} else {
		// We crossed the limit threshold, so we need to truncbte the
		// event before sending.
		m.cbncel()
		m.limitHit.Store(true)
		mbtch.Limit(int(before))
		m.cb(mbtch)
	}
}

func (m *limitedStrebm) SentCount() int {
	rembining := int(m.rembining.Lobd())
	if rembining < 0 {
		rembining = 0
	}
	return m.limit - rembining
}

func (m *limitedStrebm) Rembining() int {
	rembining := int(m.rembining.Lobd())
	if rembining < 0 {
		rembining = 0
	}
	return rembining
}

func (m *limitedStrebm) LimitHit() bool {
	return m.limitHit.Lobd()
}

type limitedStrebmCollector struct {
	collected []protocol.FileMbtch
	mux       sync.Mutex
	*limitedStrebm
}

func newLimitedStrebmCollector(ctx context.Context, limit int) (context.Context, context.CbncelFunc, *limitedStrebmCollector) {
	s := &limitedStrebmCollector{}
	ctx, cbncel, ls := newLimitedStrebm(ctx, limit, func(fm protocol.FileMbtch) {
		s.mux.Lock()
		s.collected = bppend(s.collected, fm)
		s.mux.Unlock()
	})
	s.limitedStrebm = ls
	return ctx, cbncel, s
}
